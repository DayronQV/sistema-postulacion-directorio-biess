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

// fakeFases guarda las tres fases del concurso en memoria.
type fakeFases struct {
	fases        []domain.Fase
	actualizadas int
}

func nuevasFases() *fakeFases {
	return &fakeFases{fases: []domain.Fase{
		{ID: 1, Nombre: domain.FasePostulacion, Orden: 1, Publicada: true, HabilitaFormulario: true},
		{ID: 2, Nombre: domain.FaseMeritos, Orden: 2},
		{ID: 3, Nombre: domain.FaseEvaluacion, Orden: 3},
	}}
}

func (f *fakeFases) Listar(context.Context) ([]domain.Fase, error) { return f.fases, nil }

func (f *fakeFases) Obtener(_ context.Context, id int) (domain.Fase, error) {
	for _, fase := range f.fases {
		if fase.ID == id {
			return fase, nil
		}
	}
	return domain.Fase{}, domain.NotFound("La fase no existe")
}

func (f *fakeFases) PorNombre(_ context.Context, nombre string) (domain.Fase, error) {
	for _, fase := range f.fases {
		if fase.Nombre == nombre {
			return fase, nil
		}
	}
	return domain.Fase{}, domain.NotFound("La fase no existe")
}

func (f *fakeFases) Actualizar(_ context.Context, fase domain.Fase) error {
	for i := range f.fases {
		if f.fases[i].ID == fase.ID {
			f.fases[i] = fase
			f.actualizadas++
			return nil
		}
	}
	return domain.NotFound("La fase no existe")
}

// postulacion devuelve la fase Postulación del fake.
func (f *fakeFases) postulacion() *domain.Fase { return &f.fases[0] }

func TestFaseService_ActualizaPublicacionYFechas(t *testing.T) {
	fases := nuevasFases()
	svc := service.NewFaseService(fases)
	inicio := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	fin := inicio.AddDate(0, 0, 15)

	f, err := svc.Actualizar(context.Background(), 2, service.CambiosFase{Publicada: true, FechaInicio: &inicio, FechaFin: &fin})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !f.Publicada || f.Nombre != domain.FaseMeritos || !f.FechaInicio.Equal(inicio) || !f.FechaFin.Equal(fin) {
		t.Fatalf("fase = %+v", f)
	}
	if fases.actualizadas != 1 || !fases.fases[1].Publicada {
		t.Fatalf("no se guardó: %+v", fases.fases[1])
	}
}

func TestFaseService_SoloPostulacionHabilitaElFormulario(t *testing.T) {
	fases := nuevasFases()
	svc := service.NewFaseService(fases)

	_, err := svc.Actualizar(context.Background(), 2, service.CambiosFase{Publicada: true, HabilitaFormulario: true})
	if camposDe(t, err)["habilita_formulario"] == "" {
		t.Fatalf("se esperaba el campo habilita_formulario: %v", err)
	}
	if fases.actualizadas != 0 {
		t.Fatal("no debía guardar")
	}
}

func TestFaseService_HabilitarExigePublicar(t *testing.T) {
	svc := service.NewFaseService(nuevasFases())
	_, err := svc.Actualizar(context.Background(), 1, service.CambiosFase{HabilitaFormulario: true})
	if camposDe(t, err)["habilita_formulario"] == "" {
		t.Fatalf("se esperaba el campo habilita_formulario: %v", err)
	}
}

func TestFaseService_FinNoAnteriorAlInicio(t *testing.T) {
	svc := service.NewFaseService(nuevasFases())
	inicio := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	fin := inicio.Add(-time.Hour)
	_, err := svc.Actualizar(context.Background(), 1, service.CambiosFase{Publicada: true, FechaInicio: &inicio, FechaFin: &fin})
	if camposDe(t, err)["fecha_fin"] == "" {
		t.Fatalf("se esperaba el campo fecha_fin: %v", err)
	}
}

