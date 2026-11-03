package domain_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Criterio de la issue #42: las respuestas correctas nunca se envían al postulante. La vista
// pública no tiene ningún campo donde llevarla, así que no puede salir por descuido.
func TestPreguntaPublica_NoTieneCampoParaLaRespuestaCorrecta(t *testing.T) {
	tipo := reflect.TypeOf(domain.PreguntaPublica{})
	for i := 0; i < tipo.NumField(); i++ {
		nombre := strings.ToLower(tipo.Field(i).Name)
		if strings.Contains(nombre, "respuesta") || strings.Contains(nombre, "correcta") {
			t.Fatalf("PreguntaPublica no debe tener el campo %s", tipo.Field(i).Name)
		}
	}
	if tipo.NumField() != 3 {
		t.Fatalf("PreguntaPublica debe tener solo id, enunciado y opciones, tiene %d campos", tipo.NumField())
	}
}

func TestPublica_CopiaLoQueVeElPostulante(t *testing.T) {
	p := domain.Pregunta{ID: 7, Enunciado: "¿Capital de Ecuador?", Opciones: []string{"Guayaquil", "Quito"}, RespuestaCorrecta: 1}
	pub := p.Publica()
	if pub.ID != 7 || pub.Enunciado != p.Enunciado || !slices.Equal(pub.Opciones, p.Opciones) {
		t.Fatalf("la vista pública perdió datos: %+v", pub)
	}
}

// El dominio no conoce el formato JSON: los handlers arman sus propias respuestas.
func TestDominioDeEvaluacion_NoLlevaEtiquetasJSON(t *testing.T) {
	for _, valor := range []any{domain.Pregunta{}, domain.PreguntaPublica{}} {
		tipo := reflect.TypeOf(valor)
		for i := 0; i < tipo.NumField(); i++ {
			if etiqueta, ok := tipo.Field(i).Tag.Lookup("json"); ok {
				t.Errorf("%s.%s no debe llevar la etiqueta json:%q", tipo.Name(), tipo.Field(i).Name, etiqueta)
			}
		}
	}
}
