package service_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

type fakeMatrizRepo struct {
	filtro    domain.FiltroMatrizAdmin
	filas     []domain.FilaMatrizAdmin
	err       error
	consultas int
}

func (f *fakeMatrizRepo) ListarMatriz(_ context.Context, filtro domain.FiltroMatrizAdmin) ([]domain.FilaMatrizAdmin, error) {
	f.consultas++
	f.filtro = filtro
	return f.filas, f.err
}

func filaMatrizPrueba() domain.FilaMatrizAdmin {
	return domain.FilaMatrizAdmin{
		Codigo: "BIESS-2026-000001", Cedula: "0912345678", Nombres: "Ana María", Apellidos: "Pérez",
		Cargo: domain.CargoAfiliados, Estado: domain.EstadoEnviada, Elegible: true,
		NivelAcademico: domain.NivelMaestria, AniosExperiencia: 5.25, Capacitaciones: 3, Publicaciones: 2,
	}
}

func TestMatrizExportar_XLSXSeAbreYRespetaFiltros(t *testing.T) {
	fila := filaMatrizPrueba()
	// Con más de dos decimales en el dato, solo se distingue si la celda queda realmente en
	// formato "0.00" de si queda en el formato general del Excel (que no redondearía).
	fila.AniosExperiencia = 5.2567
	repo := &fakeMatrizRepo{filas: []domain.FilaMatrizAdmin{fila}}
	archivo, err := service.NewMatrizService(repo).Exportar(context.Background(), service.FiltroMatrizInput{
		Formato: " XLSX ", Cargo: " AFILIADOS ", Estado: " ENVIADA ", Elegible: " TRUE ",
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if repo.filtro.Cargo != domain.CargoAfiliados || repo.filtro.Estado != domain.EstadoEnviada ||
		repo.filtro.Elegible == nil || !*repo.filtro.Elegible {
		t.Fatalf("filtro = %+v", repo.filtro)
	}
	if !strings.HasSuffix(archivo.Nombre, ".xlsx") || archivo.ContentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("archivo = %+v", archivo)
	}

	libro, err := excelize.OpenReader(bytes.NewReader(archivo.Contenido))
	if err != nil {
		t.Fatalf("el Excel generado no se puede abrir: %v", err)
	}
	defer func() { _ = libro.Close() }()
	filas, err := libro.GetRows("Matriz")
	if err != nil || len(filas) != 2 {
		t.Fatalf("filas = %v, err = %v", filas, err)
	}
	if filas[0][0] != "Código" || filas[1][0] != "BIESS-2026-000001" ||
		filas[1][2] != "Ana María Pérez" || filas[1][6] != "Maestría" || filas[1][8] != "3" {
		t.Fatalf("contenido = %v", filas)
	}
	// Los años de experiencia se muestran redondeados a dos decimales (formato "0.00" de la
	// celda), pero el valor sigue siendo numérico para poder ordenar y filtrar la columna.
	if filas[1][7] != "5.26" {
		t.Fatalf("años de experiencia = %q, se esperaba 5.26 (redondeado)", filas[1][7])
	}
	valor, err := libro.GetCellValue("Matriz", "H2", excelize.Options{RawCellValue: true})
	if err != nil || valor != "5.2567" {
		t.Fatalf("el valor crudo de la celda debe conservar los decimales originales: %q, err = %v", valor, err)
	}
}

// El reporte se genera en el servidor en UTC; el nombre del archivo debe usar la hora de
// Ecuador, igual que el reporte PDF de la evaluación.
func TestMatrizExportar_NombreUsaLaFechaDeEcuador(t *testing.T) {
	repo := &fakeMatrizRepo{filas: []domain.FilaMatrizAdmin{filaMatrizPrueba()}}
	// 03:00 UTC del día 21 son las 22:00 del día 20 en Ecuador (UTC-5).
	medianocheUTC := time.Date(2026, 11, 21, 3, 0, 0, 0, time.UTC)
	servicio := service.NewMatrizService(repo).ConReloj(func() time.Time { return medianocheUTC })

	archivo, err := servicio.Exportar(context.Background(), service.FiltroMatrizInput{Formato: "xlsx"})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if archivo.Nombre != "matriz-postulantes-20261120.xlsx" {
		t.Fatalf("nombre = %q, se esperaba la fecha en Ecuador (20261120)", archivo.Nombre)
	}
}

func TestMatrizExportar_PDFValido(t *testing.T) {
	repo := &fakeMatrizRepo{filas: []domain.FilaMatrizAdmin{filaMatrizPrueba()}}
	archivo, err := service.NewMatrizService(repo).Exportar(context.Background(), service.FiltroMatrizInput{Formato: "pdf"})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !strings.HasSuffix(archivo.Nombre, ".pdf") || archivo.ContentType != "application/pdf" ||
		!bytes.HasPrefix(archivo.Contenido, []byte("%PDF-")) || !bytes.Contains(archivo.Contenido, []byte("%%EOF")) {
		t.Fatalf("PDF inválido: nombre=%q tipo=%q bytes=%d", archivo.Nombre, archivo.ContentType, len(archivo.Contenido))
	}
}

// Sin postulantes que cumplan los filtros, el PDF debe seguir siendo válido y distinto
// del de una lista con datos (muestra el aviso en vez de una tabla vacía).
func TestMatrizExportar_PDFSinFilasEsValidoYDistinto(t *testing.T) {
	conFilas, err := service.NewMatrizService(&fakeMatrizRepo{filas: []domain.FilaMatrizAdmin{filaMatrizPrueba()}}).
		Exportar(context.Background(), service.FiltroMatrizInput{Formato: "pdf"})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	sinFilas, err := service.NewMatrizService(&fakeMatrizRepo{}).Exportar(context.Background(), service.FiltroMatrizInput{Formato: "pdf"})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !bytes.HasPrefix(sinFilas.Contenido, []byte("%PDF-")) || !bytes.Contains(sinFilas.Contenido, []byte("%%EOF")) {
		t.Fatalf("PDF sin filas inválido: bytes=%d", len(sinFilas.Contenido))
	}
	if len(sinFilas.Contenido) >= len(conFilas.Contenido) {
		t.Fatalf("el PDF sin filas (%d bytes) debería ser más chico que el de una fila con datos (%d bytes)",
			len(sinFilas.Contenido), len(conFilas.Contenido))
	}
}

func TestMatrizExportar_RechazaFormatoYFiltrosInvalidos(t *testing.T) {
	casos := []service.FiltroMatrizInput{
		{}, {Formato: "csv"}, {Formato: "pdf", Cargo: "otro"},
		{Formato: "xlsx", Estado: "otro"}, {Formato: "pdf", Elegible: "si"},
	}
	for _, in := range casos {
		repo := &fakeMatrizRepo{}
		_, err := service.NewMatrizService(repo).Exportar(context.Background(), in)
		if !errors.Is(err, domain.ErrValidation) || repo.consultas != 0 {
			t.Fatalf("entrada %+v: consultas=%d, err=%v", in, repo.consultas, err)
		}
	}
}

func TestMatrizExportar_PropagaErrorDelRepositorio(t *testing.T) {
	esperado := errors.New("base no disponible")
	repo := &fakeMatrizRepo{err: esperado}
	_, err := service.NewMatrizService(repo).Exportar(context.Background(), service.FiltroMatrizInput{Formato: "xlsx"})
	if !errors.Is(err, esperado) {
		t.Fatalf("se esperaba %v, llegó %v", esperado, err)
	}
}
