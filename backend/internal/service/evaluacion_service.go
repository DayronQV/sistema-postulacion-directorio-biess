package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Mensajes de la evaluación teórica para el postulante.
const (
	MsgEvaluacionSinPostulacion = "Debe tener una postulación enviada para rendir la evaluación"
	MsgEvaluacionSinMerito      = "Aún no está habilitado para la evaluación: debe superar la Fase de Méritos"
	MsgEvaluacionNoPublicada    = "La evaluación todavía no está habilitada"
	MsgEvaluacionNoComienza     = "La evaluación aún no ha comenzado"
	MsgEvaluacionPlazoVencido   = "El plazo para iniciar la evaluación terminó"
	MsgEvaluacionBancoCorto     = "La evaluación no se puede iniciar: el banco no tiene las preguntas que pide la configuración"
	MsgEvaluacionTerminada      = "La evaluación ya terminó"
	MsgEvaluacionNoIniciada     = "Todavía no inició la evaluación"
	MsgPreguntaFueraDeRango     = "La pregunta no existe en su evaluación"
)

// BarajadorCrypto es el Barajador de producción: usa crypto/rand, así que los órdenes no se
// pueden predecir (Fisher–Yates con enteros uniformes, sin sesgo).
type BarajadorCrypto struct{}

var _ domain.Barajador = BarajadorCrypto{}

// Permutacion devuelve 0..n-1 en un orden aleatorio.
func (BarajadorCrypto) Permutacion(n int) ([]int, error) {
	orden := make([]int, n)
	for i := range orden {
		orden[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, fmt.Errorf("generando el orden aleatorio: %w", err)
		}
		orden[i], orden[j.Int64()] = orden[j.Int64()], orden[i]
	}
	return orden, nil
}

// SesionEvaluacion es el estado de la evaluación de un postulante.
type SesionEvaluacion struct {
	TotalPreguntas int
	Inicio         time.Time
	Cierre         time.Time
	// TiempoRestante lo calcula el servidor: el reloj del navegador no cuenta.
	TiempoRestante time.Duration
}

// PreguntaEvaluacion es una pregunta del examen tal como la ve el postulante: sin la
// respuesta correcta y con las opciones en el orden que le tocó.
type PreguntaEvaluacion struct {
	Numero         int
	Total          int
	Pregunta       domain.PreguntaPublica
	Cierre         time.Time
	TiempoRestante time.Duration
}

// EvaluacionService atiende al postulante en la evaluación teórica (EF-14): iniciar el
// examen con su sorteo y mostrarle las preguntas una por una.
type EvaluacionService struct {
	banco     domain.BancoEvaluacion
	intentos  domain.IntentosEvaluacion
	fases     domain.Fases
	barajador domain.Barajador
	ahora     func() time.Time
}

// NewEvaluacionService crea el servicio con sus dependencias inyectadas. fases es la tabla de
// fases del concurso (la fase Evaluación fija cuándo se puede rendir el examen) y ahora es el
// reloj (time.Now en producción).
func NewEvaluacionService(banco domain.BancoEvaluacion, intentos domain.IntentosEvaluacion, fases domain.Fases, barajador domain.Barajador, ahora func() time.Time) *EvaluacionService {
	return &EvaluacionService{banco: banco, intentos: intentos, fases: fases, barajador: barajador, ahora: ahora}
}

