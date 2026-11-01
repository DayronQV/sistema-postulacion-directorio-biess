package postgres_test

import (
	"context"
	"errors"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// Estas pruebas usan el PostgreSQL de docker compose (db/schema.sql ya cargado). Si
// DATABASE_URL no está definida, se saltan: las reglas del servicio ya se prueban con
// repositorios falsos, pero solo aquí se demuestra el COMMIT y el ROLLBACK reales.
// Necesitan una base limpia (docker compose down -v) con el esquema actual.

// newPool abre el pool de pruebas y comprueba que la base tenga el esquema actual.
func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL no definida: se salta la prueba contra PostgreSQL")
	}
	pool, err := postgres.NewPool(context.Background(), url)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	t.Cleanup(pool.Close)
	// La secuencia de los códigos de postulación vive en db/schema.sql: no se crea aquí
	// para que las pruebas fallen si el esquema no la trae.
	var secuencia *string
	if err := pool.QueryRow(context.Background(), `SELECT to_regclass('postulacion_codigo_seq')::text`).Scan(&secuencia); err != nil {
		t.Fatalf("consultando la secuencia: %v", err)
	}
	if secuencia == nil {
		t.Fatal("falta la secuencia postulacion_codigo_seq: recree la base con docker compose down -v")
	}
	return pool
}

// crearPostulante inserta un postulante con la cédula dada (que existe en el Registro
// Civil simulado) y lo borra al terminar. Si la cédula ya está registrada en la base, la
// prueba falla en lugar de saltarse: saltarla dejaría en verde una verificación que no
// corrió. El simulador solo tiene 7 cédulas, así que no se pueden usar cédulas nuevas.
func crearPostulante(t *testing.T, pool *pgxpool.Pool, cedula string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx, `
INSERT INTO postulantes (cedula, email, password_hash)
VALUES ($1, $2, 'x')
ON CONFLICT DO NOTHING
RETURNING id::text`, cedula, cedula+"@prueba.test").Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("la cédula %s ya está registrada en la base: use una base limpia (docker compose down -v)", cedula)
	}
	if err != nil {
		t.Fatalf("creando el postulante: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM postulaciones WHERE postulante_id = $1`, id)
		_, _ = pool.Exec(ctx, `DELETE FROM postulantes WHERE id = $1`, id)
	})
	return id
}

func newService(pool *pgxpool.Pool, extra ...domain.SeccionPostulacion) *service.PostulacionService {
	secciones := append([]domain.SeccionPostulacion{postgres.NewFamiliaresRepo(pool)}, extra...)
	return service.NewPostulacionService(
		postgres.NewPostulacionRepo(pool),
		registrocivil.NewMock(),
		postgres.NewTxManager(pool),
		secciones...,
	)
}

// seccionRota escribe una fila inválida (viola el CHECK de institucion) dentro de la
// transacción, después de que la postulación y los familiares ya se escribieron. Como
// PostgreSQL deja la transacción abortada, aquí hasta un COMMIT mal hecho se deshace: la
// prueba de que el TxManager hace ROLLBACK es seccionQueFallaTrasEscribir.
type seccionRota struct{}

func (seccionRota) Cargar(context.Context, string, *domain.Formulario) error { return nil }

func (seccionRota) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, _ domain.Formulario) error {
	_, err := tx.(pgx.Tx).Exec(ctx, `
INSERT INTO postulacion_familiares (postulacion_id, nombres_apellidos, parentesco, institucion, area, cargo)
VALUES ($1, 'Fila inválida', 'primo', 'ISSFA', 'x', 'x')`, postulacionID)
	return err
}

// errTrasEscribir es el error de Go (sin error de SQL) de seccionQueFallaTrasEscribir.
var errTrasEscribir = errors.New("la sección falla después de escribir")

// seccionQueFallaTrasEscribir escribe una fila válida y luego devuelve un error de Go.
// La transacción sigue sana en PostgreSQL: solo un ROLLBACK explícito del TxManager
// deshace la fila. Si el TxManager confirmara pese al error, la fila quedaría guardada.
type seccionQueFallaTrasEscribir struct{}

func (seccionQueFallaTrasEscribir) Cargar(context.Context, string, *domain.Formulario) error {
	return nil
}

