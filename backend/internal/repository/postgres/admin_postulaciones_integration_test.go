package postgres_test

import (
	"context"
	"math"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
)

func TestPostulacionesAdmin_FiltrosPaginacionYDetalle(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewPostulacionRepo(pool)
	ctx := context.Background()

	idA := postulacionConEstado(t, pool, "1799999910", domain.EstadoEnviada)
	idB := postulacionConEstado(t, pool, "1799999911", domain.EstadoEnviada)
	idC := postulacionConEstado(t, pool, "1799999912", domain.EstadoBorrador)
	if _, err := pool.Exec(ctx,
		`UPDATE postulaciones SET cargo = 'afiliados', elegible = true WHERE id IN ($1, $2)`, idA, idB,
	); err != nil {
		t.Fatalf("preparando filtros: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`UPDATE postulaciones SET cargo = 'jubilados', elegible = false WHERE id = $1`, idC,
	); err != nil {
		t.Fatalf("preparando filtros: %v", err)
	}

	elegible := true
	items, total, err := repo.ListarAdmin(ctx, domain.FiltroPostulacionesAdmin{
		Cargo: domain.CargoAfiliados, Estado: domain.EstadoEnviada, Elegible: &elegible,
		Pagina: 1, PorPagina: 1,
	})
	if err != nil || total != 2 || len(items) != 1 || items[0].Cargo != domain.CargoAfiliados || !items[0].Elegible {
		t.Fatalf("items = %+v, total = %d, err = %v", items, total, err)
	}
	segunda, total, err := repo.ListarAdmin(ctx, domain.FiltroPostulacionesAdmin{
		Cargo: domain.CargoAfiliados, Estado: domain.EstadoEnviada, Elegible: &elegible,
		Pagina: 2, PorPagina: 1,
	})
	if err != nil || total != 2 || len(segunda) != 1 || segunda[0].ID == items[0].ID {
		t.Fatalf("segunda página = %+v, total = %d, err = %v", segunda, total, err)
	}

	detalle, err := repo.ObtenerPorIDAdmin(ctx, idC)
	if err != nil || detalle.ID != idC || detalle.Estado != domain.EstadoBorrador || detalle.Formulario.Cargo != domain.CargoJubilados {
		t.Fatalf("detalle = %+v, err = %v", detalle, err)
	}
}

func TestPostulacionesAdmin_MatrizReuneIndicadoresYRespetaFiltros(t *testing.T) {
	pool := newPool(t)
	repo := postgres.NewPostulacionRepo(pool)
	ctx := context.Background()

	id := postulacionConEstado(t, pool, "1799999913", domain.EstadoEnviada)
	if _, err := pool.Exec(ctx, `UPDATE postulaciones SET cargo = 'afiliados', elegible = true WHERE id = $1`, id); err != nil {
		t.Fatalf("preparando la postulación: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO formacion_academica (postulacion_id, nivel, institucion, tipo, titulo, area, senescyt_numero, senescyt_fecha)
VALUES ($1, 'tercer_nivel', 'UPSE', 'nacional', 'Economista', 'economia', '1', '2020-01-01'),
       ($1, 'maestria', 'UPSE', 'nacional', 'Magíster', 'finanzas', '2', '2022-01-01')`, id); err != nil {
		t.Fatalf("creando formación: %v", err)
	}
	var documentoID string
	if err := pool.QueryRow(ctx, `
INSERT INTO documentos (postulacion_id, tipo, nombre, tamano_bytes, contenido)
VALUES ($1, 'experiencia', 'experiencia.pdf', 4, decode('25504446', 'hex'))
RETURNING id::text`, id).Scan(&documentoID); err != nil {
		t.Fatalf("creando documento: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO experiencia_profesional
    (postulacion_id, tipo, cargo, institucion, fecha_inicio, fecha_fin, actividades, descripcion, documento_id)
VALUES ($1, 'general', 'Analista', 'Empresa', '2020-01-01', '2021-02-03', 'finanzas', 'Funciones', $2)`, id, documentoID); err != nil {
		t.Fatalf("creando experiencia: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO capacitaciones (postulacion_id, tipo, nombre, institucion, fecha_inicio, fecha_fin, horas)
VALUES ($1, 'curso', 'Riesgos', 'UPSE', '2024-01-01', '2024-01-02', 8),
	       ($1, 'taller', 'Finanzas', 'UPSE', '2024-02-01', '2024-02-02', 8)`, id); err != nil {
		t.Fatalf("creando capacitaciones: %v", err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO publicaciones (postulacion_id, titulo, tipo, medio, fecha, relacionada)
VALUES ($1, 'Artículo', 'publicacion', 'Revista', '2024-01-01', true)`, id); err != nil {
		t.Fatalf("creando publicaciones: %v", err)
	}

	elegible := true
	filas, err := repo.ListarMatriz(ctx, domain.FiltroMatrizAdmin{
		Cargo: domain.CargoAfiliados, Estado: domain.EstadoEnviada, Elegible: &elegible,
	})
	if err != nil || len(filas) != 1 {
		t.Fatalf("filas = %+v, err = %v", filas, err)
	}
	fila := filas[0]
	if fila.NivelAcademico != domain.NivelMaestria || fila.Capacitaciones != 2 || fila.Publicaciones != 1 ||
		math.Abs(fila.AniosExperiencia-(392.0/360.0)) > 0.001 {
		t.Fatalf("indicadores = %+v", fila)
	}
}