// Iniciar deja al postulante comenzar la evaluación. Devuelve su sesión y true si la creó
// esta llamada; si ya tenía una vigente, devuelve la misma (false).
//
// Reglas: solo con la postulación enviada y la Fase de Méritos superada, y mientras la fase
// Evaluación esté publicada y la hora esté entre su fecha de inicio y su fecha de fin (sin
// fecha en un lado, no hay límite por ese lado); si no, 403. La hora de cierre del examen de
// cada postulante es la menor entre su inicio más la duración y la fecha de fin de la fase:
// quien empieza cerca del cierre solo tiene el tiempo que queda. Hay un solo intento por
// postulante: uno que ya terminó o venció no se reabre (409). El sorteo toma las preguntas que
// pide la configuración, en orden aleatorio, y cada pregunta con sus opciones en orden aleatorio.
func (s *EvaluacionService) Iniciar(ctx context.Context, postulanteID string) (SesionEvaluacion, bool, error) {
	post, ok, err := s.intentos.PostulacionDelPostulante(ctx, postulanteID)
	if err != nil {
		return SesionEvaluacion{}, false, err
	}
	if !ok || post.Estado != domain.EstadoEnviada {
		return SesionEvaluacion{}, false, domain.Forbidden(MsgEvaluacionSinPostulacion)
	}
	if !post.SuperoMerito {
		return SesionEvaluacion{}, false, domain.Forbidden(MsgEvaluacionSinMerito)
	}

	// PostgreSQL guarda microsegundos: se trunca aquí para que la hora de cierre que se
	// devuelve al crear el examen sea idéntica a la que se lee después.
	ahora := s.ahora().Truncate(time.Microsecond)
	// Un intento que ya empezó se devuelve aunque el horario haya cambiado después.
	if existente, hay, err := s.intentos.ObtenerIntento(ctx, post.ID); err != nil {
		return SesionEvaluacion{}, false, err
	} else if hay {
		sesion, err := sesionVigente(existente, ahora)
		return sesion, false, err
	}

	fase, err := s.fases.PorNombre(ctx, domain.FaseEvaluacion)
	if errors.Is(err, domain.ErrNotFound) {
		return SesionEvaluacion{}, false, domain.Forbidden(MsgEvaluacionNoPublicada)
	}
	if err != nil {
		return SesionEvaluacion{}, false, err
	}
	if !fase.Publicada {
		return SesionEvaluacion{}, false, domain.Forbidden(MsgEvaluacionNoPublicada)
	}
	if fase.FechaInicio != nil && ahora.Before(*fase.FechaInicio) {
		return SesionEvaluacion{}, false, &domain.Error{
			Base: domain.ErrForbidden, Message: MsgEvaluacionNoComienza,
			Fields: map[string]string{"inicia_en": fase.FechaInicio.UTC().Format(time.RFC3339)},
		}
	}
	// En la fecha de fin exacta ya no queda tiempo: el examen cierra en ese instante.
	if fase.FechaFin != nil && !ahora.Before(*fase.FechaFin) {
		return SesionEvaluacion{}, false, &domain.Error{
			Base: domain.ErrForbidden, Message: MsgEvaluacionPlazoVencido,
			Fields: map[string]string{"cierra_en": fase.FechaFin.UTC().Format(time.RFC3339)},
		}
	}

	// Duración y número de preguntas: lo configurado, o los valores por defecto que muestra el
	// panel mientras no se configure.
	cfg, configurada, err := s.banco.ObtenerConfig(ctx)
	if err != nil {
		return SesionEvaluacion{}, false, err
	}
	if !configurada {
		cfg = domain.ConfigEvaluacion{DuracionMinutos: DuracionMinutosPorDefecto, NumPreguntas: NumPreguntasPorDefecto}
	}

	preguntas, err := s.sortear(ctx, cfg.NumPreguntas)
	if err != nil {
		return SesionEvaluacion{}, false, err
	}
	// El examen dura la duración configurada, pero nunca pasa de la fecha de fin de la fase.
	cierre := ahora.Add(time.Duration(cfg.DuracionMinutos) * time.Minute)
	if fase.FechaFin != nil && fase.FechaFin.Before(cierre) {
		cierre = fase.FechaFin.Truncate(time.Microsecond)
	}
	intento, creado, err := s.intentos.CrearIntento(ctx, domain.IntentoEvaluacion{
		PostulacionID: post.ID,
		Preguntas:     preguntas,
		Inicio:        ahora,
		Cierre:        cierre,
	})
	if err != nil {
		return SesionEvaluacion{}, false, err
	}
	sesion, err := sesionVigente(intento, ahora)
	return sesion, creado, err
}

