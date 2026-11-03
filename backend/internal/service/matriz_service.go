package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/xuri/excelize/v2"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

const (
	FormatoMatrizXLSX = "xlsx"
	FormatoMatrizPDF  = "pdf"
)

var cabecerasMatriz = []string{
	"Código", "Cédula", "Nombres", "Cargo", "Estado", "Elegible",
	"Nivel académico más alto", "Experiencia total (años)", "Capacitaciones", "Publicaciones",
}

// FiltroMatrizInput conserva los valores de la URL para validarlos en el servicio.
type FiltroMatrizInput struct {
	Formato  string
	Cargo    string
	Estado   string
	Elegible string
}

// ArchivoMatriz es el archivo generado y sus metadatos de descarga.
type ArchivoMatriz struct {
	Nombre      string
	ContentType string
	Contenido   []byte
}

// MatrizService genera el cuadro comparativo sin conocer HTTP ni SQL.
type MatrizService struct {
	repo  domain.MatrizPostulaciones
	ahora func() time.Time
}

// NewMatrizService crea el servicio con su repositorio inyectado.
func NewMatrizService(repo domain.MatrizPostulaciones) *MatrizService {
	return &MatrizService{repo: repo, ahora: time.Now}
}

// ConReloj reemplaza el reloj (para las pruebas).
func (s *MatrizService) ConReloj(ahora func() time.Time) *MatrizService {
	s.ahora = ahora
	return s
}

// Exportar valida los filtros, consulta las filas y genera el formato solicitado.
func (s *MatrizService) Exportar(ctx context.Context, in FiltroMatrizInput) (ArchivoMatriz, error) {
	formato := strings.ToLower(strings.TrimSpace(in.Formato))
	filtro, err := validarFiltroMatriz(in)
	if err != nil {
		return ArchivoMatriz{}, err
	}
	if formato != FormatoMatrizXLSX && formato != FormatoMatrizPDF {
		return ArchivoMatriz{}, domain.Validation("Formato de exportación inválido", map[string]string{
			"formato": "Debe ser xlsx o pdf",
		})
	}

	filas, err := s.repo.ListarMatriz(ctx, filtro)
	if err != nil {
		return ArchivoMatriz{}, err
	}
	// El servidor corre en UTC; la fecha del nombre y la del PDF van en hora de Ecuador.
	fecha := s.ahora().In(zonaEcuador)
	if formato == FormatoMatrizXLSX {
		contenido, err := generarMatrizXLSX(filas)
		if err != nil {
			return ArchivoMatriz{}, fmt.Errorf("generando la matriz Excel: %w", err)
		}
		return ArchivoMatriz{
			Nombre:      "matriz-postulantes-" + fecha.Format("20060102") + ".xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Contenido:   contenido,
		}, nil
	}

	contenido, err := generarMatrizPDF(filas, fecha)
	if err != nil {
		return ArchivoMatriz{}, fmt.Errorf("generando la matriz PDF: %w", err)
	}
	return ArchivoMatriz{
		Nombre:      "matriz-postulantes-" + fecha.Format("20060102") + ".pdf",
		ContentType: "application/pdf",
		Contenido:   contenido,
	}, nil
}

func validarFiltroMatriz(in FiltroMatrizInput) (domain.FiltroMatrizAdmin, error) {
	filtro := domain.FiltroMatrizAdmin{
		Cargo:  strings.ToLower(strings.TrimSpace(in.Cargo)),
		Estado: strings.ToLower(strings.TrimSpace(in.Estado)),
	}
	campos := map[string]string{}
	if filtro.Cargo != "" && !domain.CargoValido(filtro.Cargo) {
		campos["cargo"] = "Cargo no válido"
	}
	if filtro.Estado != "" && filtro.Estado != domain.EstadoBorrador &&
		filtro.Estado != domain.EstadoEnviada && filtro.Estado != domain.EstadoCerrada {
		campos["estado"] = "Estado no válido"
	}
	elegible := strings.ToLower(strings.TrimSpace(in.Elegible))
	if elegible != "" {
		if elegible != "true" && elegible != "false" {
			campos["elegible"] = "Debe ser true o false"
		} else {
			valor := elegible == "true"
			filtro.Elegible = &valor
		}
	}
	if len(campos) > 0 {
		return domain.FiltroMatrizAdmin{}, domain.Validation("Filtros inválidos", campos)
	}
	return filtro, nil
}