func TestFaseService_FaseInexistenteEs404(t *testing.T) {
	svc := service.NewFaseService(nuevasFases())
	if _, err := svc.Actualizar(context.Background(), 99, service.CambiosFase{}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("se esperaba 404, llegó %v", err)
	}
}

// fixturePeriodo es una postulación ya creada con las fases conectadas.
func fixturePeriodo(t *testing.T) (postFixture, *fakeFases) {
	t.Helper()
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	fases := nuevasFases()
	f.svc.ConFases(fases)
	if _, err := f.svc.Obtener(context.Background(), postulanteID); err != nil {
		t.Fatal(err)
	}
	return f, fases
}

func esPeriodoCerrado(err error) bool {
	return errors.Is(err, domain.ErrForbidden) && mensaje(err) == domain.MsgPeriodoCerrado
}

func TestPeriodo_AbiertoPermiteGuardarEInformaElEstado(t *testing.T) {
	f, _ := fixturePeriodo(t)
	p, err := f.svc.Guardar(context.Background(), postulanteID, domain.Formulario{})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !p.PeriodoAbierto {
		t.Fatal("periodo_abierto debía ser true")
	}
}

func TestPeriodo_CerradoBloqueaGuardarValidarYEnviar(t *testing.T) {
	ayer := time.Now().AddDate(0, 0, -1)
	cierres := map[string]func(*domain.Fase){
		"sin publicar":             func(f *domain.Fase) { f.Publicada = false },
		"formulario deshabilitado": func(f *domain.Fase) { f.HabilitaFormulario = false },
		"fuera de fechas":          func(f *domain.Fase) { f.FechaFin = &ayer },
	}
	for nombre, cerrar := range cierres {
		t.Run(nombre, func(t *testing.T) {
			f, fases := fixturePeriodo(t)
			cerrar(fases.postulacion())
			ctx := context.Background()

			if _, err := f.svc.Guardar(ctx, postulanteID, domain.Formulario{}); !esPeriodoCerrado(err) {
				t.Fatalf("Guardar: se esperaba 403 del periodo, llegó %v", err)
			}
			if _, err := f.svc.ValidarCargo(ctx, postulanteID, domain.CargoAfiliados); !esPeriodoCerrado(err) {
				t.Fatalf("ValidarCargo: se esperaba 403 del periodo, llegó %v", err)
			}
			if _, err := f.svc.Enviar(ctx, postulanteID); !esPeriodoCerrado(err) {
				t.Fatalf("Enviar: se esperaba 403 del periodo, llegó %v", err)
			}
			if f.store.actualizadas != 0 || f.store.enviadas != 0 || f.tx.ejecutadas != 0 {
				t.Fatalf("no debía tocar la base: %+v", f.store)
			}

			// Leer la postulación existente sigue funcionando y avisa que está cerrado.
			p, err := f.svc.Obtener(ctx, postulanteID)
			if err != nil {
				t.Fatalf("Obtener: error inesperado: %v", err)
			}
			if p.PeriodoAbierto {
				t.Fatal("periodo_abierto debía ser false")
			}
		})
	}
}

func TestPeriodo_CerradoNoCreaLaPostulacion(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	fases := nuevasFases()
	fases.postulacion().Publicada = false
	f.svc.ConFases(fases)

	if _, err := f.svc.Obtener(context.Background(), postulanteID); !esPeriodoCerrado(err) {
		t.Fatalf("se esperaba 403 del periodo, llegó %v", err)
	}
	if f.store.creadas != 0 {
		t.Fatal("no debía crear la postulación")
	}
}

func TestPeriodo_SinLaFasePostulacionNoRestringe(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	f.svc.ConFases(&fakeFases{})
	p, err := f.svc.Obtener(context.Background(), postulanteID)
	if err != nil || !p.PeriodoAbierto {
		t.Fatalf("sin la fase no hay restricción: p = %+v, err = %v", p.PeriodoAbierto, err)
	}
}
