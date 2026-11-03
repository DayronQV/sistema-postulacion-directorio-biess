package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// fakeBanco imita el banco de preguntas y la configuración en memoria.
type fakeBanco struct {
	preguntas []domain.Pregunta
	sigID     int
	cfg       *domain.ConfigEvaluacion
	intentos  bool
	llamadas  int // veces que se llamó a CrearPreguntas
	filtro    domain.FiltroPreguntas
}

func (f *fakeBanco) CrearPreguntas(_ context.Context, ps []domain.Pregunta) ([]domain.Pregunta, error) {
	f.llamadas++
	creadas := make([]domain.Pregunta, len(ps))
	for i, p := range ps {
		f.sigID++
		p.ID = f.sigID
		creadas[i] = p
		f.preguntas = append(f.preguntas, p)
	}
	return creadas, nil
}

func (f *fakeBanco) ListarPreguntas(_ context.Context, fl domain.FiltroPreguntas) ([]domain.Pregunta, int, error) {
	f.filtro = fl
	return f.preguntas, len(f.preguntas), nil
}

func (f *fakeBanco) ObtenerPregunta(_ context.Context, id int) (domain.Pregunta, bool, error) {
	for _, p := range f.preguntas {
		if p.ID == id {
			return p, true, nil
		}
	}
	return domain.Pregunta{}, false, nil
}