func generarMatrizXLSX(filas []domain.FilaMatrizAdmin) ([]byte, error) {
	libro := excelize.NewFile()
	defer func() { _ = libro.Close() }()
	const hoja = "Matriz"
	if err := libro.SetSheetName("Sheet1", hoja); err != nil {
		return nil, err
	}

	estiloCabecera, err := libro.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1B4965"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})
	if err != nil {
		return nil, err
	}
	estiloDatos, err := libro.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "bottom", Color: "D7E3EA", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}
	// Los años de experiencia se muestran con dos decimales (NumFmt 2 = "0.00") sin perder
	// el valor numérico, para que la columna se pueda ordenar y filtrar.
	estiloAnios, err := libro.NewStyle(&excelize.Style{
		NumFmt:    2,
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "D7E3EA", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	for columna, cabecera := range cabecerasMatriz {
		celda, _ := excelize.CoordinatesToCellName(columna+1, 1)
		if err := libro.SetCellValue(hoja, celda, cabecera); err != nil {
			return nil, err
		}
	}
	if err := libro.SetCellStyle(hoja, "A1", "J1", estiloCabecera); err != nil {
		return nil, err
	}
	if err := libro.SetRowHeight(hoja, 1, 32); err != nil {
		return nil, err
	}

	for i, fila := range filas {
		valores := valoresMatriz(fila)
		for columna, valor := range valores {
			celda, _ := excelize.CoordinatesToCellName(columna+1, i+2)
			if err := libro.SetCellValue(hoja, celda, valor); err != nil {
				return nil, err
			}
		}
	}
	ultimaFila := len(filas) + 1
	if ultimaFila > 1 {
		if err := libro.SetCellStyle(hoja, "A2", fmt.Sprintf("J%d", ultimaFila), estiloDatos); err != nil {
			return nil, err
		}
		if err := libro.SetCellStyle(hoja, "H2", fmt.Sprintf("H%d", ultimaFila), estiloAnios); err != nil {
			return nil, err
		}
	}
	if err := libro.AutoFilter(hoja, fmt.Sprintf("A1:J%d", ultimaFila), []excelize.AutoFilterOptions{}); err != nil {
		return nil, err
	}
	if err := libro.SetPanes(hoja, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		return nil, err
	}
	anchos := map[string]float64{"A": 22, "B": 15, "C": 32, "D": 16, "E": 14, "F": 12, "G": 27, "H": 23, "I": 17, "J": 16}
	for columna, ancho := range anchos {
		if err := libro.SetColWidth(hoja, columna, columna, ancho); err != nil {
			return nil, err
		}
	}
	buffer, err := libro.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// La tabla del PDF usa una grilla de 24 columnas para repartir el ancho de la hoja horizontal
// entre las diez columnas de la matriz (la suma de anchosMatrizPDF debe ser 24).
const gridMatrizPDF = 24

var anchosMatrizPDF = []int{3, 2, 4, 2, 2, 2, 3, 2, 2, 2}

// generarMatrizPDF arma la matriz en A4 horizontal con maroto, la misma librería del reporte de
// la evaluación: encabezado institucional con la fecha de generación, la cabecera de la tabla
// repetida en cada página y filas alternadas cuya altura crece si un nombre no cabe en una línea.
func generarMatrizPDF(filas []domain.FilaMatrizAdmin, fecha time.Time) ([]byte, error) {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Horizontal).
		WithMaxGridSize(gridMatrizPDF).
		WithLeftMargin(10).WithTopMargin(10).WithRightMargin(10).
		WithTitle("Matriz comparativa de postulantes", true).
		WithCreator("Sistema de postulación al Directorio del BIESS", true).
		WithPageNumber(props.PageNumber{Pattern: "Página {current} de {total}", Place: props.RightBottom, Size: 8, Color: colorGris}).
		Build()
	m := maroto.New(cfg)

	if err := m.RegisterHeader(encabezadoMatriz(fecha)...); err != nil {
		return nil, fmt.Errorf("armando el encabezado de la matriz: %w", err)
	}
	if len(filas) == 0 {
		m.AddRows(text.NewRow(12, "No hay postulantes con los filtros seleccionados.",
			props.Text{Top: 4, Size: 9, Style: fontstyle.Italic, Color: colorGris, Align: align.Center}))
	}
	for i, f := range filas {
		fila := filaMatrizPDF(valoresMatrizPDF(f), props.Text{Size: 8})
		if i%2 == 1 {
			fila.WithStyle(&props.Cell{BackgroundColor: colorFila})
		}
		m.AddRows(fila)
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return doc.GetBytes(), nil
}

func encabezadoMatriz(fecha time.Time) []core.Row {
	return []core.Row{
		row.New(20).WithStyle(&props.Cell{BackgroundColor: colorAzul}).Add(
			col.New(gridMatrizPDF).Add(
				text.New("BIESS · Concurso Público de Méritos y Oposición — Directorio", props.Text{Top: 4, Size: 9, Color: colorBlanco, Align: align.Center}),
				text.New("Matriz comparativa de postulantes", props.Text{Top: 10, Size: 14, Style: fontstyle.Bold, Color: colorBlanco, Align: align.Center}),
			),
		),
		text.NewRow(8, "Generado el "+fechaHora(fecha)+" (hora de Ecuador)", props.Text{Top: 2, Size: 8, Color: colorGris, Align: align.Right}),
		filaMatrizPDF(cabecerasMatriz, props.Text{Size: 7.5, Style: fontstyle.Bold, Color: colorBlanco}).
			WithStyle(&props.Cell{BackgroundColor: colorAzul}),
	}
}

// filaMatrizPDF reparte los valores en las columnas de la tabla. Los textos largos (nombres y
// nivel académico) van a la izquierda y el resto centrado.
func filaMatrizPDF(valores []string, base props.Text) core.Row {
	columnas := make([]core.Col, len(valores))
	for i, valor := range valores {
		p := base
		p.Top, p.Bottom, p.Left, p.Right = 1.5, 1.5, 1, 1
		p.Align = align.Center
		if i == 2 || i == 6 {
			p.Align = align.Left
		}
		columnas[i] = text.NewCol(anchosMatrizPDF[i], valor, p)
	}
	return row.New().Add(columnas...)
}

func valoresMatriz(f domain.FilaMatrizAdmin) []any {
	return []any{
		f.Codigo, f.Cedula, nombreCompletoMatriz(f), etiquetaCargoMatriz(f.Cargo),
		etiquetaEstadoMatriz(f.Estado), etiquetaSiNo(f.Elegible), etiquetaNivelMatriz(f.NivelAcademico),
		f.AniosExperiencia, f.Capacitaciones, f.Publicaciones,
	}
}

func valoresMatrizPDF(f domain.FilaMatrizAdmin) []string {
	return []string{
		f.Codigo, f.Cedula, nombreCompletoMatriz(f), etiquetaCargoMatriz(f.Cargo),
		etiquetaEstadoMatriz(f.Estado), etiquetaSiNo(f.Elegible), etiquetaNivelMatriz(f.NivelAcademico),
		fmt.Sprintf("%.2f", f.AniosExperiencia), fmt.Sprintf("%d", f.Capacitaciones), fmt.Sprintf("%d", f.Publicaciones),
	}
}

func nombreCompletoMatriz(f domain.FilaMatrizAdmin) string {
	return strings.TrimSpace(f.Nombres + " " + f.Apellidos)
}

func etiquetaCargoMatriz(cargo string) string {
	switch cargo {
	case domain.CargoAfiliados:
		return "Afiliados"
	case domain.CargoJubilados:
		return "Jubilados"
	case domain.CargoEmpleador:
		return "Empleador"
	default:
		return "Sin seleccionar"
	}
}

func etiquetaEstadoMatriz(estado string) string {
	switch estado {
	case domain.EstadoBorrador:
		return "Borrador"
	case domain.EstadoEnviada:
		return "Enviada"
	case domain.EstadoCerrada:
		return "Cerrada"
	default:
		return estado
	}
}

func etiquetaNivelMatriz(nivel string) string {
	switch nivel {
	case domain.NivelTercerNivel:
		return "Tercer nivel"
	case domain.NivelEspecialidad:
		return "Especialidad"
	case domain.NivelMaestria:
		return "Maestría"
	case domain.NivelPhD:
		return "PhD / Doctorado"
	default:
		return "Sin registro"
	}
}

func etiquetaSiNo(valor bool) string {
	if valor {
		return "Sí"
	}
	return "No"
}
