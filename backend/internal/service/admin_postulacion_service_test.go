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
	// consultas cuenta las llamadas al repositorio: un id que no es UUID no debe llegar a él.
	consultas int
}

func (f *fakeMeritos) EstadoPorID(_ context.Context, id string) (string, error) {
	f.consultas++
	e, ok := f.estados[id]
	if !ok {
		return "", domain.NotFound("La postulación no existe")
	}
	return e, nil
}

func (f *fakeMeritos) ActualizarMerito(_ context.Context, id string, supero bool) (bool, error) {
	f.consultas++
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

func TestMarcarMerito_PostulacionInexistenteEs404(t *testing.T) {
	svc, _ := nuevoMeritos(domain.EstadoEnviada)
	id := "99999999-2222-3333-4444-555555555555"
	if _, err := svc.MarcarMerito(context.Background(), id, true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("id %q: se esperaba 404, llegó %v", id, err)
	}
}

// Con PostgreSQL real, un id que no es UUID haría fallar la consulta (500): el servicio lo
// responde como 404 sin llegar al repositorio. El fake lo registraría como enviada, así
// que sin la validación esta prueba falla.
func TestMarcarMerito_IdQueNoEsUUIDEs404SinConsultarElRepositorio(t *testing.T) {
	for _, id := range []string{"no-es-un-uuid", "", "   ", "11111111-2222-3333-4444-55555555555", "11111111-2222-3333-4444-5555555555555"} {
		svc, f := nuevoMeritos(domain.EstadoEnviada)
		f.estados[id] = domain.EstadoEnviada
		if _, err := svc.MarcarMerito(context.Background(), id, true); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("id %q: se esperaba 404, llegó %v", id, err)
		}
		if f.consultas != 0 || f.meritos[id] {
			t.Fatalf("id %q: no debía llegar al repositorio (consultas = %d)", id, f.consultas)
		}
	}
}

type fakePostulacionesAdmin struct {
	filtro    domain.FiltroPostulacionesAdmin
	items     []domain.ResumenPostulacionAdmin
	total     int
	detalle   domain.Postulacion
	err       error
	consultas int
}

func (f *fakePostulacionesAdmin) ListarAdmin(_ context.Context, filtro domain.FiltroPostulacionesAdmin) ([]domain.ResumenPostulacionAdmin, int, error) {
	f.consultas++
	f.filtro = filtro
	return f.items, f.total, f.err
}

func (f *fakePostulacionesAdmin) ObtenerPorIDAdmin(_ context.Context, _ string) (domain.Postulacion, error) {
	f.consultas++
	return f.detalle, f.err
}

type fakeDocumentosAdmin struct {
	postulacionID string
	documentoID   string
	documento     domain.Documento
	err           error
}

func (f *fakeDocumentosAdmin) ObtenerDePostulacion(_ context.Context, postulacionID, documentoID string) (domain.Documento, error) {
	f.postulacionID, f.documentoID = postulacionID, documentoID
	return f.documento, f.err
}

type fakeSeccionAdmin struct {
	cargar func(*domain.Formulario)
	err    error
}

func (f fakeSeccionAdmin) Cargar(_ context.Context, _ string, formulario *domain.Formulario) error {
	if f.cargar != nil {
		f.cargar(formulario)
	}
	return f.err
}

func (fakeSeccionAdmin) Reemplazar(context.Context, domain.Tx, string, domain.Formulario) error {
	return nil
}

func nuevoAdminConsultas(repo *fakePostulacionesAdmin, documentos *fakeDocumentosAdmin, secciones ...domain.SeccionPostulacion) *service.AdminPostulacionService {
	return service.NewAdminPostulacionService(&fakeMeritos{}).ConConsultas(repo, documentos, secciones...)
}

