package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// fakeIntentos imita al repositorio de Redis en memoria.
type fakeIntentos struct {
	fallos   map[string]int
	bloqueos map[string]time.Duration
	err      error
}

func newFakeIntentos() *fakeIntentos {
	return &fakeIntentos{fallos: map[string]int{}, bloqueos: map[string]time.Duration{}}
}

func (f *fakeIntentos) Bloqueo(_ context.Context, cedula string) (time.Duration, error) {
	return f.bloqueos[cedula], f.err
}

func (f *fakeIntentos) RegistrarFallo(_ context.Context, cedula string, _ time.Duration) (int, error) {
	f.fallos[cedula]++
	return f.fallos[cedula], f.err
}

func (f *fakeIntentos) Bloquear(_ context.Context, cedula string, d time.Duration) error {
	f.bloqueos[cedula] = d
	delete(f.fallos, cedula)
	return f.err
}

func (f *fakeIntentos) Reiniciar(_ context.Context, cedula string) error {
	delete(f.fallos, cedula)
	return f.err
}

type fakeTickets struct {
	emitidos map[string]string
}

func (f *fakeTickets) Emitir(_ context.Context, cedula string, _ time.Duration) (string, error) {
	id := "ticket-" + cedula
	f.emitidos[id] = cedula
	return id, nil
}

type fakePostulantes struct {
	registradas map[string]bool
}

func (f *fakePostulantes) ExisteCedula(_ context.Context, cedula string) (bool, error) {
	return f.registradas[cedula], nil
}

type fixture struct {
	svc         *service.RegistroService
	intentos    *fakeIntentos
	tickets     *fakeTickets
	postulantes *fakePostulantes
}

func newFixture() fixture {
	f := fixture{
		intentos:    newFakeIntentos(),
		tickets:     &fakeTickets{emitidos: map[string]string{}},
		postulantes: &fakePostulantes{registradas: map[string]bool{}},
	}
	f.svc = service.NewRegistroService(registrocivil.NewMock(), f.intentos, f.tickets, f.postulantes)
	return f
}

const cedula = registrocivil.CedulaBloqueoPorIntento // fecha 05/12/2018, código E5566V7788

func validar(f fixture, in service.IdentidadInput) (string, error) {
	return f.svc.ValidarIdentidad(context.Background(), in)
}

func mensaje(err error) string {
	var de *domain.Error
	if errors.As(err, &de) {
		return de.Message
	}
	return ""
}

func TestValidarIdentidad_Exito(t *testing.T) {
	casos := map[string]service.IdentidadInput{
		"fecha ISO":         {Cedula: cedula, FechaExpedicion: "2018-12-05"},
		"fecha DD/MM/AAAA":  {Cedula: cedula, FechaExpedicion: "05/12/2018"},
		"código dactilar":   {Cedula: cedula, CodigoDactilar: "E5566V7788"},
		"código minúsculas": {Cedula: cedula, CodigoDactilar: "e5566v7788"},
	}
	for nombre, in := range casos {
		t.Run(nombre, func(t *testing.T) {
			f := newFixture()
			f.intentos.fallos[cedula] = 2 // un acierto antes del tercero reinicia el contador
			ticket, err := validar(f, in)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if f.tickets.emitidos[ticket] != cedula {
				t.Fatalf("ticket %q no quedó asociado a la cédula", ticket)
			}
			if f.intentos.fallos[cedula] != 0 {
				t.Fatalf("el contador no se reinició: %d", f.intentos.fallos[cedula])
			}
		})
	}
}

func TestValidarIdentidad_DatoIncorrecto(t *testing.T) {
	f := newFixture()
	_, err := validar(f, service.IdentidadInput{Cedula: cedula, FechaExpedicion: "06/12/2018"})
	if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgIdentidadIncorrecta {
		t.Fatalf("se esperaba 401 %q, llegó %v", service.MsgIdentidadIncorrecta, err)
	}
	if f.intentos.fallos[cedula] != 1 {
		t.Fatalf("fallos = %d, se esperaba 1", f.intentos.fallos[cedula])
	}
}

func TestValidarIdentidad_CedulaInexistenteCuentaComoFallo(t *testing.T) {
	f := newFixture()
	const inexistente = "1710034065" // válida por módulo 10, no está en el simulador
	_, err := validar(f, service.IdentidadInput{Cedula: inexistente, CodigoDactilar: "V1234V5678"})
	if !errors.Is(err, domain.ErrUnauthorized) || mensaje(err) != service.MsgIdentidadIncorrecta {
		t.Fatalf("se esperaba el mismo 401 que un dato incorrecto, llegó %v", err)
	}
	if f.intentos.fallos[inexistente] != 1 {
		t.Fatalf("la cédula inexistente no contó como fallo")
	}
}

