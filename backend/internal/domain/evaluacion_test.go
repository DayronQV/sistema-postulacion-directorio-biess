package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Criterio de la issue #42: las respuestas correctas nunca se envían al postulante.
func TestPreguntaPublica_NoIncluyeLaRespuestaCorrecta(t *testing.T) {
	p := domain.Pregunta{ID: 7, Enunciado: "¿Capital de Ecuador?", Opciones: []string{"Guayaquil", "Quito"}, RespuestaCorrecta: 1}

	pub := p.Publica()
	if pub.ID != 7 || pub.Enunciado != p.Enunciado || len(pub.Opciones) != 2 {
		t.Fatalf("la vista pública perdió datos: %+v", pub)
	}
	cuerpo, err := json.Marshal(pub)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cuerpo), "respuesta") || strings.Contains(string(cuerpo), "correcta") {
		t.Fatalf("la vista pública no debe traer la respuesta correcta: %s", cuerpo)
	}
	// La vista del administrador sí la trae.
	admin, _ := json.Marshal(p)
	if !strings.Contains(string(admin), `"respuesta_correcta":1`) {
		t.Fatalf("la vista del administrador debe traer la respuesta correcta: %s", admin)
	}
}
