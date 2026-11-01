package service_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// fakeCatalogos imita al repositorio de PostgreSQL y cuenta las consultas.
type fakeCatalogos struct {
	paises     []domain.ItemCatalogo
	provincias map[int][]domain.ItemCatalogo
	cantones   map[int][]domain.ItemCatalogo
	err        error
	llamadas   int
}

func (f *fakeCatalogos) Paises(context.Context) ([]domain.ItemCatalogo, error) {
	f.llamadas++
	return f.paises, f.err
}

func (f *fakeCatalogos) Provincias(_ context.Context, paisID int) ([]domain.ItemCatalogo, error) {
	f.llamadas++
	if f.err != nil {
		return nil, f.err
	}
	items, ok := f.provincias[paisID]
	if !ok {
		return nil, domain.NotFound("El país no existe")
	}
	return items, nil
}

func (f *fakeCatalogos) Cantones(_ context.Context, provinciaID int) ([]domain.ItemCatalogo, error) {
	f.llamadas++
	if f.err != nil {
		return nil, f.err
	}
	items, ok := f.cantones[provinciaID]
	if !ok {
		return nil, domain.NotFound("La provincia no existe")
	}
	return items, nil
}

// fakeCache imita a la caché de Redis en memoria. getErr y setErr simulan Redis caído.
type fakeCache struct {
	datos  map[string][]domain.ItemCatalogo
	ttls   map[string]time.Duration
	getErr error
	setErr error
	gets   int
	sets   int
}

func newFakeCache() *fakeCache {
	return &fakeCache{datos: map[string][]domain.ItemCatalogo{}, ttls: map[string]time.Duration{}}
}

func (f *fakeCache) Obtener(ctx context.Context, clave string) ([]domain.ItemCatalogo, bool, error) {
	f.gets++
	if err := ctx.Err(); err != nil {
		return nil, false, err // como go-redis con un contexto cancelado
	}
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	items, ok := f.datos[clave]
	return items, ok, nil
}

func (f *fakeCache) Guardar(_ context.Context, clave string, items []domain.ItemCatalogo, ttl time.Duration) error {
	f.sets++
	if f.setErr != nil {
		return f.setErr
	}
	f.datos[clave] = items
	f.ttls[clave] = ttl
	return nil
}

type catalogoFixture struct {
	svc       *service.CatalogoService
	catalogos *fakeCatalogos
	cache     *fakeCache
	logs      *bytes.Buffer
}

func newCatalogoFixture() catalogoFixture {
	f := catalogoFixture{
		catalogos: &fakeCatalogos{
			paises: []domain.ItemCatalogo{{ID: 60, Nombre: "Ecuador"}, {ID: 1, Nombre: "Afganistán"}},
			provincias: map[int][]domain.ItemCatalogo{
				60: {{ID: 1, Nombre: "Azuay"}, {ID: 2, Nombre: "Bolívar"}},
				1:  {},
			},
			cantones: map[int][]domain.ItemCatalogo{
				1: {{ID: 10, Nombre: "Cuenca"}},
			},
		},
		cache: newFakeCache(),
		logs:  &bytes.Buffer{},
	}
	f.svc = service.NewCatalogoService(f.catalogos, f.cache, slog.New(slog.NewTextHandler(f.logs, nil)))
	return f
}

func TestCatalogo_FalloDeCache_ConsultaYGuardaConTTLDeUnaHora(t *testing.T) {
	f := newCatalogoFixture()

	got, err := f.svc.Paises(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, f.catalogos.paises) {
		t.Fatalf("Paises = %v", got)
	}
	if f.catalogos.llamadas != 1 {
		t.Fatalf("llamadas a PostgreSQL = %d; se esperaba 1", f.catalogos.llamadas)
	}
	if !reflect.DeepEqual(f.cache.datos["paises"], f.catalogos.paises) {
		t.Fatalf("la lista no se guardó en la caché: %v", f.cache.datos)
	}
	if f.cache.ttls["paises"] != time.Hour {
		t.Fatalf("TTL = %v; se esperaba 1h", f.cache.ttls["paises"])
	}
}

func TestCatalogo_AciertoDeCache_NoConsultaPostgres(t *testing.T) {
	f := newCatalogoFixture()
	enCache := []domain.ItemCatalogo{{ID: 7, Nombre: "Desde la caché"}}
	f.cache.datos["paises"] = enCache

	got, err := f.svc.Paises(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, enCache) {
		t.Fatalf("Paises = %v; se esperaba lo guardado en la caché", got)
	}
	if f.catalogos.llamadas != 0 || f.cache.sets != 0 {
		t.Fatalf("un acierto no debe consultar ni escribir: PostgreSQL %d, escrituras %d", f.catalogos.llamadas, f.cache.sets)
	}
}