func (seccionQueFallaTrasEscribir) Reemplazar(ctx context.Context, tx domain.Tx, postulacionID string, _ domain.Formulario) error {
	_, err := tx.(pgx.Tx).Exec(ctx, `
INSERT INTO postulacion_familiares (postulacion_id, nombres_apellidos, parentesco, institucion, area, cargo)
VALUES ($1, 'Fila que debe deshacerse', 'primo', 'IESS', 'x', 'x')`, postulacionID)
	if err != nil {
		return err
	}
	return errTrasEscribir
}

func familiar(nombre string) domain.Familiar {
	return domain.Familiar{NombresApellidos: nombre, Parentesco: "hermano/a", Institucion: "IESS", Area: "Finanzas", Cargo: "Analista"}
}

func nombresFamiliares(t *testing.T, pool *pgxpool.Pool, postulanteID string) []string {
	t.Helper()
	rows, err := pool.Query(context.Background(), `
SELECT f.nombres_apellidos FROM postulacion_familiares f
JOIN postulaciones p ON p.id = f.postulacion_id
WHERE p.postulante_id = $1 ORDER BY f.nombres_apellidos`, postulanteID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		out = append(out, n)
	}
	return out
}

func TestGuardar_COMMIT_PersisteTodoEnUnaTransaccion(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	paisID := 1
	p, err := svc.Guardar(ctx, id, domain.Formulario{
		Domicilio:       domain.Domicilio{PaisID: &paisID, TelefonoCelular: "0999999999"},
		Cargo:           domain.CargoAfiliados,
		TieneFamiliares: true,
		Familiares:      []domain.Familiar{familiar("Ana"), familiar("Beto")},
		Consentimiento:  true,
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Formulario.Cargo != domain.CargoAfiliados || p.Formulario.Domicilio.TelefonoCelular != "0999999999" ||
		!p.Formulario.Consentimiento || len(p.Formulario.Familiares) != 2 {
		t.Fatalf("no se guardó el formulario: %+v", p.Formulario)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 2 {
		t.Fatalf("familiares en la base = %v", got)
	}

	// Un segundo PUT reemplaza el estado completo, no lo acumula.
	p, err = svc.Guardar(ctx, id, domain.Formulario{Cargo: domain.CargoJubilados, TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Carla")}})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 1 || got[0] != "Carla" {
		t.Fatalf("familiares en la base = %v, se esperaba [Carla]", got)
	}
	if p.Formulario.Cargo != domain.CargoJubilados || p.Formulario.Consentimiento {
		t.Fatalf("el segundo PUT no reemplazó el estado: %+v", p.Formulario)
	}
}

func TestGuardar_ROLLBACK_ErrorEnTablaHijaNoDejaNadaGuardado(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)

	// Estado inicial confirmado: cargo "afiliados" y un familiar.
	bueno := newService(pool)
	if _, err := bueno.Guardar(ctx, id, domain.Formulario{
		Cargo: domain.CargoAfiliados, TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Original")},
	}); err != nil {
		t.Fatalf("estado inicial: %v", err)
	}

	// Este guardado escribe la postulación y los familiares nuevos y luego falla en una
	// tabla hija: todo debe deshacerse, incluido el borrado de los familiares anteriores.
	roto := newService(pool, seccionRota{})
	_, err := roto.Guardar(ctx, id, domain.Formulario{
		Cargo:           domain.CargoJubilados,
		TieneFamiliares: true,
		Familiares:      []domain.Familiar{familiar("Nuevo 1"), familiar("Nuevo 2")},
		Consentimiento:  true,
	})
	if err == nil {
		t.Fatal("se esperaba el error de la tabla hija")
	}

	var cargo string
	var consentimiento bool
	if err := pool.QueryRow(ctx,
		`SELECT cargo, consentimiento_at IS NOT NULL FROM postulaciones WHERE postulante_id = $1`, id,
	).Scan(&cargo, &consentimiento); err != nil {
		t.Fatal(err)
	}
	if cargo != domain.CargoAfiliados || consentimiento {
		t.Fatalf("la postulación quedó con escrituras parciales: cargo=%q consentimiento=%v", cargo, consentimiento)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 1 || got[0] != "Original" {
		t.Fatalf("familiares tras el ROLLBACK = %v, se esperaba [Original]", got)
	}
}

func TestGuardar_ROLLBACK_ErrorDeGoTrasEscribirDeshaceLoEscrito(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)

	if _, err := newService(pool).Guardar(ctx, id, domain.Formulario{
		Cargo: domain.CargoAfiliados, TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Original")},
	}); err != nil {
		t.Fatalf("estado inicial: %v", err)
	}

	_, err := newService(pool, seccionQueFallaTrasEscribir{}).Guardar(ctx, id, domain.Formulario{
		Cargo:           domain.CargoJubilados,
		TieneFamiliares: true,
		Familiares:      []domain.Familiar{familiar("Nuevo")},
		Consentimiento:  true,
	})
	if !errors.Is(err, errTrasEscribir) {
		t.Fatalf("se esperaba el error de la sección, llegó %v", err)
	}

	var cargo string
	var consentimiento bool
	if err := pool.QueryRow(ctx,
		`SELECT cargo, consentimiento_at IS NOT NULL FROM postulaciones WHERE postulante_id = $1`, id,
	).Scan(&cargo, &consentimiento); err != nil {
		t.Fatal(err)
	}
	if cargo != domain.CargoAfiliados || consentimiento {
		t.Fatalf("la postulación quedó con escrituras parciales: cargo=%q consentimiento=%v", cargo, consentimiento)
	}
	if got := nombresFamiliares(t, pool, id); len(got) != 1 || got[0] != "Original" {
		t.Fatalf("familiares tras el ROLLBACK = %v, se esperaba [Original]", got)
	}
}

func TestGuardar_ROLLBACK_ErrorDeUnaSeccionConocida(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	if _, err := newService(pool).Obtener(ctx, id); err != nil {
		t.Fatal(err)
	}

	boom := errors.New("falla de la sección")
	svc := service.NewPostulacionService(
		postgres.NewPostulacionRepo(pool), registrocivil.NewMock(), postgres.NewTxManager(pool),
		postgres.NewFamiliaresRepo(pool), errSeccion{boom},
	)
	_, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: domain.CargoEmpleador, TieneFamiliares: true, Familiares: []domain.Familiar{familiar("X")}})
	if !errors.Is(err, boom) {
		t.Fatalf("se esperaba el error original sin envolver, llegó %v", err)
	}
	var cargo *string
	if err := pool.QueryRow(ctx, `SELECT cargo FROM postulaciones WHERE postulante_id = $1`, id).Scan(&cargo); err != nil {
		t.Fatal(err)
	}
	if cargo != nil || len(nombresFamiliares(t, pool, id)) != 0 {
		t.Fatalf("quedó algo escrito tras el ROLLBACK: cargo=%v", cargo)
	}
}

type errSeccion struct{ err error }

func (errSeccion) Cargar(context.Context, string, *domain.Formulario) error { return nil }
func (s errSeccion) Reemplazar(context.Context, domain.Tx, string, domain.Formulario) error {
	return s.err
}

func TestObtener_CreaBorradorConCodigoCorrelativo(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	primero := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	segundo := crearPostulante(t, pool, registrocivil.CedulaCasada)
	svc := newService(pool)

	p1, err := svc.Obtener(ctx, primero)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	p2, err := svc.Obtener(ctx, segundo)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	formato := regexp.MustCompile(`^BIESS-\d{4}-\d{6}$`)
	if !formato.MatchString(p1.Codigo) || !formato.MatchString(p2.Codigo) || p1.Codigo >= p2.Codigo {
		t.Fatalf("códigos %q y %q: se esperaba formato BIESS-AAAA-NNNNNN y crecientes", p1.Codigo, p2.Codigo)
	}
	if p1.Estado != domain.EstadoBorrador || p1.Personal.Nombres == "" || p1.Personal.Nacionalidad != domain.NacionalidadEcuatoriana {
		t.Fatalf("no se copió el Registro Civil: %+v", p1)
	}
	if p2.Personal.Pareja == nil {
		t.Fatalf("la persona casada debe traer los datos de la pareja")
	}

	// Leer de nuevo no crea otra ni cambia el código.
	otra, err := svc.Obtener(ctx, primero)
	if err != nil || otra.Codigo != p1.Codigo || otra.ID != p1.ID {
		t.Fatalf("segunda lectura: %+v, %v", otra, err)
	}
}

func TestObtener_PeticionesSimultaneasCreanUnaSolaPostulacion(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	const n = 8
	var wg sync.WaitGroup
	codigos := make([]string, n)
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := svc.Obtener(ctx, id)
			codigos[i], errs[i] = p.Codigo, err
		}()
	}
	wg.Wait()

	for i := range n {
		if errs[i] != nil {
			t.Fatalf("petición %d: %v", i, errs[i])
		}
		if codigos[i] != codigos[0] {
			t.Fatalf("códigos distintos para la misma postulación: %v", codigos)
		}
	}
	var total int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM postulaciones WHERE postulante_id = $1`, id).Scan(&total); err != nil || total != 1 {
		t.Fatalf("postulaciones = %d (%v), se esperaba 1", total, err)
	}
}

func TestGuardar_NoEditableFueraDeBorrador(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)
	if _, err := svc.Obtener(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE postulaciones SET estado = 'enviada' WHERE postulante_id = $1`, id); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: domain.CargoAfiliados})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("se esperaba 409, llegó %v", err)
	}
	var cargo *string
	if err := pool.QueryRow(ctx, `SELECT cargo FROM postulaciones WHERE postulante_id = $1`, id).Scan(&cargo); err != nil || cargo != nil {
		t.Fatalf("una postulación enviada no debe modificarse: cargo=%v err=%v", cargo, err)
	}
}