// Pregunta devuelve la pregunta n (desde 1) del examen del postulante, sin la respuesta
// correcta y con las opciones en su orden, más el tiempo que le queda según el servidor.
func (s *EvaluacionService) Pregunta(ctx context.Context, postulanteID string, n int) (PreguntaEvaluacion, error) {
	post, ok, err := s.intentos.PostulacionDelPostulante(ctx, postulanteID)
	if err != nil {
		return PreguntaEvaluacion{}, err
	}
	if !ok {
		return PreguntaEvaluacion{}, domain.NotFound(MsgEvaluacionNoIniciada)
	}
	intento, hay, err := s.intentos.ObtenerIntento(ctx, post.ID)
	if err != nil {
		return PreguntaEvaluacion{}, err
	}
	if !hay {
		return PreguntaEvaluacion{}, domain.NotFound(MsgEvaluacionNoIniciada)
	}
	ahora := s.ahora()
	if !intento.Vigente(ahora) {
		return PreguntaEvaluacion{}, domain.Conflict(MsgEvaluacionTerminada, nil)
	}
	if n < 1 || n > len(intento.Preguntas) {
		return PreguntaEvaluacion{}, domain.NotFound(MsgPreguntaFueraDeRango)
	}

	sorteada := intento.Preguntas[n-1]
	p, existe, err := s.banco.ObtenerPregunta(ctx, sorteada.PreguntaID)
	if err != nil {
		return PreguntaEvaluacion{}, err
	}
	if !existe {
		return PreguntaEvaluacion{}, fmt.Errorf("la pregunta %d del intento %s ya no está en el banco", sorteada.PreguntaID, intento.ID)
	}
	opciones := make([]string, len(sorteada.Orden))
	for k, original := range sorteada.Orden {
		if original < 0 || original >= len(p.Opciones) {
			return PreguntaEvaluacion{}, fmt.Errorf("el orden de opciones del intento %s no coincide con la pregunta %d", intento.ID, p.ID)
		}
		opciones[k] = p.Opciones[original]
	}
	return PreguntaEvaluacion{
		Numero:         n,
		Total:          len(intento.Preguntas),
		Pregunta:       domain.PreguntaPublica{ID: p.ID, Enunciado: p.Enunciado, Opciones: opciones},
		Cierre:         intento.Cierre,
		TiempoRestante: intento.Cierre.Sub(ahora),
	}, nil
}

// sortear elige n preguntas distintas del banco en orden aleatorio y, para cada una, un
// orden aleatorio de sus opciones.
func (s *EvaluacionService) sortear(ctx context.Context, n int) ([]domain.PreguntaDelIntento, error) {
	banco, err := s.intentos.PreguntasParaSorteo(ctx)
	if err != nil {
		return nil, err
	}
	if n < 1 || n > len(banco) {
		return nil, domain.Unprocessable(MsgEvaluacionBancoCorto, map[string]string{
			"num_preguntas": fmt.Sprintf("La configuración pide %d y el banco tiene %d", n, len(banco)),
		})
	}
	orden, err := s.barajador.Permutacion(len(banco))
	if err != nil {
		return nil, err
	}
	if len(orden) != len(banco) {
		return nil, fmt.Errorf("el barajador devolvió %d posiciones para %d preguntas", len(orden), len(banco))
	}
	elegidas := make([]domain.PreguntaDelIntento, 0, n)
	for _, pos := range orden[:n] {
		p := banco[pos]
		opciones, err := s.barajador.Permutacion(p.NumOpciones)
		if err != nil {
			return nil, err
		}
		elegidas = append(elegidas, domain.PreguntaDelIntento{PreguntaID: p.ID, Orden: opciones})
	}
	return elegidas, nil
}

// sesionVigente arma la sesión de un intento abierto; uno terminado o vencido es 409.
func sesionVigente(i domain.IntentoEvaluacion, ahora time.Time) (SesionEvaluacion, error) {
	if !i.Vigente(ahora) {
		return SesionEvaluacion{}, domain.Conflict(MsgEvaluacionTerminada, nil)
	}
	return SesionEvaluacion{
		TotalPreguntas: len(i.Preguntas), Inicio: i.Inicio, Cierre: i.Cierre,
		TiempoRestante: i.Cierre.Sub(ahora),
	}, nil
}