func TestValidarIdentidad_TercerFalloBloquea(t *testing.T) {
	f := newFixture()
	malo := service.IdentidadInput{Cedula: cedula, CodigoDactilar: "AAAAAAAAAA"}
	for i := 1; i <= 2; i++ {
		if _, err := validar(f, malo); !errors.Is(err, domain.ErrUnauthorized) {
			t.Fatalf("intento %d: se esperaba 401, llegó %v", i, err)
		}
	}
	_, err := validar(f, malo)
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrLocked) {
		t.Fatalf("tercer intento: se esperaba 429, llegó %v", err)
	}
	if de.Message != service.MsgRegistroBloqueado || de.RetryAfter != service.BloqueoRegistro {
		t.Fatalf("bloqueo = %q %v", de.Message, de.RetryAfter)
	}
	if f.intentos.bloqueos[cedula] != service.BloqueoRegistro {
		t.Fatalf("no se creó el bloqueo de 10 minutos")
	}
}

func TestValidarIdentidad_BloqueoVigenteAunqueElDatoSeaCorrecto(t *testing.T) {
	f := newFixture()
	f.intentos.bloqueos[cedula] = 4 * time.Minute
	_, err := validar(f, service.IdentidadInput{Cedula: cedula, CodigoDactilar: "E5566V7788"})
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrLocked) {
		t.Fatalf("se esperaba 429, llegó %v", err)
	}
	if de.RetryAfter != 4*time.Minute {
		t.Fatalf("retry_after = %v, se esperaba el TTL restante", de.RetryAfter)
	}
	if len(f.tickets.emitidos) != 0 {
		t.Fatalf("no debe emitirse un ticket durante el bloqueo")
	}
}

func TestValidarIdentidad_CedulaYaRegistrada(t *testing.T) {
	f := newFixture()
	f.postulantes.registradas[cedula] = true
	_, err := validar(f, service.IdentidadInput{Cedula: cedula, CodigoDactilar: "E5566V7788"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("se esperaba 409, llegó %v", err)
	}
	if len(f.tickets.emitidos) != 0 {
		t.Fatalf("no debe emitirse un ticket para una cédula registrada")
	}
}

func TestValidarIdentidad_ErroresDeValidacion(t *testing.T) {
	casos := map[string]struct {
		in    service.IdentidadInput
		campo string
	}{
		"cédula mal formada":  {service.IdentidadInput{Cedula: "1234567890", CodigoDactilar: "E5566V7788"}, "cedula"},
		"sin dato":            {service.IdentidadInput{Cedula: cedula}, "fecha_expedicion"},
		"los dos datos":       {service.IdentidadInput{Cedula: cedula, FechaExpedicion: "2018-12-05", CodigoDactilar: "E5566V7788"}, "codigo_dactilar"},
		"fecha inválida":      {service.IdentidadInput{Cedula: cedula, FechaExpedicion: "31/02/2018"}, "fecha_expedicion"},
		"código muy corto":    {service.IdentidadInput{Cedula: cedula, CodigoDactilar: "E5566"}, "codigo_dactilar"},
		"código con símbolos": {service.IdentidadInput{Cedula: cedula, CodigoDactilar: "E5566-7788"}, "codigo_dactilar"},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			f := newFixture()
			_, err := validar(f, c.in)
			var de *domain.Error
			if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("se esperaba 400, llegó %v", err)
			}
			if de.Fields[c.campo] == "" {
				t.Fatalf("falta el campo %q en fields: %v", c.campo, de.Fields)
			}
			if len(f.intentos.fallos) != 0 {
				t.Fatalf("un error de formato no debe contar como intento fallido")
			}
		})
	}
}

func TestValidarIdentidad_ErrorDeInfraestructura(t *testing.T) {
	f := newFixture()
	f.intentos.err = errors.New("redis caído")
	_, err := validar(f, service.IdentidadInput{Cedula: cedula, CodigoDactilar: "E5566V7788"})
	var de *domain.Error
	if err == nil || errors.As(err, &de) {
		t.Fatalf("se esperaba el error interno sin convertir, llegó %v", err)
	}
}