func TestGuardar_ExtranjeroCierraLaPostulacionYSeConfirma(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaExtranjera)
	svc := newService(pool)

	_, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: domain.CargoAfiliados})
	if !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("se esperaba 422, llegó %v", err)
	}
	var estado string
	var cargo *string
	if err := pool.QueryRow(ctx, `SELECT estado, cargo FROM postulaciones WHERE postulante_id = $1`, id).Scan(&estado, &cargo); err != nil {
		t.Fatal(err)
	}
	if estado != domain.EstadoCerrada || cargo != nil {
		t.Fatalf("estado=%q cargo=%v: debía quedar cerrada sin guardar el formulario", estado, cargo)
	}
}

func TestGuardar_ReferenciaInexistenteEsErrorDelCliente(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)

	inexistente := 999999
	_, err := svc.Guardar(ctx, id, domain.Formulario{Domicilio: domain.Domicilio{PaisID: &inexistente}})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba 400, llegó %v", err)
	}
}

func TestGuardar_DocumentoDeRespaldoDebeSerDeLaMismaPostulacion(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	idA := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	idB := crearPostulante(t, pool, registrocivil.CedulaUnionDeHecho)
	svc := newService(pool)
	pA, err := svc.Obtener(ctx, idA)
	if err != nil {
		t.Fatal(err)
	}
	pB, err := svc.Obtener(ctx, idB)
	if err != nil {
		t.Fatal(err)
	}
	documento := func(postulacionID, tipo string) string {
		var id string
		if err := pool.QueryRow(ctx, `
INSERT INTO documentos (postulacion_id, tipo, nombre, tamano_bytes, contenido)
VALUES ($1, $2, 'respaldo.pdf', 4, '%504446') RETURNING id::text`, postulacionID, tipo).Scan(&id); err != nil {
			t.Fatalf("creando el documento: %v", err)
		}
		return id
	}
	respaldoDe := func(postulanteID string) *string {
		var doc *string
		if err := pool.QueryRow(ctx,
			`SELECT respaldo_organizacion_doc_id::text FROM postulaciones WHERE postulante_id = $1`, postulanteID,
		).Scan(&doc); err != nil {
			t.Fatal(err)
		}
		return doc
	}
	guardar := func(doc string) error {
		_, err := svc.Guardar(ctx, idA, domain.Formulario{
			Cargo:                domain.CargoAfiliados,
			RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: "Asociación", DocumentoID: &doc},
		})
		return err
	}

	ajeno := documento(pB.ID, "organizacion")
	if err := guardar(ajeno); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("el documento de otro postulante debe responder 400, llegó %v", err)
	}
	if doc := respaldoDe(idA); doc != nil {
		t.Fatalf("la postulación de A quedó apuntando al documento %s de B", *doc)
	}

	if err := guardar(documento(pA.ID, "formacion")); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("un documento de otro tipo debe responder 400, llegó %v", err)
	}
	if err := guardar("no-es-un-uuid"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("un identificador mal formado debe responder 400, llegó %v", err)
	}

	propio := documento(pA.ID, "organizacion")
	if err := guardar(propio); err != nil {
		t.Fatalf("el documento propio debe aceptarse: %v", err)
	}
	if doc := respaldoDe(idA); doc == nil || *doc != propio {
		t.Fatalf("respaldo guardado = %v, se esperaba %s", doc, propio)
	}
}
