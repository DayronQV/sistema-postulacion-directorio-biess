package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

const idPostulacion = "11111111-2222-3333-4444-555555555555"

// fakeMeritos imita la tabla postulaciones: estado y marca por id.
type fakeMeritos struct {
	estados map[string]string
	meritos map[string]bool
	// cambiaAntesDeEscribir simula que la postulación dejó de estar enviada entre la
	// lectura y la escritura.
	cambiaAntesDeEscribir bool
}

func (f *fakeMeritos) EstadoPorID(_ context.Context, id string) (string, error) {
	e, ok := f.estados[id]
	if !ok {
		return "", domain.NotFound("La postulación no existe")
	}
	return e, nil
}

func (f *fakeMeritos) ActualizarMerito(_ context.Context, id string, supero bool) (bool, error) {
	if f.cambiaAntesDeEscribir || f.estados[id] != domain.EstadoEnviada {
		return false, nil
	}
	f.meritos[id] = supero
	return true, nil
}

func nuevoMeritos(estado string) (*service.AdminPostulacionService, *fakeMeritos) {
	f := &fakeMeritos{estados: map[string]string{idPostulacion: estado}, meritos: map[string]bool{}}
	return service.NewAdminPostulacionService(f), f
}

func TestMarcarMerito_EnviadaPasaAEvaluacionTecnica(t *testing.T) {
	svc, f := nuevoMeritos(domain.EstadoEnviada)
	r, err := svc.MarcarMerito(context.Background(), idPostulacion, true)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !f.meritos[idPostulacion] || !r.SuperoMerito || r.EstadoVisible != domain.EstadoVisibleEvaluacionTecnica {
		t.Fatalf("resultado = %+v, marca = %v", r, f.meritos[idPostulacion])
	}
}

func TestMarcarMerito_SePuedeQuitarLaMarca(t *testing.T) {
	svc, f := nuevoMeritos(domain.EstadoEnviada)
	f.meritos[idPostulacion] = true
	r, err := svc.MarcarMerito(context.Background(), idPostulacion, false)
	if err != nil || f.meritos[idPostulacion] || r.EstadoVisible != domain.EstadoVisibleEnviada {
		t.Fatalf("resultado = %+v, err = %v", r, err)
	}
}

func TestMarcarMerito_SoloPostulacionesEnviadas(t *testing.T) {
	for _, estado := range []string{domain.EstadoBorrador, domain.EstadoCerrada} {
		svc, f := nuevoMeritos(estado)
		_, err := svc.MarcarMerito(context.Background(), idPostulacion, true)
		if !errors.Is(err, domain.ErrConflict) || mensaje(err) != service.MsgMeritoSoloEnviada {
			t.Fatalf("%s: se esperaba 409, llegó %v", estado, err)
		}
		if f.meritos[idPostulacion] {
			t.Fatalf("%s: no debe marcarse", estado)
		}
	}
}

func TestMarcarMerito_CambioDeEstadoConcurrenteEs409(t *testing.T) {
	svc, f := nuevoMeritos(domain.EstadoEnviada)
	f.cambiaAntesDeEscribir = true
	if _, err := svc.MarcarMerito(context.Background(), idPostulacion, true); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("se esperaba 409, llegó %v", err)
	}
}

func TestMarcarMerito_PostulacionInexistenteOIdInvalidoEs404(t *testing.T) {
	svc, _ := nuevoMeritos(domain.EstadoEnviada)
	for _, id := range []string{"99999999-2222-3333-4444-555555555555", "no-es-un-uuid", ""} {
		if _, err := svc.MarcarMerito(context.Background(), id, true); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("id %q: se esperaba 404, llegó %v", id, err)
		}
	}
}
