// Package padron lee el archivo JSON con las personas que consultan los simuladores del
// Registro Civil y del Historial Laboral del IESS.
//
// Sirve para desplegar con personas concretas sin escribir sus datos en el código: el
// archivo vive fuera del repositorio y se indica con PADRON_ARCHIVO. Sin esa variable,
// los simuladores usan sus ciudadanos de prueba fijos.
//
// Formato: un arreglo JSON; padron.example.json, junto a este archivo, es un ejemplo
// completo (una prueba lo carga, así que siempre es válido). Por persona:
//   - cedula: 10 dígitos válidos por módulo 10.
//   - nombres, apellidos, nacionalidad: texto (solo "Ecuatoriana" puede postular).
//   - genero: "Masculino" o "Femenino".
//   - fecha_nacimiento, fecha_expedicion: "AAAA-MM-DD", en el pasado.
//   - codigo_dactilar: 10 letras mayúsculas o dígitos, por ejemplo "V1234V5678".
//   - estado_civil: "Soltero/a", "Casado/a", "Unión de hecho", "Divorciado/a" o "Viudo/a".
//   - pareja: {"nombres_apellidos", "cedula"}; obligatoria con "Casado/a" o "Unión de
//     hecho" y prohibida con los demás.
//   - calidades: lista con "afiliado", "jubilado" y/o "empleador" (vacía = ninguna).
//     Decide qué cargo puede elegir la persona al postular.
package padron

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Calidades que admite el archivo, en el campo "calidades" de cada persona.
const (
	CalidadAfiliado  = "afiliado"
	CalidadJubilado  = "jubilado"
	CalidadEmpleador = "empleador"
)

const layoutFecha = "2006-01-02"

var codigoDactilar = regexp.MustCompile(`^[A-Z0-9]{10}$`)

// Registro es lo que el padrón sabe de una persona: sus datos del Registro Civil y su
// calidad ante el IESS.
type Registro struct {
	Ciudadano domain.Ciudadano
	Calidad   domain.Calidad
}

type personaJSON struct {
	Cedula          string      `json:"cedula"`
	Nombres         string      `json:"nombres"`
	Apellidos       string      `json:"apellidos"`
	Genero          string      `json:"genero"`
	FechaNacimiento string      `json:"fecha_nacimiento"`
	Nacionalidad    string      `json:"nacionalidad"`
	EstadoCivil     string      `json:"estado_civil"`
	Pareja          *parejaJSON `json:"pareja,omitempty"`
	FechaExpedicion string      `json:"fecha_expedicion"`
	CodigoDactilar  string      `json:"codigo_dactilar"`
	Calidades       []string    `json:"calidades"`
}

type parejaJSON struct {
	NombresApellidos string `json:"nombres_apellidos"`
	Cedula           string `json:"cedula"`
}

// Cargar lee y valida el archivo. Si algo está mal devuelve un solo error que lista todos
// los problemas (con la posición de la persona en el archivo, sin sus datos), para que el
// servidor no arranque con un padrón a medias.
func Cargar(ruta string) ([]Registro, error) {
	// Si el archivo no existe en el servidor, Docker monta en su lugar una carpeta vacía
	// con ese nombre; se avisa con un mensaje que dice qué hacer.
	if info, err := os.Stat(ruta); err == nil && info.IsDir() {
		return nil, fmt.Errorf("el padrón %s es una carpeta, no un archivo: falta copiar padron.json al servidor (ver PADRON_ARCHIVO en .env.example)", ruta)
	}
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("leyendo el padrón %s: %w", ruta, err)
	}
	dec := json.NewDecoder(bytes.NewReader(datos))
	dec.DisallowUnknownFields()
	var personas []personaJSON
	if err := dec.Decode(&personas); err != nil {
		return nil, fmt.Errorf("el padrón %s no es un arreglo JSON válido: %w", ruta, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("el padrón %s tiene texto después del arreglo JSON", ruta)
	}
	if len(personas) == 0 {
		return nil, fmt.Errorf("el padrón %s no tiene ninguna persona", ruta)
	}

	var problemas []string
	vistas := make(map[string]bool, len(personas))
	registros := make([]Registro, 0, len(personas))
	for i, p := range personas {
		r, errs := convertir(p)
		if vistas[p.Cedula] && p.Cedula != "" {
			errs = append(errs, "cédula repetida")
		}
		vistas[p.Cedula] = true
		for _, e := range errs {
			problemas = append(problemas, fmt.Sprintf("persona %d: %s", i+1, e))
		}
		registros = append(registros, r)
	}
	if len(problemas) > 0 {
		return nil, errors.New("padrón inválido (" + ruta + "): " + strings.Join(problemas, "; "))
	}
	return registros, nil
}