func TestListarPostulaciones_NormalizaFiltrosYPagina(t *testing.T) {
	repo := &fakePostulacionesAdmin{
		items: []domain.ResumenPostulacionAdmin{{ID: idPostulacion, Codigo: "BIESS-2026-000001"}},
		total: 21,
	}
	r, err := nuevoAdminConsultas(repo, &fakeDocumentosAdmin{}).Listar(context.Background(), service.FiltroPostulacionesInput{
		Cargo: " AFILIADOS ", Estado: " ENVIADA ", Elegible: " TRUE ", Pagina: 2,
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if repo.filtro.Cargo != domain.CargoAfiliados || repo.filtro.Estado != domain.EstadoEnviada ||
		repo.filtro.Elegible == nil || !*repo.filtro.Elegible || repo.filtro.Pagina != 2 || repo.filtro.PorPagina != 20 {
		t.Fatalf("filtro = %+v", repo.filtro)
	}
	if r.Total != 21 || r.Pagina != 2 || r.PorPagina != 20 || len(r.Items) != 1 {
		t.Fatalf("página = %+v", r)
	}
}

func TestListarPostulaciones_PaginaPorDefectoYFiltrosInvalidos(t *testing.T) {
	repo := &fakePostulacionesAdmin{}
	svc := nuevoAdminConsultas(repo, &fakeDocumentosAdmin{})
	if _, err := svc.Listar(context.Background(), service.FiltroPostulacionesInput{}); err != nil || repo.filtro.Pagina != 1 {
		t.Fatalf("página por defecto = %d, err = %v", repo.filtro.Pagina, err)
	}
	casos := []service.FiltroPostulacionesInput{
		{Cargo: "otro"}, {Estado: "otro"}, {Elegible: "sí"}, {Pagina: -1},
	}
	for _, in := range casos {
		antes := repo.consultas
		if _, err := svc.Listar(context.Background(), in); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("entrada %+v: se esperaba 400, llegó %v", in, err)
		}
		if repo.consultas != antes {
			t.Fatalf("entrada %+v no debía consultar el repositorio", in)
		}
	}
}

func TestObtenerPostulacionAdmin_CargaSeccionesYCalculaExperiencia(t *testing.T) {
	repo := &fakePostulacionesAdmin{detalle: domain.Postulacion{ID: idPostulacion}}
	seccion := fakeSeccionAdmin{cargar: func(f *domain.Formulario) {
		f.Experiencia = []domain.ExperienciaProfesional{{FechaInicio: "2020-01-01", FechaFin: "2021-02-03"}}
	}}
	p, err := nuevoAdminConsultas(repo, &fakeDocumentosAdmin{}, seccion).Obtener(context.Background(), idPostulacion)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(p.Formulario.Experiencia) != 1 || p.Formulario.Experiencia[0].Duracion.Anios != 1 ||
		p.Formulario.Experiencia[0].Duracion.Meses != 1 || p.Formulario.Experiencia[0].Duracion.Dias != 2 ||
		p.Formulario.DuracionExperiencia.Anios != 1 {
		t.Fatalf("duraciones = %+v / %+v", p.Formulario.Experiencia[0].Duracion, p.Formulario.DuracionExperiencia)
	}
}

func TestObtenerPostulacionAdmin_IdInvalidoNoConsulta(t *testing.T) {
	repo := &fakePostulacionesAdmin{}
	_, err := nuevoAdminConsultas(repo, &fakeDocumentosAdmin{}).Obtener(context.Background(), "no-es-uuid")
	if !errors.Is(err, domain.ErrNotFound) || repo.consultas != 0 {
		t.Fatalf("consultas = %d, err = %v", repo.consultas, err)
	}
}

func TestDescargarDocumentoAdmin_ExigeAmbosIdsYConservaAislamiento(t *testing.T) {
	documentoID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	documentos := &fakeDocumentosAdmin{documento: domain.Documento{ID: documentoID, Nombre: "titulo.pdf"}}
	svc := nuevoAdminConsultas(&fakePostulacionesAdmin{}, documentos)
	doc, err := svc.DescargarDocumento(context.Background(), idPostulacion, documentoID)
	if err != nil || doc.ID != documentoID || documentos.postulacionID != idPostulacion || documentos.documentoID != documentoID {
		t.Fatalf("documento = %+v, ids = %s/%s, err = %v", doc, documentos.postulacionID, documentos.documentoID, err)
	}
	if _, err := svc.DescargarDocumento(context.Background(), idPostulacion, "malo"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("se esperaba 404, llegó %v", err)
	}
}
