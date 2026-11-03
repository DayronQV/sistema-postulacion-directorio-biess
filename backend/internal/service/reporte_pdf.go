package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// EstadoRespuesta es cómo quedó una pregunta del examen al calificarlo.
type EstadoRespuesta int

const (
	SinResponder EstadoRespuesta = iota
	Correcta
	Incorrecta
)

// DetallePregunta es una fila del detalle del reporte. No lleva la respuesta correcta: el
// reporte no debe filtrar el banco a quien todavía esté rindiendo.
type DetallePregunta struct {
	Numero    int
	Enunciado string
	Estado    EstadoRespuesta
}

// DatosReporte es todo lo que muestra el reporte PDF de la evaluación.
type DatosReporte struct {
	Codigo       string
	Nombres      string
	Apellidos    string
	Cedula       string
	Cargo        string
	Inicio       time.Time
	Fin          time.Time
	Emitido      time.Time
	Nota         float64
	Aciertos     int
	Total        int
	Detalle      []DetallePregunta
	Verificacion string
}

// zonaEcuador es la hora continental de Ecuador (UTC-5, sin horario de verano). Se fija aquí
// porque la imagen del backend no trae la base de zonas horarias.
var zonaEcuador = time.FixedZone("ECT", -5*60*60)

const maxLargoEnunciadoReporte = 95

// Paleta del reporte.
var (
	colorAzul      = &props.Color{Red: 0, Green: 51, Blue: 102}
	colorAzulClaro = &props.Color{Red: 232, Green: 240, Blue: 248}
	colorFila      = &props.Color{Red: 246, Green: 248, Blue: 250}
	colorBlanco    = &props.Color{Red: 255, Green: 255, Blue: 255}
	colorGris      = &props.Color{Red: 110, Green: 110, Blue: 110}
	colorVerde     = &props.Color{Red: 22, Green: 128, Blue: 61}
	colorRojo      = &props.Color{Red: 185, Green: 28, Blue: 28}
)

// GenerarReportePDF arma el reporte del resultado de la evaluación teórica: encabezado,
// datos del postulante, nota destacada, detalle por pregunta (correcta, incorrecta o sin
// responder) y un código de verificación con QR.
func GenerarReportePDF(d DatosReporte) ([]byte, error) {
	cfg := config.NewBuilder().
		WithLeftMargin(15).WithTopMargin(12).WithRightMargin(15).
		WithTitle("Reporte de la evaluación teórica "+d.Codigo, true).
		WithCreator("Sistema de postulación al Directorio del BIESS", true).
		WithPageNumber(props.PageNumber{Pattern: "Página {current} de {total}", Place: props.RightBottom, Size: 8, Color: colorGris}).
		Build()
	m := maroto.New(cfg)

	if err := m.RegisterHeader(encabezadoReporte()...); err != nil {
		return nil, fmt.Errorf("armando el encabezado del reporte: %w", err)
	}
	m.AddRows(datosPostulante(d)...)
	m.AddRows(resultadoDestacado(d)...)
	m.AddRows(detallePorPregunta(d.Detalle)...)
	m.AddRows(verificacion(d)...)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("generando el reporte PDF: %w", err)
	}
	return doc.GetBytes(), nil
}

func encabezadoReporte() []core.Row {
	return []core.Row{
		row.New(22).WithStyle(&props.Cell{BackgroundColor: colorAzul}).Add(
			col.New(12).Add(
				text.New("BIESS · Concurso Público de Méritos y Oposición — Directorio", props.Text{Top: 4, Size: 9, Color: colorBlanco, Align: align.Center}),
				text.New("Reporte de la evaluación teórica", props.Text{Top: 11, Size: 15, Style: fontstyle.Bold, Color: colorBlanco, Align: align.Center}),
			),
		),
		row.New(6),
	}
}

func datosPostulante(d DatosReporte) []core.Row {
	dato := func(etiqueta, valor string) core.Row {
		return row.New(7).Add(
			text.NewCol(4, etiqueta, props.Text{Size: 9, Style: fontstyle.Bold}),
			text.NewCol(8, valor, props.Text{Size: 9}),
		)
	}
	filas := []core.Row{
		text.NewRow(7, "Datos del postulante", props.Text{Size: 11, Style: fontstyle.Bold, Color: colorAzul}),
		line.NewRow(2, props.Line{Color: colorAzul, Thickness: 0.4}),
		dato("Nombres y apellidos:", strings.TrimSpace(d.Nombres+" "+d.Apellidos)),
		dato("Cédula:", d.Cedula),
		dato("Código de postulación:", d.Codigo),
	}
	if d.Cargo != "" {
		filas = append(filas, dato("Cargo:", nombreCargo(d.Cargo)))
	}
	filas = append(filas,
		dato("Inicio y fin:", fechaHora(d.Inicio)+" – "+fechaHora(d.Fin)+" (hora de Ecuador)"),
		row.New(6),
	)
	return filas
}

