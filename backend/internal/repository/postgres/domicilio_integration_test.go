package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// Estas pruebas usan la semilla real de la división política de db/schema.sql: los ids se buscan por
// nombre, no se escriben a mano.

func newServiceDomicilio(pool *pgxpool.Pool) *service.PostulacionService {
	return newService(pool).ConCatalogos(postgres.NewCatalogoRepo(pool))
}

func idDe(t *testing.T, pool *pgxpool.Pool, consulta string, args ...any) int {
	t.Helper()
	var id int
	if err := pool.QueryRow(context.Background(), consulta, args...).Scan(&id); err != nil {
		t.Fatalf("buscando un id de la semilla (%s %v): %v", consulta, args, err)
	}
	return id
}

type idsGeo struct {
	ecuador, colombia int
	pichincha, guayas int
	quito, guayaquil  int
}

func idsDeLaSemilla(t *testing.T, pool *pgxpool.Pool) idsGeo {
	t.Helper()
	var g idsGeo
	g.ecuador = idDe(t, pool, `SELECT id FROM paises WHERE nombre = 'Ecuador'`)
	g.colombia = idDe(t, pool, `SELECT id FROM paises WHERE nombre = 'Colombia'`)
	provincia := func(nombre string) int {
		return idDe(t, pool, `SELECT id FROM provincias WHERE nombre = $1 AND pais_id = $2`, nombre, g.ecuador)
	}
	canton := func(provincia int, nombre string) int {
		return idDe(t, pool, `SELECT id FROM cantones WHERE nombre = $1 AND provincia_id = $2`, nombre, provincia)
	}
	g.pichincha, g.guayas = provincia("Pichincha"), provincia("Guayas")
	g.quito, g.guayaquil = canton(g.pichincha, "Quito"), canton(g.guayas, "Guayaquil")
	return g
}

// leerDomicilio lee las columnas del domicilio directamente de la tabla.
func leerDomicilio(t *testing.T, pool *pgxpool.Pool, postulanteID string) (pais, provincia, canton *int, provTexto, ciudadTexto, celular *string) {
	t.Helper()
	err := pool.QueryRow(context.Background(), `
SELECT pais_id, provincia_id, canton_id, provincia_texto, ciudad_texto, telefono_celular
FROM postulaciones WHERE postulante_id = $1`, postulanteID,
	).Scan(&pais, &provincia, &canton, &provTexto, &ciudadTexto, &celular)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestGuardarDomicilio_RechazaCantonQueNoEsDeLaProvincia(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	g := idsDeLaSemilla(t, pool)
	svc := newServiceDomicilio(pool)
	if _, err := svc.Obtener(ctx, id); err != nil {
		t.Fatal(err)
	}

	// Quito pertenece a Pichincha, no a Guayas.
	_, err := svc.Guardar(ctx, id, domain.Formulario{Domicilio: domain.Domicilio{
		PaisID: &g.ecuador, ProvinciaID: &g.guayas, CantonID: &g.quito, TelefonoCelular: "0999999999",
	}})

	var de *domain.Error
	if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &de) || de.Fields["domicilio.canton_id"] == "" {
		t.Fatalf("se esperaba 400 con el error de domicilio.canton_id, llegó %v", err)
	}
	pais, _, _, _, _, celular := leerDomicilio(t, pool, id)
	if pais != nil || celular != nil {
		t.Fatalf("el rechazo no debe dejar nada escrito: pais=%v celular=%v", pais, celular)
	}
}

func TestGuardarDomicilio_RechazaIdsConUnPaisExtranjero(t *testing.T) {
	pool := newPool(t)
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	g := idsDeLaSemilla(t, pool)
	svc := newServiceDomicilio(pool)

	// La combinación incoherente que la base sola acepta (cada clave foránea es válida):
	// un país que no es Ecuador con una provincia y un cantón de Ecuador.
	_, err := svc.Guardar(context.Background(), id, domain.Formulario{Domicilio: domain.Domicilio{
		PaisID: &g.colombia, ProvinciaID: &g.pichincha, CantonID: &g.quito, ProvinciaTexto: "x", CiudadTexto: "y",
	}})

	var de *domain.Error
	if !errors.Is(err, domain.ErrValidation) || !errors.As(err, &de) ||
		de.Fields["domicilio.provincia_id"] == "" || de.Fields["domicilio.canton_id"] == "" {
		t.Fatalf("se esperaba 400 por los ids con un país extranjero, llegó %v", err)
	}

}