func (f *fakeBanco) ActualizarPregunta(_ context.Context, p domain.Pregunta) (bool, error) {
	for i := range f.preguntas {
		if f.preguntas[i].ID == p.ID {
			f.preguntas[i] = p
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeBanco) EliminarPregunta(_ context.Context, id int) (bool, error) {
	for i := range f.preguntas {
		if f.preguntas[i].ID == id {
			f.preguntas = append(f.preguntas[:i], f.preguntas[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeBanco) ContarPreguntas(context.Context) (int, error) { return len(f.preguntas), nil }
func (f *fakeBanco) HayIntentos(context.Context) (bool, error)    { return f.intentos, nil }

func (f *fakeBanco) ObtenerConfig(context.Context) (domain.ConfigEvaluacion, bool, error) {
	if f.cfg == nil {
		return domain.ConfigEvaluacion{}, false, nil
	}
	return *f.cfg, true, nil
}

func (f *fakeBanco) GuardarConfig(_ context.Context, c domain.ConfigEvaluacion) error {
	f.cfg = &c
	return nil
}

func nuevoBanco(n int) (*service.EvaluacionAdminService, *fakeBanco) {
	f := &fakeBanco{}
	for i := 0; i < n; i++ {
		_, _ = f.CrearPreguntas(context.Background(), []domain.Pregunta{{
			Enunciado: fmt.Sprintf("Pregunta %d", i+1), Opciones: []string{"a", "b", "c"}, RespuestaCorrecta: i % 3,
		}})
	}
	f.llamadas = 0
	return service.NewEvaluacionAdminService(f), f
}

func preguntaValida(n int) service.PreguntaInput {
	return service.PreguntaInput{
		Enunciado: fmt.Sprintf("Pregunta %d", n), Opciones: []string{"Uno", "Dos", "Tres", "Cuatro"}, RespuestaCorrecta: ptr(n % 4),
	}
}

func camposEval(t *testing.T, err error) map[string]string {
	t.Helper()
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("se esperaba un error de dominio, llegó %v", err)
	}
	return de.Fields
}

func TestCrearPregunta_GuardaConTextosLimpios(t *testing.T) {
	svc, f := nuevoBanco(0)
	p, err := svc.CrearPregunta(context.Background(), service.PreguntaInput{
		Enunciado: "  ¿Capital de Ecuador?  ", Opciones: []string{" Guayaquil ", "Quito"}, RespuestaCorrecta: ptr(1),
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.ID == 0 || p.Enunciado != "¿Capital de Ecuador?" || p.Opciones[0] != "Guayaquil" || p.RespuestaCorrecta != 1 {
		t.Fatalf("pregunta = %+v", p)
	}
	if len(f.preguntas) != 1 {
		t.Fatal("no se guardó")
	}
}

func TestCrearPregunta_Invalida(t *testing.T) {
	largo := strings.Repeat("a", domain.MaxLargoEnunciado+1)
	opcionLarga := strings.Repeat("a", domain.MaxLargoOpcion+1)
	casos := []struct {
		nombre string
		in     service.PreguntaInput
		campo  string
	}{
		{"sin enunciado", service.PreguntaInput{Opciones: []string{"a", "b"}, RespuestaCorrecta: ptr(0)}, "enunciado"},
		{"enunciado solo con espacios", service.PreguntaInput{Enunciado: "   ", Opciones: []string{"a", "b"}, RespuestaCorrecta: ptr(0)}, "enunciado"},
		{"enunciado demasiado largo", service.PreguntaInput{Enunciado: largo, Opciones: []string{"a", "b"}, RespuestaCorrecta: ptr(0)}, "enunciado"},
		{"sin opciones", service.PreguntaInput{Enunciado: "x", RespuestaCorrecta: ptr(0)}, "opciones"},
		{"una sola opción", service.PreguntaInput{Enunciado: "x", Opciones: []string{"a"}, RespuestaCorrecta: ptr(0)}, "opciones"},
		{"demasiadas opciones", service.PreguntaInput{Enunciado: "x", Opciones: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}, RespuestaCorrecta: ptr(0)}, "opciones"},
		{"opción vacía", service.PreguntaInput{Enunciado: "x", Opciones: []string{"a", "  "}, RespuestaCorrecta: ptr(0)}, "opciones[1]"},
		{"opción demasiado larga", service.PreguntaInput{Enunciado: "x", Opciones: []string{"a", opcionLarga}, RespuestaCorrecta: ptr(0)}, "opciones[1]"},
		{"opción repetida", service.PreguntaInput{Enunciado: "x", Opciones: []string{"Sí", "no", "SÍ"}, RespuestaCorrecta: ptr(0)}, "opciones[2]"},
		{"sin respuesta correcta", service.PreguntaInput{Enunciado: "x", Opciones: []string{"a", "b"}}, "respuesta_correcta"},
		{"respuesta correcta negativa", service.PreguntaInput{Enunciado: "x", Opciones: []string{"a", "b"}, RespuestaCorrecta: ptr(-1)}, "respuesta_correcta"},
		{"respuesta correcta fuera de rango", service.PreguntaInput{Enunciado: "x", Opciones: []string{"a", "b"}, RespuestaCorrecta: ptr(2)}, "respuesta_correcta"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			svc, f := nuevoBanco(0)
			_, err := svc.CrearPregunta(context.Background(), c.in)
			if !errors.Is(err, domain.ErrValidation) || camposEval(t, err)[c.campo] == "" {
				t.Fatalf("se esperaba 400 con fields[%q], llegó %v (%v)", c.campo, err, camposEval(t, err))
			}
			if len(f.preguntas) != 0 {
				t.Fatal("una pregunta inválida no se guarda")
			}
		})
	}
}

func TestCrearPregunta_ElIndiceCeroEsValido(t *testing.T) {
	svc, _ := nuevoBanco(0)
	if _, err := svc.CrearPregunta(context.Background(), service.PreguntaInput{Enunciado: "x", Opciones: []string{"a", "b"}, RespuestaCorrecta: ptr(0)}); err != nil {
		t.Fatalf("0 es un índice válido: %v", err)
	}
}

func TestCrearPregunta_SinOpcionesValidasNoReportaTambienElIndice(t *testing.T) {
	svc, _ := nuevoBanco(0)
	_, err := svc.CrearPregunta(context.Background(), service.PreguntaInput{Enunciado: "x", Opciones: []string{"a"}, RespuestaCorrecta: ptr(5)})
	campos := camposEval(t, err)
	if campos["opciones"] == "" || campos["respuesta_correcta"] != "" {
		t.Fatalf("solo debe reportar las opciones: %v", campos)
	}
}

// Criterio de la issue #42: carga masiva probada con al menos 20 preguntas.
func TestCrearPreguntas_CargaMasivaDe25EnUnSoloGuardado(t *testing.T) {
	svc, f := nuevoBanco(0)
	entradas := make([]service.PreguntaInput, 25)
	for i := range entradas {
		entradas[i] = preguntaValida(i + 1)
	}
	n, err := svc.CrearPreguntas(context.Background(), entradas)
	if err != nil || n != 25 {
		t.Fatalf("creadas = %d, err = %v", n, err)
	}
	if f.llamadas != 1 || len(f.preguntas) != 25 {
		t.Fatalf("debe guardarse en un solo bloque: llamadas = %d, guardadas = %d", f.llamadas, len(f.preguntas))
	}
}

func TestCrearPreguntas_TodasONinguna_ErroresPorPosicion(t *testing.T) {
	svc, f := nuevoBanco(0)
	entradas := make([]service.PreguntaInput, 20)
	for i := range entradas {
		entradas[i] = preguntaValida(i + 1)
	}
	entradas[3].Enunciado = ""
	entradas[11].Opciones = []string{"solo una"}
	entradas[17].RespuestaCorrecta = ptr(9)

	_, err := svc.CrearPreguntas(context.Background(), entradas)
	campos := camposEval(t, err)
	for _, esperado := range []string{"preguntas[3].enunciado", "preguntas[11].opciones", "preguntas[17].respuesta_correcta"} {
		if campos[esperado] == "" {
			t.Errorf("falta %q en %v", esperado, campos)
		}
	}
	if f.llamadas != 0 || len(f.preguntas) != 0 {
		t.Fatal("con un solo error no se debe guardar ninguna")
	}
}

func TestCrearPreguntas_ListaVaciaOExcesivaEs400(t *testing.T) {
	svc, f := nuevoBanco(0)
	if _, err := svc.CrearPreguntas(context.Background(), nil); !errors.Is(err, domain.ErrValidation) || camposEval(t, err)["preguntas"] == "" {
		t.Fatalf("lista vacía: %v", err)
	}
	demasiadas := make([]service.PreguntaInput, domain.MaxPreguntasPorCarga+1)
	for i := range demasiadas {
		demasiadas[i] = preguntaValida(i + 1)
	}
	if _, err := svc.CrearPreguntas(context.Background(), demasiadas); !errors.Is(err, domain.ErrValidation) || camposEval(t, err)["preguntas"] == "" {
		t.Fatalf("más del máximo: %v", err)
	}
	if f.llamadas != 0 {
		t.Fatal("no se debe guardar")
	}
}

func TestCrearPreguntas_LimitaLosErroresReportados(t *testing.T) {
	svc, _ := nuevoBanco(0)
	entradas := make([]service.PreguntaInput, 100)
	for i := range entradas {
		entradas[i] = service.PreguntaInput{} // todo inválido
	}
	_, err := svc.CrearPreguntas(context.Background(), entradas)
	campos := camposEval(t, err)
	if len(campos) != domain.MaxErroresReportados {
		t.Fatalf("se reportan %d errores, se esperaban %d", len(campos), domain.MaxErroresReportados)
	}
	if !strings.Contains(mensaje(err), "primeros 50") || !strings.Contains(mensaje(err), "300") {
		t.Fatalf("el mensaje debe decir cuántos errores hay en total: %q", mensaje(err))
	}
}

func TestListarPreguntas_NormalizaLaPaginacion(t *testing.T) {
	casos := []struct {
		pagina, porPagina, wantPagina, wantPor int
	}{
		{0, 0, 1, service.PorPaginaPorDefecto},
		{-3, -1, 1, service.PorPaginaPorDefecto},
		{2, 500, 2, service.PorPaginaMaximo},
		{3, 10, 3, 10},
	}
	for _, c := range casos {
		svc, f := nuevoBanco(2)
		r, err := svc.ListarPreguntas(context.Background(), "  capital ", c.pagina, c.porPagina)
		if err != nil || r.Pagina != c.wantPagina || r.PorPagina != c.wantPor || r.Total != 2 {
			t.Fatalf("%+v → %+v, %v", c, r, err)
		}
		if f.filtro.Buscar != "capital" {
			t.Fatalf("la búsqueda debe llegar sin espacios: %q", f.filtro.Buscar)
		}
	}
}

func TestActualizarPregunta(t *testing.T) {
	ctx := context.Background()

	t.Run("reemplaza el contenido", func(t *testing.T) {
		svc, f := nuevoBanco(3)
		p, err := svc.ActualizarPregunta(ctx, "2", service.PreguntaInput{Enunciado: " Nueva ", Opciones: []string{"x", "y"}, RespuestaCorrecta: ptr(1)})
		if err != nil || p.ID != 2 || p.Enunciado != "Nueva" {
			t.Fatalf("pregunta = %+v, err = %v", p, err)
		}
		if f.preguntas[1].Enunciado != "Nueva" || f.preguntas[0].Enunciado != "Pregunta 1" {
			t.Fatalf("solo debe cambiar la pregunta 2: %+v", f.preguntas)
		}
	})
	t.Run("id inexistente o que no es un entero positivo es 404", func(t *testing.T) {
		svc, _ := nuevoBanco(3)
		for _, id := range []string{"99", "0", "-1", "abc", "", "1.5"} {
			if _, err := svc.ActualizarPregunta(ctx, id, preguntaValida(1)); !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("id %q: se esperaba 404, llegó %v", id, err)
			}
		}
	})
	t.Run("datos inválidos es 400", func(t *testing.T) {
		svc, _ := nuevoBanco(3)
		if _, err := svc.ActualizarPregunta(ctx, "1", service.PreguntaInput{}); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("se esperaba 400, llegó %v", err)
		}
	})
	t.Run("con la evaluación iniciada es 409", func(t *testing.T) {
		svc, f := nuevoBanco(3)
		f.intentos = true
		_, err := svc.ActualizarPregunta(ctx, "1", preguntaValida(1))
		if !errors.Is(err, domain.ErrConflict) || mensaje(err) != service.MsgEvaluacionIniciada {
			t.Fatalf("se esperaba 409, llegó %v", err)
		}
		if f.preguntas[0].Enunciado != "Pregunta 1" {
			t.Fatal("no debe cambiar")
		}
	})
}

func TestEliminarPregunta(t *testing.T) {
	ctx := context.Background()

	t.Run("elimina sin configuración", func(t *testing.T) {
		svc, f := nuevoBanco(3)
		if err := svc.EliminarPregunta(ctx, "2"); err != nil || len(f.preguntas) != 2 {
			t.Fatalf("err = %v, quedan %d", err, len(f.preguntas))
		}
	})
	t.Run("id inexistente o inválido es 404", func(t *testing.T) {
		svc, _ := nuevoBanco(3)
		for _, id := range []string{"99", "0", "x"} {
			if err := svc.EliminarPregunta(ctx, id); !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("id %q: se esperaba 404, llegó %v", id, err)
			}
		}
	})
	t.Run("con la evaluación iniciada es 409", func(t *testing.T) {
		svc, f := nuevoBanco(3)
		f.intentos = true
		if err := svc.EliminarPregunta(ctx, "1"); !errors.Is(err, domain.ErrConflict) || mensaje(err) != service.MsgEvaluacionIniciada {
			t.Fatalf("se esperaba 409, llegó %v", err)
		}
		if len(f.preguntas) != 3 {
			t.Fatal("no debe borrarse")
		}
	})
	t.Run("no deja el banco con menos preguntas que la configuración", func(t *testing.T) {
		svc, f := nuevoBanco(3)
		f.cfg = &domain.ConfigEvaluacion{DuracionMinutos: 30, NumPreguntas: 3}
		err := svc.EliminarPregunta(ctx, "1")
		if !errors.Is(err, domain.ErrConflict) || mensaje(err) != service.MsgBorrarDejaCorto {
			t.Fatalf("se esperaba 409, llegó %v", err)
		}
		if len(f.preguntas) != 3 {
			t.Fatal("no debe borrarse")
		}
		// Con una pregunta de sobra sí se puede.
		f.cfg.NumPreguntas = 2
		if err := svc.EliminarPregunta(ctx, "1"); err != nil || len(f.preguntas) != 2 {
			t.Fatalf("con margen debe poder borrarse: %v", err)
		}
	})
	t.Run("un id inexistente no se confunde con el límite del banco", func(t *testing.T) {
		svc, f := nuevoBanco(3)
		f.cfg = &domain.ConfigEvaluacion{DuracionMinutos: 30, NumPreguntas: 3}
		if err := svc.EliminarPregunta(ctx, "99"); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("se esperaba 404, llegó %v", err)
		}
	})
}

func int2ptr(n int) *int { return &n }

func configValida() service.ConfigInput {
	return service.ConfigInput{DuracionMinutos: int2ptr(45), NumPreguntas: int2ptr(5)}
}

func TestObtenerConfig_SinConfigurarMuestraLosValoresPorDefecto(t *testing.T) {
	svc, _ := nuevoBanco(4)
	c, err := svc.ObtenerConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if c.Configurada || c.DuracionMinutos != service.DuracionMinutosPorDefecto ||
		c.NumPreguntas != service.NumPreguntasPorDefecto || c.TotalBanco != 4 {
		t.Fatalf("config = %+v", c)
	}
}

func TestGuardarConfig_Valida(t *testing.T) {
	svc, f := nuevoBanco(5)
	c, err := svc.GuardarConfig(context.Background(), configValida())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !c.Configurada || c.DuracionMinutos != 45 || c.NumPreguntas != 5 || c.TotalBanco != 5 {
		t.Fatalf("config = %+v", c)
	}
	if f.cfg == nil || f.cfg.NumPreguntas != 5 || f.cfg.DuracionMinutos != 45 {
		t.Fatalf("no se guardó: %+v", f.cfg)
	}
	// Lo guardado se lee de vuelta.
	leida, _ := svc.ObtenerConfig(context.Background())
	if !leida.Configurada || leida.NumPreguntas != 5 {
		t.Fatalf("lectura = %+v", leida)
	}
}

// Criterio de la issue #42: la configuración no acepta más preguntas de las que hay en el banco.
func TestGuardarConfig_NoAceptaMasPreguntasQueElBanco(t *testing.T) {
	svc, f := nuevoBanco(5)
	in := configValida()
	in.NumPreguntas = int2ptr(6)
	_, err := svc.GuardarConfig(context.Background(), in)
	if !errors.Is(err, domain.ErrUnprocessable) || mensaje(err) != service.MsgBancoInsuficiente || camposEval(t, err)["num_preguntas"] == "" {
		t.Fatalf("se esperaba 422 con fields.num_preguntas, llegó %v (%v)", err, camposEval(t, err))
	}
	if f.cfg != nil {
		t.Fatal("no se debe guardar")
	}
	// Exactamente el tamaño del banco sí se acepta.
	in.NumPreguntas = int2ptr(5)
	if _, err := svc.GuardarConfig(context.Background(), in); err != nil {
		t.Fatalf("igual al tamaño del banco debe aceptarse: %v", err)
	}
}

func TestGuardarConfig_BancoVacioNoAceptaNada(t *testing.T) {
	svc, _ := nuevoBanco(0)
	if _, err := svc.GuardarConfig(context.Background(), configValida()); !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("se esperaba 422, llegó %v", err)
	}
}

