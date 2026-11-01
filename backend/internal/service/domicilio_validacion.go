package service

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// paisEcuador es el nombre del país cuyo domicilio se elige por listas (provincia y
// cantón por id). Con cualquier otro país, provincia y ciudad son texto libre.
const paisEcuador = "Ecuador"

// Longitudes máximas de los campos de texto del domicilio, en caracteres.
const (
	maxCalle       = 100
	maxNumeroCalle = 20
	maxSector      = 100
	maxReferencia  = 200
	maxTextoGeo    = 100
	maxCorreo      = 254
)

const (
	msgObligatorio   = "Campo obligatorio"
	msgNoCorresponde = "No corresponde con el país seleccionado"
)

var (
	// Celular ecuatoriano: 10 dígitos que empiezan en 09.
	reCelular = regexp.MustCompile(`^09\d{8}$`)
	// Teléfono de domicilio: solo dígitos, con o sin código de área y de país.
	reTelefonoDomicilio = regexp.MustCompile(`^\d{7,15}$`)
)

// ConCatalogos activa la validación geográfica del domicilio con los catálogos dados
// (país → provincia → cantón). Sin ella solo se valida el formato de los campos, así que
// el cableado de producción siempre debe llamarla. Devuelve el mismo servicio para
// encadenarla al constructor.
func (s *PostulacionService) ConCatalogos(c domain.Catalogos) *PostulacionService {
	s.catalogos = c
	return s
}

// campoDomicilio es el nombre del campo en `fields`, con la ruta del JSON de la API.
func campoDomicilio(nombre string) string { return "domicilio." + nombre }

// validarDomicilioFormato agrega a fields los errores de formato y longitud del
// domicilio. No consulta los catálogos. Todos los campos son opcionales mientras la
// postulación es un borrador: solo se valida lo que viene escrito.
func validarDomicilioFormato(d domain.Domicilio, fields map[string]string) {
	if d.TelefonoCelular != "" && !reCelular.MatchString(d.TelefonoCelular) {
		fields[campoDomicilio("telefono_celular")] = "El celular debe tener 10 dígitos y empezar con 09"
	}
	if d.TelefonoDomicilio != "" && !reTelefonoDomicilio.MatchString(d.TelefonoDomicilio) {
		fields[campoDomicilio("telefono_domicilio")] = "El teléfono debe tener entre 7 y 15 dígitos"
	}
	if d.EmailSecundario != "" {
		switch {
		case utf8.RuneCountInString(d.EmailSecundario) > maxCorreo:
			fields[campoDomicilio("email_secundario")] = "Correo electrónico demasiado largo"
		case validarEmail(d.EmailSecundario) != "":
			fields[campoDomicilio("email_secundario")] = "Correo electrónico no válido"
		}
	}

	limites := []struct {
		campo, valor string
		max          int
	}{
		{"calle_principal", d.CallePrincipal, maxCalle},
		{"numero", d.Numero, maxNumeroCalle},
		{"calle_secundaria", d.CalleSecundaria, maxCalle},
		{"sector", d.Sector, maxSector},
		{"referencia", d.Referencia, maxReferencia},
		{"provincia_texto", d.ProvinciaTexto, maxTextoGeo},
		{"ciudad_texto", d.CiudadTexto, maxTextoGeo},
	}
	for _, l := range limites {
		if utf8.RuneCountInString(l.valor) > l.max {
			fields[campoDomicilio(l.campo)] = "Máximo " + strconv.Itoa(l.max) + " caracteres"
		}
	}
}

// validarGeografia comprueba la coherencia del país, la provincia y el cantón contra los
// catálogos. Si el país es Ecuador, la provincia y el cantón son obligatorios por id, la
// provincia debe pertenecer al país y el cantón a la provincia. Con otro país, provincia
// y ciudad se escriben como texto y no se aceptan ids. Sin país no se acepta nada de lo
// demás. Devuelve domain.Validation con `fields`, o el error del catálogo si falla.
func (s *PostulacionService) validarGeografia(ctx context.Context, d domain.Domicilio) error {
	if s.catalogos == nil {
		return nil
	}
	fields := map[string]string{}

	switch {
	case d.PaisID == nil:
		if d.ProvinciaID != nil || d.CantonID != nil || d.ProvinciaTexto != "" || d.CiudadTexto != "" {
			fields[campoDomicilio("pais_id")] = "Seleccione primero el país"
		}
	default:
		paises, err := s.catalogos.Paises(ctx)
		if err != nil {
			return err
		}
		pais, ok := buscarItem(paises, *d.PaisID)
		switch {
		case !ok:
			fields[campoDomicilio("pais_id")] = "País no válido"
		case pais.Nombre == paisEcuador:
			if err := s.validarEcuador(ctx, d, fields); err != nil {
				return err
			}
		default:
			validarExtranjero(d, fields)
		}
	}

	if len(fields) > 0 {
		return domain.Validation("Datos inválidos", fields)
	}
	return nil
}

// validarEcuador exige provincia y cantón por id, coherentes entre sí y con el país.
func (s *PostulacionService) validarEcuador(ctx context.Context, d domain.Domicilio, fields map[string]string) error {
	if d.ProvinciaTexto != "" {
		fields[campoDomicilio("provincia_texto")] = "Con Ecuador se elige la provincia de la lista"
	}
	if d.CiudadTexto != "" {
		fields[campoDomicilio("ciudad_texto")] = "Con Ecuador se elige el cantón de la lista"
	}

	provinciaValida := false
	switch {
	case d.ProvinciaID == nil:
		fields[campoDomicilio("provincia_id")] = msgObligatorio
	default:
		provincias, err := s.catalogos.Provincias(ctx, *d.PaisID)
		if err != nil {
			return err
		}
		if _, ok := buscarItem(provincias, *d.ProvinciaID); ok {
			provinciaValida = true
		} else {
			fields[campoDomicilio("provincia_id")] = "La provincia no pertenece al país"
		}
	}

	switch {
	case d.CantonID == nil:
		fields[campoDomicilio("canton_id")] = msgObligatorio
	case provinciaValida:
		cantones, err := s.catalogos.Cantones(ctx, *d.ProvinciaID)
		if err != nil {
			return err
		}
		if _, ok := buscarItem(cantones, *d.CantonID); !ok {
			fields[campoDomicilio("canton_id")] = "El cantón no pertenece a la provincia"
		}
	}
	return nil
}

// validarExtranjero exige provincia y ciudad como texto y rechaza los ids.
func validarExtranjero(d domain.Domicilio, fields map[string]string) {
	if d.ProvinciaID != nil {
		fields[campoDomicilio("provincia_id")] = msgNoCorresponde
	}
	if d.CantonID != nil {
		fields[campoDomicilio("canton_id")] = msgNoCorresponde
	}
	if strings.TrimSpace(d.ProvinciaTexto) == "" {
		fields[campoDomicilio("provincia_texto")] = msgObligatorio
	}
	if strings.TrimSpace(d.CiudadTexto) == "" {
		fields[campoDomicilio("ciudad_texto")] = msgObligatorio
	}
}

func buscarItem(items []domain.ItemCatalogo, id int) (domain.ItemCatalogo, bool) {
	for _, it := range items {
		if it.ID == id {
			return it, true
		}
	}
	return domain.ItemCatalogo{}, false
}
