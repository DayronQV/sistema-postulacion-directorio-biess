package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Mensajes de las reglas de negocio del banco y de la configuración de la evaluación.
const (
	MsgEvaluacionIniciada = "La evaluación ya comenzó: no se pueden modificar ni eliminar preguntas"
	MsgBancoInsuficiente  = "La configuración pide más preguntas de las que hay en el banco"
	MsgPreguntaNoExiste   = "La pregunta no existe"
	MsgBorrarDejaCorto    = "No se puede eliminar: el banco quedaría con menos preguntas de las que pide la configuración"
)

// Valores que se muestran mientras la evaluación no se haya configurado (son los de la
// tabla eval_config).
const (
	DuracionMinutosPorDefecto = 30
	NumPreguntasPorDefecto    = 10
)

// Paginación del listado del banco.
const (
	PorPaginaPorDefecto = 20
	PorPaginaMaximo     = 100
)

// PreguntaInput es una pregunta como llega del cliente. RespuestaCorrecta es un puntero
// para distinguir "no vino" de 0, que es un índice válido.
type PreguntaInput struct {
	Enunciado         string
	Opciones          []string
	RespuestaCorrecta *int
}

// ConfigInput son los parámetros del examen como llegan del cliente. El horario no va aquí:
// lo fija la fase Evaluación.
type ConfigInput struct {
	DuracionMinutos *int
	NumPreguntas    *int
}

// ConfigEvaluacion es la configuración vigente más el tamaño actual del banco.
type ConfigEvaluacion struct {
	DuracionMinutos int
	NumPreguntas    int
	TotalBanco      int
	Configurada     bool
}

// PaginaPreguntas es una página del listado del banco.
type PaginaPreguntas struct {
	Items     []domain.Pregunta
	Total     int
	Pagina    int
	PorPagina int
}

// EvaluacionAdminService reúne las acciones del administrador sobre el banco de preguntas
// y la configuración de la evaluación teórica (EF-14). Que solo el administrador pueda
// ejecutarlas lo controla la ruta (grupo de escritura del panel).
type EvaluacionAdminService struct {
	banco domain.BancoEvaluacion
}

// NewEvaluacionAdminService crea el servicio con sus dependencias inyectadas.
func NewEvaluacionAdminService(banco domain.BancoEvaluacion) *EvaluacionAdminService {
	return &EvaluacionAdminService{banco: banco}
}

// CrearPregunta valida y guarda una pregunta.
func (s *EvaluacionAdminService) CrearPregunta(ctx context.Context, in PreguntaInput) (domain.Pregunta, error) {
	rep := &reporte{}
	p, ok := validarPregunta(in, "", rep)
	if !ok {
		return domain.Pregunta{}, rep.error()
	}
	creadas, err := s.banco.CrearPreguntas(ctx, []domain.Pregunta{p})
	if err != nil {
		return domain.Pregunta{}, err
	}
	return creadas[0], nil
}

// CrearPreguntas valida todas las preguntas y las guarda en bloque (todas o ninguna). Los
// errores se reportan por posición: preguntas[3].opciones[1].
func (s *EvaluacionAdminService) CrearPreguntas(ctx context.Context, entradas []PreguntaInput) (int, error) {
	rep := &reporte{}
	if len(entradas) == 0 {
		rep.agregar("preguntas", "Debe enviar al menos una pregunta")
		return 0, rep.error()
	}
	if len(entradas) > domain.MaxPreguntasPorCarga {
		rep.agregar("preguntas", fmt.Sprintf("Máximo %d preguntas por carga", domain.MaxPreguntasPorCarga))
		return 0, rep.error()
	}
	preguntas := make([]domain.Pregunta, 0, len(entradas))
	for i, in := range entradas {
		if p, ok := validarPregunta(in, fmt.Sprintf("preguntas[%d].", i), rep); ok {
			preguntas = append(preguntas, p)
		}
	}
	if rep.hay() {
		return 0, rep.error()
	}
	creadas, err := s.banco.CrearPreguntas(ctx, preguntas)
	if err != nil {
		return 0, err
	}
	return len(creadas), nil
}

// CrearPreguntasCSV lee un CSV con las columnas enunciado, opciones (separadas por |) y
// respuesta_correcta, y las guarda en bloque. Los errores se reportan por número de línea
// del archivo: filas[4].opciones.
func (s *EvaluacionAdminService) CrearPreguntasCSV(ctx context.Context, r io.Reader) (int, error) {
	rep := &reporte{}
	filas := leerPreguntasCSV(r, rep)
	preguntas := make([]domain.Pregunta, 0, len(filas))
	for _, f := range filas {
		if p, ok := validarPregunta(f.input, fmt.Sprintf("filas[%d].", f.linea), rep); ok {
			preguntas = append(preguntas, p)
		}
	}
	if rep.hay() {
		return 0, rep.error()
	}
	creadas, err := s.banco.CrearPreguntas(ctx, preguntas)
	if err != nil {
		return 0, err
	}
	return len(creadas), nil
}