func resultadoDestacado(d DatosReporte) []core.Row {
	return []core.Row{
		row.New(28).WithStyle(&props.Cell{BackgroundColor: colorAzulClaro}).Add(
			col.New(6).Add(
				text.New("Nota obtenida", props.Text{Top: 5, Size: 10, Color: colorGris, Align: align.Center}),
				text.New(formatoNota(d.Nota)+" / 100", props.Text{Top: 12, Size: 22, Style: fontstyle.Bold, Color: colorAzul, Align: align.Center}),
			),
			col.New(6).Add(
				text.New("Respuestas correctas", props.Text{Top: 5, Size: 10, Color: colorGris, Align: align.Center}),
				text.New(fmt.Sprintf("%d de %d", d.Aciertos, d.Total), props.Text{Top: 12, Size: 22, Style: fontstyle.Bold, Color: colorAzul, Align: align.Center}),
			),
		),
		row.New(6),
	}
}

func detallePorPregunta(detalle []DetallePregunta) []core.Row {
	filas := []core.Row{
		text.NewRow(7, "Detalle por pregunta", props.Text{Size: 11, Style: fontstyle.Bold, Color: colorAzul}),
		row.New(8).WithStyle(&props.Cell{BackgroundColor: colorAzul}).Add(
			text.NewCol(1, "N.º", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Color: colorBlanco, Align: align.Center}),
			text.NewCol(9, "Pregunta", props.Text{Top: 2, Left: 2, Size: 9, Style: fontstyle.Bold, Color: colorBlanco}),
			text.NewCol(2, "Resultado", props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Color: colorBlanco, Align: align.Center}),
		),
	}
	for i, p := range detalle {
		etiqueta, color := "Sin responder", colorGris
		switch p.Estado {
		case Correcta:
			etiqueta, color = "Correcta", colorVerde
		case Incorrecta:
			etiqueta, color = "Incorrecta", colorRojo
		}
		fila := row.New(8).Add(
			text.NewCol(1, fmt.Sprint(p.Numero), props.Text{Top: 2, Size: 9, Align: align.Center}),
			text.NewCol(9, recortar(p.Enunciado, maxLargoEnunciadoReporte), props.Text{Top: 2, Left: 2, Size: 9}),
			text.NewCol(2, etiqueta, props.Text{Top: 2, Size: 9, Style: fontstyle.Bold, Color: color, Align: align.Center}),
		)
		if i%2 == 1 {
			fila.WithStyle(&props.Cell{BackgroundColor: colorFila})
		}
		filas = append(filas, fila)
	}
	return filas
}

func verificacion(d DatosReporte) []core.Row {
	return []core.Row{
		row.New(8),
		row.New(26).Add(
			col.New(9).Add(
				text.New("Código de verificación: "+d.Verificacion, props.Text{Top: 6, Size: 9, Style: fontstyle.Bold}),
				text.New("Emitido el "+fechaHora(d.Emitido)+" (hora de Ecuador). Documento generado automáticamente por el sistema de postulación.", props.Text{Top: 12, Size: 8, Color: colorGris}),
			),
			code.NewQrCol(3, "BIESS|"+d.Codigo+"|"+d.Verificacion+"|"+formatoNota(d.Nota), props.Rect{Center: true, Percent: 90}),
		),
	}
}

// fechaHora muestra la hora en Ecuador, por ejemplo "20/11/2026 10:00".
func fechaHora(t time.Time) string { return t.In(zonaEcuador).Format("02/01/2006 15:04") }

// formatoNota muestra la nota con dos decimales y coma decimal, por ejemplo "66,67".
func formatoNota(n float64) string { return strings.Replace(fmt.Sprintf("%.2f", n), ".", ",", 1) }

// recortar acorta un texto a max caracteres (no bytes) y agrega puntos suspensivos.
func recortar(s string, max int) string {
	r := []rune(strings.Join(strings.Fields(s), " "))
	if len(r) <= max {
		return string(r)
	}
	return strings.TrimSpace(string(r[:max-1])) + "…"
}

func nombreCargo(cargo string) string {
	switch cargo {
	case domain.CargoAfiliados:
		return "Representante de los afiliados"
	case domain.CargoJubilados:
		return "Representante de los jubilados"
	case domain.CargoEmpleador:
		return "Representante de los empleadores"
	default:
		return cargo
	}
}

// codigoVerificacion deriva un código corto y legible del id del intento, por ejemplo
// "4F7A-91C2-BB30", para identificar el reporte.
func codigoVerificacion(intentoID string) string {
	hex := strings.ToUpper(strings.ReplaceAll(intentoID, "-", ""))
	if len(hex) < 12 {
		return hex
	}
	return hex[0:4] + "-" + hex[4:8] + "-" + hex[8:12]
}

// nombreReporte es el nombre del archivo del reporte.
func nombreReporte(codigo string) string { return "reporte-evaluacion-" + codigo + ".pdf" }
