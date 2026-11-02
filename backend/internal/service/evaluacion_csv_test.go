package service_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

const cabeceraCSV = "enunciado,opciones,respuesta_correcta\n"

func filaCSV(n int) string {
	return fmt.Sprintf("Pregunta %d,Uno|Dos|Tres|Cuatro,%d\n", n, n%4)
}

func TestCrearPreguntasCSV_CargaMasivaDe25(t *testing.T) {
	svc, f := nuevoBanco(0)
	var sb strings.Builder
	sb.WriteString(cabeceraCSV)
	for i := 1; i <= 25; i++ {
		sb.WriteString(filaCSV(i))
	}
	n, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(sb.String()))
	if err != nil || n != 25 {
		t.Fatalf("creadas = %d, err = %v", n, err)
	}
	if f.llamadas != 1 || len(f.preguntas) != 25 {
		t.Fatalf("debe guardarse en un solo bloque: llamadas = %d, guardadas = %d", f.llamadas, len(f.preguntas))
	}
	p := f.preguntas[6] // Pregunta 7 → respuesta 3
	if p.Enunciado != "Pregunta 7" || len(p.Opciones) != 4 || p.Opciones[3] != "Cuatro" || p.RespuestaCorrecta != 3 {
		t.Fatalf("pregunta 7 = %+v", p)
	}
}

func TestCrearPreguntasCSV_FormatosAceptados(t *testing.T) {
	casos := []struct {
		nombre, csv string
	}{
		{"con BOM de UTF-8", "\xEF\xBB\xBF" + cabeceraCSV + filaCSV(1)},
		{"con punto y coma, como Excel en español", "enunciado;opciones;respuesta_correcta\nPregunta 1;Uno|Dos|Tres|Cuatro;1\n"},
		{"con saltos de línea de Windows", "enunciado,opciones,respuesta_correcta\r\nPregunta 1,Uno|Dos|Tres|Cuatro,1\r\n"},
		{"con la cabecera en otro orden y en mayúsculas", "RESPUESTA_CORRECTA,Enunciado,Opciones\n1,Pregunta 1,Uno|Dos|Tres|Cuatro\n"},
		{"con una columna de más", "id,enunciado,opciones,respuesta_correcta\n9,Pregunta 1,Uno|Dos|Tres|Cuatro,1\n"},
		{"con líneas en blanco", cabeceraCSV + "\n" + filaCSV(1) + "\n\n"},
		{"sin salto de línea final", cabeceraCSV + strings.TrimSuffix(filaCSV(1), "\n")},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			svc, f := nuevoBanco(0)
			n, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(c.csv))
			if err != nil || n != 1 {
				t.Fatalf("creadas = %d, err = %v (%v)", n, err, camposEval(t, err))
			}
			if f.preguntas[0].Enunciado != "Pregunta 1" || f.preguntas[0].RespuestaCorrecta != 1 || len(f.preguntas[0].Opciones) != 4 {
				t.Fatalf("pregunta = %+v", f.preguntas[0])
			}
		})
	}
}

func TestCrearPreguntasCSV_EnunciadoEntreComillasConComasYSaltos(t *testing.T) {
	svc, f := nuevoBanco(0)
	csv := cabeceraCSV + "\"¿Cuál es, en orden, la capital?\nResponda con cuidado\",\"Quito|Guayaquil, la grande\",0\n"
	if _, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(csv)); err != nil {
		t.Fatalf("error inesperado: %v (%v)", err, camposEval(t, err))
	}
	if !strings.Contains(f.preguntas[0].Enunciado, "capital?\nResponda") || f.preguntas[0].Opciones[1] != "Guayaquil, la grande" {
		t.Fatalf("pregunta = %+v", f.preguntas[0])
	}
}