// ListarPreguntas devuelve una página del banco, con la respuesta correcta (vista del
// administrador).
func (s *EvaluacionAdminService) ListarPreguntas(ctx context.Context, buscar string, pagina, porPagina int) (PaginaPreguntas, error) {
	if pagina < 1 {
		pagina = 1
	}
	if porPagina < 1 {
		porPagina = PorPaginaPorDefecto
	}
	if porPagina > PorPaginaMaximo {
		porPagina = PorPaginaMaximo
	}
	items, total, err := s.banco.ListarPreguntas(ctx, domain.FiltroPreguntas{
		Buscar: strings.TrimSpace(buscar), Pagina: pagina, PorPagina: porPagina,
	})
	if err != nil {
		return PaginaPreguntas{}, err
	}
	return PaginaPreguntas{Items: items, Total: total, Pagina: pagina, PorPagina: porPagina}, nil
}

// ActualizarPregunta reemplaza una pregunta del banco. Una vez que algún postulante inició
// la evaluación ya no se puede: cambiar opciones o la respuesta correcta alteraría los
// exámenes en curso y su calificación.
func (s *EvaluacionAdminService) ActualizarPregunta(ctx context.Context, idTexto string, in PreguntaInput) (domain.Pregunta, error) {
	id, ok := parsearIDPregunta(idTexto)
	if !ok {
		return domain.Pregunta{}, domain.NotFound(MsgPreguntaNoExiste)
	}
	rep := &reporte{}
	p, valida := validarPregunta(in, "", rep)
	if !valida {
		return domain.Pregunta{}, rep.error()
	}
	if err := s.exigirEvaluacionSinIniciar(ctx); err != nil {
		return domain.Pregunta{}, err
	}
	p.ID = id
	actualizada, err := s.banco.ActualizarPregunta(ctx, p)
	if err != nil {
		return domain.Pregunta{}, err
	}
	if !actualizada {
		return domain.Pregunta{}, domain.NotFound(MsgPreguntaNoExiste)
	}
	return p, nil
}

// EliminarPregunta borra una pregunta del banco. No se puede si la evaluación ya empezó, ni
// si el banco quedaría con menos preguntas de las que pide la configuración.
func (s *EvaluacionAdminService) EliminarPregunta(ctx context.Context, idTexto string) error {
	id, ok := parsearIDPregunta(idTexto)
	if !ok {
		return domain.NotFound(MsgPreguntaNoExiste)
	}
	if _, existe, err := s.banco.ObtenerPregunta(ctx, id); err != nil {
		return err
	} else if !existe {
		return domain.NotFound(MsgPreguntaNoExiste)
	}
	if err := s.exigirEvaluacionSinIniciar(ctx); err != nil {
		return err
	}
	cfg, configurada, err := s.banco.ObtenerConfig(ctx)
	if err != nil {
		return err
	}
	if configurada {
		total, err := s.banco.ContarPreguntas(ctx)
		if err != nil {
			return err
		}
		if total-1 < cfg.NumPreguntas {
			return domain.Conflict(MsgBorrarDejaCorto, map[string]string{
				"num_preguntas": fmt.Sprintf("La configuración pide %d y el banco tiene %d", cfg.NumPreguntas, total),
			})
		}
	}
	eliminada, err := s.banco.EliminarPregunta(ctx, id)
	if err != nil {
		return err
	}
	if !eliminada {
		return domain.NotFound(MsgPreguntaNoExiste)
	}
	return nil
}

// ObtenerConfig devuelve la configuración vigente (o los valores por defecto si nunca se
// configuró) y el tamaño del banco.
func (s *EvaluacionAdminService) ObtenerConfig(ctx context.Context) (ConfigEvaluacion, error) {
	cfg, configurada, err := s.banco.ObtenerConfig(ctx)
	if err != nil {
		return ConfigEvaluacion{}, err
	}
	total, err := s.banco.ContarPreguntas(ctx)
	if err != nil {
		return ConfigEvaluacion{}, err
	}
	if !configurada {
		cfg = domain.ConfigEvaluacion{DuracionMinutos: DuracionMinutosPorDefecto, NumPreguntas: NumPreguntasPorDefecto}
	}
	return ConfigEvaluacion{
		DuracionMinutos: cfg.DuracionMinutos, NumPreguntas: cfg.NumPreguntas,
		TotalBanco: total, Configurada: configurada,
	}, nil
}

