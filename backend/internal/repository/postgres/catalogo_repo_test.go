package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

// newCatalogoRepo conecta a la base de docker compose con la semilla cargada. Si
// DATABASE_URL no está definida o la base no responde, la prueba se salta: las reglas
// del catálogo ya se prueban en el servicio con repositorios falsos.
func newCatalogoRepo(t *testing.T) *postgres.CatalogoRepo {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL no definida")
	}
	pool, err := postgres.NewPool(context.Background(), url)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	t.Cleanup(pool.Close)
	return postgres.NewCatalogoRepo(pool)
}

func TestCatalogoRepo_PaisesConEcuadorPrimero(t *testing.T) {
	repo := newCatalogoRepo(t)

	paises, err := repo.Paises(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(paises) != 194 {
		t.Fatalf("países = %d; la semilla tiene 194", len(paises))
	}
	if paises[0].Nombre != "Ecuador" {
		t.Fatalf("el primero es %q; se esperaba Ecuador", paises[0].Nombre)
	}
	if paises[1].Nombre != "Afganistán" || paises[2].Nombre != "Albania" {
		t.Fatalf("el resto no empieza por orden alfabético: %v", paises[1:3])
	}
	// El orden es el de la colación de la base, que ignora la tilde: Bélgica va antes
	// que Belice aunque "é" sea posterior a "i" byte a byte.
	if indice(paises, "Bélgica") > indice(paises, "Belice") {
		t.Fatalf("Bélgica debe ir antes que Belice")
	}
}

func indice(items []domain.ItemCatalogo, nombre string) int {
	for i, it := range items {
		if it.Nombre == nombre {
			return i
		}
	}
	return -1
}

func TestCatalogoRepo_ProvinciasYCantonesDeEcuador(t *testing.T) {
	ctx := context.Background()
	repo := newCatalogoRepo(t)
	paises, err := repo.Paises(ctx)
	if err != nil {
		t.Fatal(err)
	}

	provincias, err := repo.Provincias(ctx, paises[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(provincias) != 24 {
		t.Fatalf("provincias de Ecuador = %d; se esperaban 24", len(provincias))
	}
	if provincias[0].Nombre != "Azuay" || provincias[1].Nombre != "Bolívar" || provincias[2].Nombre != "Cañar" {
		t.Fatalf("orden alfabético con tildes incorrecto: %v", provincias[:3])
	}

	cantones, err := repo.Cantones(ctx, provincias[0].ID)
	if err != nil || len(cantones) == 0 {
		t.Fatalf("cantones de %s = %v, %v", provincias[0].Nombre, cantones, err)
	}
}

func TestCatalogoRepo_PadreSinHijosEsListaVaciaEInexistenteEsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newCatalogoRepo(t)

	// Los países distintos de Ecuador no tienen provincias en la semilla.
	paises, err := repo.Paises(ctx)
	if err != nil {
		t.Fatal(err)
	}
	vacias, err := repo.Provincias(ctx, paises[1].ID)
	if err != nil || vacias == nil || len(vacias) != 0 {
		t.Fatalf("Provincias de %s = %#v, %v; se esperaba una lista vacía", paises[1].Nombre, vacias, err)
	}

	if _, err := repo.Provincias(ctx, 999999); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Provincias(999999) = %v; se esperaba NotFound", err)
	}
	if _, err := repo.Cantones(ctx, 999999); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Cantones(999999) = %v; se esperaba NotFound", err)
	}
}
