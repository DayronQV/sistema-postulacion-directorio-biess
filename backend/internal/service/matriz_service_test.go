package service_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

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
	repo := &fakeMatrizRepo{filas: []domain.FilaMatrizAdmin{filaMatrizPrueba()}}
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