func TestGuardarDomicilio_EcuadorValidoSeGuardaYSeLee(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	g := idsDeLaSemilla(t, pool)
	svc := newServiceDomicilio(pool)

	p, err := svc.Guardar(ctx, id, domain.Formulario{Domicilio: domain.Domicilio{
		PaisID: &g.ecuador, ProvinciaID: &g.pichincha, CantonID: &g.quito,
		TelefonoCelular: "0999999999", EmailSecundario: "otro@correo.com", CallePrincipal: "Av. Amazonas",
	}})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	pais, provincia, canton, provTexto, ciudadTexto, celular := leerDomicilio(t, pool, id)
	if pais == nil || *pais != g.ecuador || provincia == nil || *provincia != g.pichincha ||
		canton == nil || *canton != g.quito || provTexto != nil || ciudadTexto != nil ||
		celular == nil || *celular != "0999999999" {
		t.Fatalf("domicilio guardado incorrecto: %v %v %v %v %v %v", pais, provincia, canton, provTexto, ciudadTexto, celular)
	}
	// El correo 1 es el del registro (tabla postulantes) y viaja en la lectura.
	if want := registrocivil.CedulaFlujoFeliz + "@prueba.test"; p.Personal.Email != want {
		t.Fatalf("correo del registro = %q, se esperaba %q", p.Personal.Email, want)
	}
}

func TestGuardarDomicilio_OtroPaisPorTexto(t *testing.T) {
	pool := newPool(t)
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	g := idsDeLaSemilla(t, pool)
	svc := newServiceDomicilio(pool)

	if _, err := svc.Guardar(context.Background(), id, domain.Formulario{Domicilio: domain.Domicilio{
		PaisID: &g.colombia, ProvinciaTexto: "Antioquia", CiudadTexto: "Medellín",
	}}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	pais, provincia, canton, provTexto, ciudadTexto, _ := leerDomicilio(t, pool, id)
	if pais == nil || *pais != g.colombia || provincia != nil || canton != nil ||
		provTexto == nil || *provTexto != "Antioquia" || ciudadTexto == nil || *ciudadTexto != "Medellín" {
		t.Fatalf("domicilio guardado incorrecto: %v %v %v %v %v", pais, provincia, canton, provTexto, ciudadTexto)
	}
}

// PUT /postulacion reemplaza el estado completo: un guardado que solo trae el domicilio
// borra los familiares. Por eso el frontend lee la postulación, cambia su sección y
// envía el formulario entero.
func TestGuardarDomicilio_PutReemplazaTodoYMezclarConservaLasOtrasSecciones(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	g := idsDeLaSemilla(t, pool)
	svc := newServiceDomicilio(pool)

	if _, err := svc.Guardar(ctx, id, domain.Formulario{
		Cargo: domain.CargoAfiliados, TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Ana")},
	}); err != nil {
		t.Fatalf("estado inicial: %v", err)
	}
	domicilio := domain.Domicilio{PaisID: &g.ecuador, ProvinciaID: &g.guayas, CantonID: &g.guayaquil, TelefonoCelular: "0988888888"}

	// Mezclar: leer, cambiar solo el domicilio y enviar todo.
	actual, err := svc.Obtener(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	mezclado := actual.Formulario
	mezclado.Domicilio = domicilio
	p, err := svc.Guardar(ctx, id, mezclado)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 1 || got[0] != "Ana" || p.Formulario.Cargo != domain.CargoAfiliados {
		t.Fatalf("mezclar debía conservar las otras secciones: familiares=%v cargo=%q", got, p.Formulario.Cargo)
	}

	// Enviar solo la sección pisa lo demás.
	if _, err := svc.Guardar(ctx, id, domain.Formulario{Domicilio: domicilio}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 0 {
		t.Fatalf("un PUT con solo el domicilio debía borrar los familiares, quedan %v", got)
	}
}
