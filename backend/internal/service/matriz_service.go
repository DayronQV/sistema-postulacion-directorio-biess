package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
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
	fecha := s.ahora()
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

func generarMatrizPDF(filas []domain.FilaMatrizAdmin, fecha time.Time) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(true, 11)
	pdf.SetTitle("Matriz comparativa de postulantes", true)
	pdf.SetAuthor("BIESS", true)
	pdf.AliasNbPages("")
	traducir := pdf.UnicodeTranslatorFromDescriptor("")
	anchos := []float64{24, 24, 44, 22, 18, 17, 42, 30, 30, 30}

	pdf.SetHeaderFunc(func() {
		pdf.SetTextColor(27, 73, 101)
		pdf.SetFont("Arial", "B", 13)
		pdf.CellFormat(0, 7, traducir("BIESS - Matriz comparativa de postulantes"), "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(70, 70, 70)
		pdf.CellFormat(0, 5, traducir("Generado el "+fecha.Format("02/01/2006 15:04")), "", 1, "R", false, 0, "")
		pdf.Ln(1)
		filaPDF(pdf, cabecerasMatriz, anchos, traducir, true)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-8)
		pdf.SetFont("Arial", "", 7)
		pdf.SetTextColor(90, 90, 90)
		pdf.CellFormat(0, 5, fmt.Sprintf("Página %d/{nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	if len(filas) == 0 {
		pdf.SetFont("Arial", "I", 9)
		pdf.SetTextColor(70, 70, 70)
		pdf.CellFormat(0, 10, traducir("No hay postulantes con los filtros seleccionados."), "", 1, "C", false, 0, "")
	} else {
		for _, fila := range filas {
			valores := valoresMatrizPDF(fila)
			filaPDF(pdf, valores, anchos, traducir, false)
		}
	}

	var buffer bytes.Buffer
	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func filaPDF(pdf *gofpdf.Fpdf, valores []string, anchos []float64, traducir func(string) string, cabecera bool) {
	si := 7.0
	if cabecera {
		pdf.SetFillColor(27, 73, 101)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 6.5)
	} else {
		pdf.SetFillColor(245, 248, 250)
		pdf.SetTextColor(35, 35, 35)
		pdf.SetFont("Arial", "", 6.5)
	}
	for i, valor := range valores {
		pdf.CellFormat(anchos[i], si, traducir(limitarTexto(valor, 34)), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(si)
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

func limitarTexto(valor string, maximo int) string {
	runas := []rune(valor)
	if len(runas) <= maximo {
		return valor
	}
	return string(runas[:maximo-1]) + "…"
}