func TestGuardarConfig_Invalida(t *testing.T) {
	casos := []struct {
		nombre string
		cambia func(*service.ConfigInput)
		campo  string
	}{
		{"sin duración", func(c *service.ConfigInput) { c.DuracionMinutos = nil }, "duracion_minutos"},
		{"duración cero", func(c *service.ConfigInput) { c.DuracionMinutos = int2ptr(0) }, "duracion_minutos"},
		{"duración excesiva", func(c *service.ConfigInput) { c.DuracionMinutos = int2ptr(domain.MaxDuracionMinutos + 1) }, "duracion_minutos"},
		{"sin número de preguntas", func(c *service.ConfigInput) { c.NumPreguntas = nil }, "num_preguntas"},
		{"cero preguntas", func(c *service.ConfigInput) { c.NumPreguntas = int2ptr(0) }, "num_preguntas"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			svc, f := nuevoBanco(5)
			in := configValida()
			c.cambia(&in)
			_, err := svc.GuardarConfig(context.Background(), in)
			if !errors.Is(err, domain.ErrValidation) || camposEval(t, err)[c.campo] == "" {
				t.Fatalf("se esperaba 400 con fields[%q], llegó %v (%v)", c.campo, err, camposEval(t, err))
			}
			if f.cfg != nil {
				t.Fatal("no se debe guardar")
			}
		})
	}
}

func TestGuardarConfig_LimitesDeLaDuracionSonValidos(t *testing.T) {
	for _, d := range []int{domain.MinDuracionMinutos, domain.MaxDuracionMinutos} {
		svc, _ := nuevoBanco(5)
		in := configValida()
		in.DuracionMinutos = int2ptr(d)
		if _, err := svc.GuardarConfig(context.Background(), in); err != nil {
			t.Fatalf("%d minutos debe aceptarse: %v", d, err)
		}
	}
}