func convertir(p personaJSON) (Registro, []string) {
	var errs []string
	falla := func(msg string) { errs = append(errs, msg) }

	if !domain.CedulaValida(p.Cedula) {
		falla("cédula no válida")
	}
	if strings.TrimSpace(p.Nombres) == "" || strings.TrimSpace(p.Apellidos) == "" {
		falla("faltan nombres o apellidos")
	}
	if p.Genero != "Masculino" && p.Genero != "Femenino" {
		falla(`genero debe ser "Masculino" o "Femenino"`)
	}
	if strings.TrimSpace(p.Nacionalidad) == "" {
		falla("falta la nacionalidad")
	}
	nacimiento, err := time.Parse(layoutFecha, p.FechaNacimiento)
	if err != nil || nacimiento.After(time.Now()) {
		falla("fecha_nacimiento debe ser AAAA-MM-DD y estar en el pasado")
	}
	expedicion, err := time.Parse(layoutFecha, p.FechaExpedicion)
	if err != nil || expedicion.After(time.Now()) {
		falla("fecha_expedicion debe ser AAAA-MM-DD y estar en el pasado")
	}
	if !codigoDactilar.MatchString(p.CodigoDactilar) {
		falla("codigo_dactilar debe tener 10 caracteres alfanuméricos en mayúscula")
	}

	estado := domain.EstadoCivil(p.EstadoCivil)
	switch estado {
	case domain.EstadoSoltero, domain.EstadoCasado, domain.EstadoUnionDeHecho,
		domain.EstadoDivorciado, domain.EstadoViudo:
	default:
		falla("estado_civil no reconocido")
	}

	var pareja *domain.Pareja
	switch {
	case estado.ConPareja() && p.Pareja == nil:
		falla("el estado civil exige los datos de la pareja")
	case !estado.ConPareja() && p.Pareja != nil:
		falla("solo el estado civil casado o unión de hecho lleva pareja")
	case p.Pareja != nil:
		if strings.TrimSpace(p.Pareja.NombresApellidos) == "" || !domain.CedulaValida(p.Pareja.Cedula) {
			falla("la pareja necesita nombres y una cédula válida")
		}
		pareja = &domain.Pareja{NombresApellidos: p.Pareja.NombresApellidos, Cedula: p.Pareja.Cedula}
	}

	var calidad domain.Calidad
	for _, c := range p.Calidades {
		switch c {
		case CalidadAfiliado:
			calidad.EsAfiliado = true
		case CalidadJubilado:
			calidad.EsJubilado = true
		case CalidadEmpleador:
			calidad.EsEmpleador = true
		default:
			falla(`calidades solo admite "afiliado", "jubilado" y "empleador"`)
		}
	}

	return Registro{
		Ciudadano: domain.Ciudadano{
			Cedula: p.Cedula, Nombres: p.Nombres, Apellidos: p.Apellidos, Genero: p.Genero,
			FechaNacimiento: nacimiento, Nacionalidad: p.Nacionalidad, EstadoCivil: estado,
			Pareja: pareja, FechaExpedicion: expedicion, CodigoDactilar: p.CodigoDactilar,
		},
		Calidad: calidad,
	}, errs
}