func TestCrearPreguntasCSV_ErroresConElNumeroDeLinea(t *testing.T) {
	csv := cabeceraCSV +
		filaCSV(1) + // línea 2
		",Uno|Dos,0\n" + // línea 3: sin enunciado
		"Pregunta 3,Uno,0\n" + // línea 4: una sola opción
		"Pregunta 4,Uno|Dos,abc\n" + // línea 5: respuesta no numérica
		"Pregunta 5,Uno|Dos,7\n" + // línea 6: respuesta fuera de rango
		"Pregunta 6,Uno|Dos\n" + // línea 7: faltan columnas
		filaCSV(7) // línea 8
	svc, f := nuevoBanco(0)
	_, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(csv))
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba 400, llegó %v", err)
	}
	campos := camposEval(t, err)
	for _, esperado := range []string{
		"filas[3].enunciado", "filas[4].opciones", "filas[5].respuesta_correcta", "filas[6].respuesta_correcta", "filas[7].respuesta_correcta",
	} {
		if campos[esperado] == "" {
			t.Errorf("falta %q en %v", esperado, campos)
		}
	}
	if campos["filas[2].enunciado"] != "" || campos["filas[8].enunciado"] != "" {
		t.Errorf("las filas correctas no deben reportarse: %v", campos)
	}
	if f.llamadas != 0 || len(f.preguntas) != 0 {
		t.Fatal("con un error no se guarda ninguna fila")
	}
}

func TestCrearPreguntasCSV_ArchivosInvalidos(t *testing.T) {
	var demasiadas strings.Builder
	demasiadas.WriteString(cabeceraCSV)
	for i := 0; i <= domain.MaxPreguntasPorCarga; i++ {
		demasiadas.WriteString(filaCSV(i))
	}
	casos := []struct {
		nombre, csv, campo string
	}{
		{"vacío", "", "archivo"},
		{"solo espacios", "  \n ", "archivo"},
		{"sin la columna opciones", "enunciado,respuesta_correcta\nPregunta,1\n", "archivo"},
		{"sin la columna respuesta_correcta", "enunciado,opciones\nPregunta,Uno|Dos\n", "archivo"},
		{"sin la columna enunciado", "opciones,respuesta_correcta\nUno|Dos,1\n", "archivo"},
		{"solo la cabecera", cabeceraCSV, "archivo"},
		{"comillas sin cerrar", cabeceraCSV + "\"Pregunta,Uno|Dos,1\n", "filas[2]"},
		{"más del máximo de preguntas", demasiadas.String(), "archivo"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			svc, f := nuevoBanco(0)
			_, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(c.csv))
			if !errors.Is(err, domain.ErrValidation) || camposEval(t, err)[c.campo] == "" {
				t.Fatalf("se esperaba 400 con fields[%q], llegó %v (%v)", c.campo, err, camposEval(t, err))
			}
			if f.llamadas != 0 {
				t.Fatal("no se debe guardar")
			}
		})
	}
}

func TestCrearPreguntasCSV_ArchivoNoUTF8_400SinGuardar(t *testing.T) {
	svc, f := nuevoBanco(0)
	// "¿Capital de Ecuador?" guardado por Excel como CSV en Windows-1252: 0xBF es "¿" y no
	// es UTF-8 válido. Sin la validación, PostgreSQL rechazaría el INSERT con un 500.
	archivo := cabeceraCSV + "\xBFCapital de Ecuador?,Guayaquil|Quito|Cuenca,1\n"

	_, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(archivo))

	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) || !strings.Contains(de.Fields["archivo"], "UTF-8") {
		t.Fatalf("err = %#v; se esperaba 400 con el campo archivo pidiendo UTF-8", err)
	}
	if f.llamadas != 0 {
		t.Fatal("un archivo que no es UTF-8 no debe llegar al repositorio")
	}
}

func TestCrearPreguntasCSV_UTF8ConTildesSeAcepta(t *testing.T) {
	svc, f := nuevoBanco(0)
	archivo := "\xEF\xBB\xBF" + cabeceraCSV + "¿Capital de Ecuador?,Guayaquil|Quito|Cuenca,1\n"

	n, err := svc.CrearPreguntasCSV(context.Background(), strings.NewReader(archivo))

	if err != nil || n != 1 || f.preguntas[0].Enunciado != "¿Capital de Ecuador?" {
		t.Fatalf("creadas = %d, err = %v, preguntas = %+v", n, err, f.preguntas)
	}
}
