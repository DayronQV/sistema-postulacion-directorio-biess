package handler_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

type personalJSON struct {
	Estado              string `json:"estado"`
	InformacionPersonal struct {
		Cedula          string `json:"cedula"`
		Nombres         string `json:"nombres"`
		Apellidos       string `json:"apellidos"`
		Genero          string `json:"genero"`
		FechaNacimiento string `json:"fecha_nacimiento"`
		Nacionalidad    string `json:"nacionalidad"`
		EstadoCivil     string `json:"estado_civil"`
		Pareja          *struct {
			NombresApellidos string `json:"nombres_apellidos"`
			Cedula           string `json:"cedula"`
		} `json:"pareja"`
	} `json:"informacion_personal"`
}

func TestPostulacionObtener_InformacionPersonalCompleta(t *testing.T) {
	p := postulacionDeEjemplo()
	p.Personal = domain.InformacionPersonal{
		Cedula: "0912345675", Nombres: "María Fernanda", Apellidos: "Zambrano Cedeño", Genero: "Femenino",
		FechaNacimiento: time.Date(1978, 11, 2, 0, 0, 0, 0, time.UTC), Nacionalidad: "Ecuatoriana",
		EstadoCivil: domain.EstadoCasado,
		Pareja:      &domain.Pareja{NombresApellidos: "Luis Alberto Párraga Mera", Cedula: "0913579249"},
	}
	w := send(newPostulacionRouter(&fakePostulacion{resultado: p}), "GET", "postulante", "")
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got personalJSON
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	ip := got.InformacionPersonal
	if ip.Cedula != "0912345675" || ip.Nombres != "María Fernanda" || ip.Apellidos != "Zambrano Cedeño" ||
		ip.Genero != "Femenino" || ip.FechaNacimiento != "1978-11-02" || ip.Nacionalidad != "Ecuatoriana" ||
		ip.EstadoCivil != "Casado/a" {
		t.Fatalf("información personal inesperada: %s", w.Body.String())
	}
	if ip.Pareja == nil || ip.Pareja.NombresApellidos != "Luis Alberto Párraga Mera" || ip.Pareja.Cedula != "0913579249" {
		t.Fatalf("faltan los datos de la pareja: %s", w.Body.String())
	}
}

func TestPostulacionObtener_SinParejaNoIncluyeLaClave(t *testing.T) {
	w := send(newPostulacionRouter(&fakePostulacion{resultado: postulacionDeEjemplo()}), "GET", "postulante", "")
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), `"pareja"`) {
		t.Fatalf("una persona soltera no debe traer pareja: %s", w.Body.String())
	}
}

func TestPostulacionObtener_Cerrada(t *testing.T) {
	p := postulacionDeEjemplo()
	p.Estado = domain.EstadoCerrada
	p.Personal.Nacionalidad = "Colombiana"
	w := send(newPostulacionRouter(&fakePostulacion{resultado: p}), "GET", "postulante", "")
	if w.Code != 200 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got personalJSON
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Estado != "cerrada" || got.InformacionPersonal.Nacionalidad != "Colombiana" {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
}

func TestPostulacionGuardar_NacionalidadResponde422ConElMensaje(t *testing.T) {
	const mensaje = "Solo pueden postularse personas de nacionalidad ecuatoriana"
	uc := &fakePostulacion{err: domain.Unprocessable(mensaje, nil)}
	w := send(newPostulacionRouter(uc), "PUT", "postulante", `{}`)
	if w.Code != 422 {
		t.Fatalf("respuesta = %d %s", w.Code, w.Body.String())
	}
	var got struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Error.Code != "VALIDATION_ERROR" || got.Error.Message != mensaje {
		t.Fatalf("formato de error inesperado: %s", w.Body.String())
	}
}
