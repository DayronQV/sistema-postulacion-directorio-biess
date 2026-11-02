package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// registroConNacionalidad devuelve siempre el mismo ciudadano con la nacionalidad dada,
// para probar cómo escribe el Registro Civil el gentilicio.
type registroConNacionalidad struct{ nacionalidad string }

func (r registroConNacionalidad) ConsultarCiudadano(ctx context.Context, cedula string) (domain.Ciudadano, error) {
	c, err := registrocivil.NewMock().ConsultarCiudadano(ctx, registrocivil.CedulaFlujoFeliz)
	c.Nacionalidad = r.nacionalidad
	return c, err
}

func newPostFixtureConRegistro(registro domain.RegistroCivil) postFixture {
	store := &fakePostulaciones{cedula: registrocivil.CedulaFlujoFeliz}
	tx := &fakeTx{store: store}
	llamadas := &[]string{}
	svc := service.NewPostulacionService(store, registro, tx, fakeSeccion{nombre: "familiares", llamadas: llamadas}).ConHistorialLaboral(historiallaboral.NewMock())
	return postFixture{svc: svc, store: store, tx: tx, llamadas: llamadas}
}

func TestGuardar_LaNacionalidadEcuatorianaNoDependeDeMayusculasNiTildes(t *testing.T) {
	for _, nacionalidad := range []string{"ECUATORIANA", "ecuatoriana", "  Ecuatoriana ", "Ecuatoriano", "Écuatoriana"} {
		t.Run(nacionalidad, func(t *testing.T) {
			f := newPostFixtureConRegistro(registroConNacionalidad{nacionalidad})

			p, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido())
			if err != nil {
				t.Fatalf("una persona ecuatoriana no debe bloquearse: %v", err)
			}
			if p.Estado != domain.EstadoBorrador || f.store.actualizadas != 1 {
				t.Fatalf("estado = %q, actualizadas = %d", p.Estado, f.store.actualizadas)
			}
		})
	}
}

func TestGuardar_OtrasNacionalidadesCierranLaPostulacion(t *testing.T) {
	for _, nacionalidad := range []string{"Colombiana", "Peruana", "", "No ecuatoriana", "Ecuador"} {
		t.Run(nacionalidad, func(t *testing.T) {
			f := newPostFixtureConRegistro(registroConNacionalidad{nacionalidad})

			_, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido())
			if !errors.Is(err, domain.ErrUnprocessable) {
				t.Fatalf("se esperaba 422, llegó %v", err)
			}
			if f.store.p.Estado != domain.EstadoCerrada {
				t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
			}
		})
	}
}

// Un cuerpo inválido no debe ocultar el cierre: la primera llamada de un extranjero es
// siempre 422 y la postulación queda cerrada, aunque el formulario traiga errores.
func TestGuardar_ExtranjeroConCuerpoInvalidoSeCierraIgual(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaExtranjera, nil)
	invalido := formularioValido()
	invalido.Cargo = "gerente"
	invalido.TieneFamiliares = false

	_, err := f.svc.Guardar(context.Background(), postulanteID, invalido)
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrUnprocessable) || de.Message != service.MsgNacionalidadNoEcuatoriana {
		t.Fatalf("se esperaba 422 %q, llegó %v", service.MsgNacionalidadNoEcuatoriana, err)
	}
	if f.store.p.Estado != domain.EstadoCerrada || f.store.actualizadas != 0 {
		t.Fatalf("estado = %q, actualizadas = %d", f.store.p.Estado, f.store.actualizadas)
	}

	// Desde entonces solo hay conflicto, con cualquier cuerpo.
	for _, fm := range []domain.Formulario{invalido, formularioValido(), {}} {
		if _, err := f.svc.Guardar(context.Background(), postulanteID, fm); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("se esperaba 409, llegó %v", err)
		}
	}
}

func TestObtener_TrasElCierreDevuelveLaPostulacionCerrada(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaExtranjera, nil)
	if _, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido()); !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("se esperaba 422, llegó %v", err)
	}

	p, err := f.svc.Obtener(context.Background(), postulanteID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Estado != domain.EstadoCerrada || p.Personal.Nacionalidad != "Colombiana" {
		t.Fatalf("postulación = %+v", p)
	}
}

func TestObtener_DatosDeLaParejaSegunElEstadoCivil(t *testing.T) {
	casos := map[string]struct {
		cedula      string
		estadoCivil domain.EstadoCivil
		pareja      *domain.Pareja
	}{
		"casada":           {registrocivil.CedulaCasada, domain.EstadoCasado, &domain.Pareja{NombresApellidos: "Luis Alberto Párraga Mera", Cedula: "0913579249"}},
		"unión de hecho":   {registrocivil.CedulaUnionDeHecho, domain.EstadoUnionDeHecho, &domain.Pareja{NombresApellidos: "Ana Lucía Moreira Zambrano", Cedula: "1313579243"}},
		"soltera":          {registrocivil.CedulaFlujoFeliz, domain.EstadoSoltero, nil},
		"divorciada":       {registrocivil.CedulaRecuperacion, domain.EstadoDivorciado, nil},
		"extranjera, sola": {registrocivil.CedulaExtranjera, domain.EstadoSoltero, nil},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			f := newPostFixture(c.cedula, nil)

			p, err := f.svc.Obtener(context.Background(), postulanteID)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if p.Personal.EstadoCivil != c.estadoCivil {
				t.Fatalf("estado civil = %q, se esperaba %q", p.Personal.EstadoCivil, c.estadoCivil)
			}
			switch {
			case c.pareja == nil && p.Personal.Pareja != nil:
				t.Fatalf("no debía traer pareja: %+v", p.Personal.Pareja)
			case c.pareja != nil && (p.Personal.Pareja == nil || *p.Personal.Pareja != *c.pareja):
				t.Fatalf("pareja = %+v, se esperaba %+v", p.Personal.Pareja, c.pareja)
			}
		})
	}
}