// GuardarConfig valida y guarda la duración y el número de preguntas del examen. El número de
// preguntas no puede superar las que hay en el banco (422). Cuándo se puede rendir lo fija la
// fase Evaluación, no esta configuración.
func (s *EvaluacionAdminService) GuardarConfig(ctx context.Context, in ConfigInput) (ConfigEvaluacion, error) {
	rep := &reporte{}
	switch d := in.DuracionMinutos; {
	case d == nil:
		rep.agregar("duracion_minutos", "Campo obligatorio")
	case *d < domain.MinDuracionMinutos || *d > domain.MaxDuracionMinutos:
		rep.agregar("duracion_minutos", fmt.Sprintf("Debe estar entre %d y %d minutos", domain.MinDuracionMinutos, domain.MaxDuracionMinutos))
	}
	switch n := in.NumPreguntas; {
	case n == nil:
		rep.agregar("num_preguntas", "Campo obligatorio")
	case *n < 1:
		rep.agregar("num_preguntas", "Debe ser al menos 1")
	}
	if rep.hay() {
		return ConfigEvaluacion{}, rep.error()
	}

	total, err := s.banco.ContarPreguntas(ctx)
	if err != nil {
		return ConfigEvaluacion{}, err
	}
	if *in.NumPreguntas > total {
		return ConfigEvaluacion{}, domain.Unprocessable(MsgBancoInsuficiente, map[string]string{
			"num_preguntas": fmt.Sprintf("El banco tiene %d pregunta(s) y se piden %d", total, *in.NumPreguntas),
		})
	}
	cfg := domain.ConfigEvaluacion{DuracionMinutos: *in.DuracionMinutos, NumPreguntas: *in.NumPreguntas}
	if err := s.banco.GuardarConfig(ctx, cfg); err != nil {
		return ConfigEvaluacion{}, err
	}
	return ConfigEvaluacion{
		DuracionMinutos: cfg.DuracionMinutos, NumPreguntas: cfg.NumPreguntas,
		TotalBanco: total, Configurada: true,
	}, nil
}

func (s *EvaluacionAdminService) exigirEvaluacionSinIniciar(ctx context.Context) error {
	hay, err := s.banco.HayIntentos(ctx)
	if err != nil {
		return err
	}
	if hay {
		return domain.Conflict(MsgEvaluacionIniciada, nil)
	}
	return nil
}

// parsearIDPregunta convierte el id de la ruta; uno que no es un entero positivo no puede
// existir (404).
func parsearIDPregunta(texto string) (int, bool) {
	id, err := strconv.Atoi(strings.TrimSpace(texto))
	return id, err == nil && id > 0
}

// reporte junta los errores por campo de una validación. Guarda solo los primeros
// MaxErroresReportados para no devolver una respuesta enorme con un archivo de miles de
// errores, pero cuenta todos.
type reporte struct {
	campos map[string]string
	total  int
}

func (r *reporte) agregar(campo, mensaje string) {
	r.total++
	if len(r.campos) >= domain.MaxErroresReportados {
		return
	}
	if r.campos == nil {
		r.campos = map[string]string{}
	}
	r.campos[campo] = mensaje
}

func (r *reporte) hay() bool { return r.total > 0 }

func (r *reporte) error() error {
	mensaje := "Datos inválidos"
	if r.total > len(r.campos) {
		mensaje = fmt.Sprintf("Datos inválidos: se muestran los primeros %d errores de %d", len(r.campos), r.total)
	}
	return domain.Validation(mensaje, r.campos)
}

// validarPregunta revisa una pregunta y agrega al reporte cada problema, con el prefijo de
// su posición. Devuelve la pregunta normalizada (textos sin espacios sobrantes) y si es
// válida.
func validarPregunta(in PreguntaInput, prefijo string, rep *reporte) (domain.Pregunta, bool) {
	antes := rep.total
	enunciado := strings.TrimSpace(in.Enunciado)
	switch {
	case enunciado == "":
		rep.agregar(prefijo+"enunciado", "Campo obligatorio")
	case len([]rune(enunciado)) > domain.MaxLargoEnunciado:
		rep.agregar(prefijo+"enunciado", fmt.Sprintf("Máximo %d caracteres", domain.MaxLargoEnunciado))
	}

	opciones := make([]string, len(in.Opciones))
	cantidadValida := len(in.Opciones) >= domain.MinOpcionesPregunta && len(in.Opciones) <= domain.MaxOpcionesPregunta
	if !cantidadValida {
		rep.agregar(prefijo+"opciones", fmt.Sprintf("Debe tener entre %d y %d opciones", domain.MinOpcionesPregunta, domain.MaxOpcionesPregunta))
	} else {
		vistas := map[string]bool{}
		for i, o := range in.Opciones {
			o = strings.TrimSpace(o)
			opciones[i] = o
			campo := fmt.Sprintf("%sopciones[%d]", prefijo, i)
			switch {
			case o == "":
				rep.agregar(campo, "La opción no puede estar vacía")
			case len([]rune(o)) > domain.MaxLargoOpcion:
				rep.agregar(campo, fmt.Sprintf("Máximo %d caracteres", domain.MaxLargoOpcion))
			case vistas[strings.ToLower(o)]:
				rep.agregar(campo, "La opción está repetida")
			}
			vistas[strings.ToLower(o)] = true
		}
	}

	correcta := 0
	switch c := in.RespuestaCorrecta; {
	case c == nil:
		rep.agregar(prefijo+"respuesta_correcta", "Campo obligatorio")
	case !cantidadValida:
		// Sin una lista de opciones válida no hay contra qué comparar el índice.
	case *c < 0 || *c >= len(opciones):
		rep.agregar(prefijo+"respuesta_correcta",
			fmt.Sprintf("Debe ser el índice de una opción, de 0 a %d", len(opciones)-1))
	default:
		correcta = *c
	}

	return domain.Pregunta{Enunciado: enunciado, Opciones: opciones, RespuestaCorrecta: correcta}, rep.total == antes
}
