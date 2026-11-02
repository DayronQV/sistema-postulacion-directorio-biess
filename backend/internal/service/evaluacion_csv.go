package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Columnas del CSV de preguntas. Las opciones van en una sola celda separadas por "|" y
// respuesta_correcta es el índice de la opción correcta, desde 0:
//
//	enunciado,opciones,respuesta_correcta
//	"¿Cuál es la capital de Ecuador?",Guayaquil|Quito|Cuenca,1
const (
	colEnunciado = "enunciado"
	colOpciones  = "opciones"
	colCorrecta  = "respuesta_correcta"

	separadorOpciones = "|"
)

// filaPregunta es una fila del CSV con su número de línea en el archivo, para que un error
// se pueda ubicar (la línea 1 es la cabecera).
type filaPregunta struct {
	linea int
	input PreguntaInput
}

// leerPreguntasCSV convierte el archivo en filas y agrega al reporte lo que no se pueda
// leer. Acepta BOM de UTF-8 y, como Excel en español, punto y coma como separador de
// columnas.
func leerPreguntasCSV(r io.Reader, rep *reporte) []filaPregunta {
	contenido, err := io.ReadAll(r)
	if err != nil {
		rep.agregar("archivo", "No se pudo leer el archivo")
		return nil
	}
	contenido = bytes.TrimPrefix(contenido, []byte("\xEF\xBB\xBF"))
	if len(bytes.TrimSpace(contenido)) == 0 {
		rep.agregar("archivo", "El archivo está vacío")
		return nil
	}

	lector := csv.NewReader(bytes.NewReader(contenido))
	lector.Comma = separadorDeColumnas(contenido)
	lector.FieldsPerRecord = -1
	lector.TrimLeadingSpace = true

	cabecera, err := lector.Read()
	if err != nil {
		rep.agregar("archivo", "No se pudo leer la cabecera del archivo")
		return nil
	}
	indice := map[string]int{}
	for i, nombre := range cabecera {
		indice[strings.ToLower(strings.TrimSpace(nombre))] = i
	}
	for _, col := range []string{colEnunciado, colOpciones, colCorrecta} {
		if _, ok := indice[col]; !ok {
			rep.agregar("archivo", fmt.Sprintf("Falta la columna %q: la cabecera debe ser %s,%s,%s", col, colEnunciado, colOpciones, colCorrecta))
			return nil
		}
	}

	var filas []filaPregunta
	for {
		registro, err := lector.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			var pe *csv.ParseError
			if errors.As(err, &pe) {
				rep.agregar(fmt.Sprintf("filas[%d]", pe.Line), "Formato de CSV no válido (revise las comillas)")
			} else {
				rep.agregar("archivo", "No se pudo leer el archivo")
			}
			return nil
		}
		linea, _ := lector.FieldPos(0)
		if len(filas) >= domain.MaxPreguntasPorCarga {
			rep.agregar("archivo", fmt.Sprintf("Máximo %d preguntas por carga", domain.MaxPreguntasPorCarga))
			return nil
		}
		filas = append(filas, filaDesdeRegistro(registro, indice, linea, rep))
	}
	if len(filas) == 0 {
		rep.agregar("archivo", "El archivo no tiene preguntas debajo de la cabecera")
	}
	return filas
}

// filaDesdeRegistro arma la pregunta de una fila; lo que no sea un número entero o falte
// se reporta con la línea, y la validación general del servicio revisa el resto.
func filaDesdeRegistro(registro []string, indice map[string]int, linea int, rep *reporte) filaPregunta {
	celda := func(col string) string {
		if i := indice[col]; i < len(registro) {
			return registro[i]
		}
		return ""
	}
	in := PreguntaInput{Enunciado: celda(colEnunciado)}
	if texto := strings.TrimSpace(celda(colOpciones)); texto != "" {
		in.Opciones = strings.Split(texto, separadorOpciones)
	}
	switch texto := strings.TrimSpace(celda(colCorrecta)); {
	case texto == "":
		// Queda nil: la validación lo reporta como campo obligatorio.
	default:
		n, err := strconv.Atoi(texto)
		if err != nil {
			rep.agregar(fmt.Sprintf("filas[%d].respuesta_correcta", linea), "Debe ser un número entero (índice de la opción correcta, desde 0)")
			// Se deja el campo en 0 para que la validación no lo reporte dos veces.
			cero := 0
			in.RespuestaCorrecta = &cero
		} else {
			in.RespuestaCorrecta = &n
		}
	}
	return filaPregunta{linea: linea, input: in}
}

// separadorDeColumnas elige ";" si la cabecera lo usa y no usa ",", como hace Excel con la
// configuración regional en español; en cualquier otro caso, ",".
func separadorDeColumnas(contenido []byte) rune {
	primera := contenido
	if i := bytes.IndexByte(contenido, '\n'); i >= 0 {
		primera = contenido[:i]
	}
	if bytes.Contains(primera, []byte(";")) && !bytes.Contains(primera, []byte(",")) {
		return ';'
	}
	return ','
}
