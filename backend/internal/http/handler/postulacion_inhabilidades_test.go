package handler_test

import (
	"encoding/json"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestPostulacionGuardar_LeeLasInhabilidades(t *testing.T) {
	uc := &fakePostulacion{resultado: postulacionDeEjemplo()}
	body := `{"inhabilidades": {"1": true, "2": false, "7": true, "13": null, "99": true,
		"removido_detalle": {"institucion": "SRI", "cargo": "Analista", "fecha_fin": "2024-01-31", "motivo": "Fin"}}}`
	w := send(newPostulacionRouter(uc), "PUT", "postulante", body)
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	got := uc.guardado.Inhabilidades
	if len(got.Respuestas) != 3 || !got.Respuestas[1] || got.Respuestas[2] || !got.Respuestas[7] {
		t.Fatalf("respuestas = %v (la 13 en null y la 99 no cuentan)", got.Respuestas)
	}
	want := domain.RemovidoDetalle{Institucion: "SRI", Cargo: "Analista", FechaFin: "2024-01-31", Motivo: "Fin"}
	if got.Removido != want {
		t.Fatalf("detalle = %+v", got.Removido)
	}
}

func TestPostulacionGuardar_RespuestaDeInhabilidadInvalidaEs400(t *testing.T) {
	uc := &fakePostulacion{resultado: postulacionDeEjemplo()}
	w := send(newPostulacionRouter(uc), "PUT", "postulante", `{"inhabilidades": {"1": "si"}}`)
	if w.Code != 400 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
}

func TestPostulacion_RespondeLasInhabilidadesYLaElegibilidad(t *testing.T) {
	p := postulacionDeEjemplo()
	p.Formulario.Inhabilidades = domain.Inhabilidades{
		Respuestas: map[int]bool{1: true, 7: true},
		Removido:   domain.RemovidoDetalle{Institucion: "SRI", Motivo: "Fin"},
	}
	p.Formulario.Elegible = false
	w := send(newPostulacionRouter(&fakePostulacion{resultado: p}), "GET", "postulante", "")
	var got struct {
		Inhabilidades map[string]json.RawMessage `json:"inhabilidades"`
		Elegible      *bool                      `json:"elegible"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if string(got.Inhabilidades["1"]) != "true" || string(got.Inhabilidades["7"]) != "true" || got.Inhabilidades["removido_detalle"] == nil {
		t.Fatalf("inhabilidades = %v", got.Inhabilidades)
	}
	if _, ok := got.Inhabilidades["2"]; ok {
		t.Fatal("una pregunta sin responder no debe aparecer")
	}
	if got.Elegible == nil || *got.Elegible {
		t.Fatalf("elegible = %v", got.Elegible)
	}
}

func TestPostulacion_SinInhabilidadesRespondeObjetoVacio(t *testing.T) {
	w := send(newPostulacionRouter(&fakePostulacion{resultado: postulacionDeEjemplo()}), "GET", "postulante", "")
	var got map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if string(got["inhabilidades"]) != "{}" {
		t.Fatalf("inhabilidades = %s", got["inhabilidades"])
	}
}

func TestPostulacionGuardar_RespuestaInvalidaIndicaElCampo(t *testing.T) {
	w := send(newPostulacionRouter(&fakePostulacion{}), "PUT", "postulante", `{"inhabilidades": {"3": "si"}}`)
	var got struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error.Fields["inhabilidades.3"] == "" {
		t.Fatalf("fields = %v", got.Error.Fields)
	}
}