func TestCatalogo_SegundaLecturaSaleDeLaCache(t *testing.T) {
	f := newCatalogoFixture()

	for i := 0; i < 3; i++ {
		if _, err := f.svc.Provincias(context.Background(), 60); err != nil {
			t.Fatal(err)
		}
	}
	if f.catalogos.llamadas != 1 {
		t.Fatalf("llamadas a PostgreSQL = %d; se esperaba 1", f.catalogos.llamadas)
	}
}

func TestCatalogo_RedisCaido_RespondeDesdePostgresYRegistraElAviso(t *testing.T) {
	f := newCatalogoFixture()
	f.cache.getErr = errors.New("connection refused")

	got, err := f.svc.Paises(context.Background())
	if err != nil {
		t.Fatalf("Redis caído no debe tumbar el servicio: %v", err)
	}
	if !reflect.DeepEqual(got, f.catalogos.paises) {
		t.Fatalf("Paises = %v", got)
	}
	if f.cache.sets != 0 {
		t.Fatalf("con la lectura caída no se debe intentar guardar (%d escrituras)", f.cache.sets)
	}
	if log := f.logs.String(); !strings.Contains(log, "level=WARN") || !strings.Contains(log, "connection refused") {
		t.Fatalf("no se registró el aviso: %q", log)
	}
}

func TestCatalogo_PeticionCancelada_NoRegistraAvisoNiEscribeEnCache(t *testing.T) {
	f := newCatalogoFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // el cliente cortó la conexión

	if _, err := f.svc.Paises(ctx); err != nil {
		t.Fatalf("el falso de PostgreSQL ignora el contexto; no debería fallar: %v", err)
	}
	if f.logs.Len() != 0 {
		t.Fatalf("una petición cancelada no es una caché caída; no debe haber aviso: %q", f.logs.String())
	}
	if f.cache.sets != 0 {
		t.Fatalf("no se debe escribir en la caché con la petición cancelada (%d escrituras)", f.cache.sets)
	}
}

func TestCatalogo_FalloAlGuardar_RespondeIgualYRegistraElAviso(t *testing.T) {
	f := newCatalogoFixture()
	f.cache.setErr = errors.New("OOM command not allowed")

	got, err := f.svc.Paises(context.Background())
	if err != nil {
		t.Fatalf("un fallo al guardar no debe tumbar el servicio: %v", err)
	}
	if !reflect.DeepEqual(got, f.catalogos.paises) {
		t.Fatalf("Paises = %v", got)
	}
	if !strings.Contains(f.logs.String(), "OOM command not allowed") {
		t.Fatalf("no se registró el aviso: %q", f.logs.String())
	}
}

func TestCatalogo_ClavesPorRecurso(t *testing.T) {
	f := newCatalogoFixture()

	if _, err := f.svc.Provincias(context.Background(), 60); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Cantones(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	for _, clave := range []string{"provincias:60", "cantones:1"} {
		if _, ok := f.cache.datos[clave]; !ok {
			t.Errorf("falta la clave %q en la caché: %v", clave, f.cache.datos)
		}
	}
}

func TestCatalogo_IdInexistente_NotFoundYNoSeGuarda(t *testing.T) {
	f := newCatalogoFixture()

	_, err := f.svc.Provincias(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Provincias(999) = %v; se esperaba NotFound", err)
	}
	_, err = f.svc.Cantones(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Cantones(999) = %v; se esperaba NotFound", err)
	}
	if f.cache.sets != 0 {
		t.Fatalf("lo que no existe no se guarda en la caché (%d escrituras)", f.cache.sets)
	}
}

func TestCatalogo_PadreSinHijos_ListaVaciaNoNil(t *testing.T) {
	f := newCatalogoFixture()
	f.catalogos.provincias[2] = nil

	got, err := f.svc.Provincias(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("Provincias(2) = %#v; se esperaba una lista vacía no nula", got)
	}
	if enCache, ok := f.cache.datos["provincias:2"]; !ok || enCache == nil {
		t.Fatalf("la lista vacía se guarda para no repetir la consulta: %#v", f.cache.datos)
	}
}

func TestCatalogo_ErrorDeLaBase_SePropagaYNoSeGuarda(t *testing.T) {
	f := newCatalogoFixture()
	f.catalogos.err = errors.New("conexión perdida")

	_, err := f.svc.Paises(context.Background())
	if err == nil || !strings.Contains(err.Error(), "conexión perdida") {
		t.Fatalf("Paises = %v; se esperaba el error de la base", err)
	}
	if f.cache.sets != 0 {
		t.Fatalf("no se guarda nada si la base falló (%d escrituras)", f.cache.sets)
	}
}
