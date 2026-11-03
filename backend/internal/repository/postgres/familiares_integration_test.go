package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

// Estas pruebas también necesitan el PostgreSQL de docker compose (ver
// postulacion_integration_test.go).

func nombres(fams []domain.Familiar) []string {
	out := make([]string, len(fams))
	for i, f := range fams {
		out[i] = f.NombresApellidos
	}
	return out
}

func TestFamiliares_ConservanElOrdenDeCaptura(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	// Ni alfabético ni por inserción casual: 12 filas para cruzar los 4 bits bajos.
	capturado := []string{"Zoila", "Ana", "Marta", "Beto", "Yolanda", "Carla", "Xavier", "Diego", "Wilson", "Elena", "Víctor", "Fabián"}
	form := func(orden []string) domain.Formulario {
		f := domain.Formulario{Cargo: domain.CargoAfiliados, TieneFamiliares: true}
		for _, n := range orden {
			f.Familiares = append(f.Familiares, familiar(n))
		}
		return f
	}

	p, err := svc.Guardar(ctx, id, form(capturado))
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got := nombres(p.Formulario.Familiares); !slices.Equal(got, capturado) {
		t.Fatalf("orden tras guardar = %v, se esperaba %v", got, capturado)
	}
	// Al recargar sigue igual, también tras editar el orden (mover una fila al principio).
	reordenado := append([]string{"Fabián"}, capturado[:len(capturado)-1]...)
	if _, err := svc.Guardar(ctx, id, form(reordenado)); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	for i := 0; i < 3; i++ {
		p, err = svc.Obtener(ctx, id)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if got := nombres(p.Formulario.Familiares); !slices.Equal(got, reordenado) {
			t.Fatalf("orden al recargar = %v, se esperaba %v", got, reordenado)
		}
	}
}

func TestFamiliares_NoBorraLasFilas(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	if _, err := svc.Guardar(ctx, id, domain.Formulario{TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Ana"), familiar("Beto")}}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	p, err := svc.Guardar(ctx, id, domain.Formulario{TieneFamiliares: false})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Formulario.TieneFamiliares || len(p.Formulario.Familiares) != 0 {
		t.Fatalf("formulario = %+v", p.Formulario)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 0 {
		t.Fatalf("quedaron filas en la base: %v", got)
	}
}

func TestFamiliares_RechazosNoTocanLoGuardado(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	if _, err := svc.Guardar(ctx, id, domain.Formulario{TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Original")}}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	fuera := familiar("Otro")
	fuera.Parentesco = "primo"
	casos := map[string]domain.Formulario{
		"sí sin filas":      {TieneFamiliares: true},
		"valor fuera lista": {TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Nuevo"), fuera}},
		"no con filas":      {Familiares: []domain.Familiar{familiar("Nuevo")}},
	}
	for nombre, f := range casos {
		_, err := svc.Guardar(ctx, id, f)
		var de *domain.Error
		if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("%s: se esperaba 400, llegó %v", nombre, err)
		}
		if got := nombresFamiliares(t, pool, id); len(got) != 1 || got[0] != "Original" {
			t.Fatalf("%s: familiares en la base = %v, se esperaba [Original]", nombre, got)
		}
	}
}

// Una fila inválida en medio del guardado (aquí la segunda viola el CHECK de institucion,
// saltándose al servicio) deshace también el borrado y las filas ya insertadas.
func TestFamiliares_ROLLBACK_UnaFilaInvalidaDeshaceTodoElReemplazo(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)
	if _, err := svc.Guardar(ctx, id, domain.Formulario{TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Original")}}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	var postulacionID string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM postulaciones WHERE postulante_id = $1`, id).Scan(&postulacionID); err != nil {
		t.Fatal(err)
	}

	mala := familiar("Mala")
	mala.Institucion = "ISSFA"
	err := postgres.NewTxManager(pool).Ejecutar(ctx, func(ctx context.Context, tx domain.Tx) error {
		return postgres.NewFamiliaresRepo(pool).Reemplazar(ctx, tx, postulacionID, domain.Formulario{
			TieneFamiliares: true,
			Familiares:      []domain.Familiar{familiar("Buena"), mala, familiar("Tercera")},
		})
	})
	if err == nil {
		t.Fatal("se esperaba un error por la institución inválida")
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 1 || got[0] != "Original" {
		t.Fatalf("familiares tras el ROLLBACK = %v, se esperaba [Original]", got)
	}
}

// PUT /postulacion reemplaza el estado completo: guardar solo la sección de familiares
// borra el cargo, el domicilio y el consentimiento. Por eso el frontend lee la
// postulación, cambia su sección y envía el formulario entero.
func TestGuardar_ReemplazaElEstadoCompleto(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	paisID := 1
	inicial := domain.Formulario{
		Domicilio:      domain.Domicilio{PaisID: &paisID, TelefonoCelular: "0999999999", CallePrincipal: "Av. Amazonas"},
		Cargo:          domain.CargoAfiliados,
		Consentimiento: true,
	}
	if _, err := svc.Guardar(ctx, id, inicial); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	// Leer, cambiar solo familiares y enviar el formulario entero conserva lo demás.
	p, err := svc.Obtener(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	completo := p.Formulario
	completo.TieneFamiliares = true
	completo.Familiares = []domain.Familiar{familiar("Ana")}
	p, err = svc.Guardar(ctx, id, completo)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	f := p.Formulario
	if f.Cargo != domain.CargoAfiliados || !f.Consentimiento || f.Domicilio.CallePrincipal != "Av. Amazonas" ||
		f.Domicilio.PaisID == nil || *f.Domicilio.PaisID != paisID || len(f.Familiares) != 1 {
		t.Fatalf("el formulario entero debía conservar lo demás: %+v", f)
	}

	// Enviar solo la sección de familiares borra el resto.
	p, err = svc.Guardar(ctx, id, domain.Formulario{TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Ana")}})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	f = p.Formulario
	if f.Cargo != "" || f.Consentimiento || f.Domicilio.CallePrincipal != "" || f.Domicilio.PaisID != nil {
		t.Fatalf("enviar solo una sección debía borrar las demás: %+v", f)
	}
}
