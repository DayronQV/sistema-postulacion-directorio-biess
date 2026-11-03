package postgres_test

import (
	"context"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func TestReferencias_SeGuardanEnLaTransaccionYSeRecargan(t *testing.T) {
	pool := newPool(t)
	postulanteID := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool, postgres.NewReferenciasRepo(pool))
	if _, err := svc.Obtener(context.Background(), postulanteID); err != nil {
		t.Fatal(err)
	}
	referencia := domain.ReferenciaPersonal{
		NombresApellidos: "Ana Pérez",
		Relacion:         "Amiga",
		Telefono:         "0999999999",
		Correo:           "ana@example.com",
	}
	guardada, err := svc.Guardar(context.Background(), postulanteID, domain.Formulario{
		Referencias: []domain.ReferenciaPersonal{referencia},
	})
	if err != nil {
		t.Fatalf("guardando: %v", err)
	}
	if len(guardada.Formulario.Referencias) != 1 || guardada.Formulario.Referencias[0] != referencia {
		t.Fatalf("referencias recargadas = %+v", guardada.Formulario.Referencias)
	}
}
