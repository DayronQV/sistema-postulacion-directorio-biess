package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// BindJSON lee el cuerpo JSON de la petición en dst y valida sus etiquetas `binding`.
// Si algo falla responde 400 VALIDATION_ERROR (con `fields` cuando se puede saber el
// campo) y devuelve false: el handler debe terminar sin hacer nada más.
//
//	var req RegistroRequest
//	if !middleware.BindJSON(c, &req) {
//		return
//	}
func BindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		if verr := bindingError(err); verr != nil {
			err = verr
		}
		Fail(c, err)
		return false
	}
	return true
}

var registerNamesOnce sync.Once

// RegisterJSONFieldNames hace que los errores de validación usen el nombre JSON del
// campo (`cedula`) y no el de Go (`Cedula`). Se llama una vez al armar el router.
func RegisterJSONFieldNames() {
	registerNamesOnce.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				return f.Name
			}
			return name
		})
	})
}

// bindingError convierte los errores de lectura y validación del cuerpo (JSON mal
// formado, tipo incorrecto, campo obligatorio ausente) en un error de dominio 400.
// Así un cuerpo inválido nunca termina en 500, aunque el handler use ShouldBindJSON.
// Devuelve nil si el error no es de lectura ni de validación del cuerpo.
func bindingError(err error) error {
	var (
		syntax  *json.SyntaxError
		typeErr *json.UnmarshalTypeError
		verrs   validator.ValidationErrors
	)
	switch {
	case errors.As(err, &verrs):
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fieldPath(fe)] = fieldMessage(fe)
		}
		return domain.Validation("Datos inválidos", fields)
	case errors.As(err, &typeErr):
		var fields map[string]string
		if typeErr.Field != "" {
			fields = map[string]string{typeErr.Field: "Tipo de dato no válido"}
		}
		return domain.Validation("Datos inválidos", fields)
	case errors.As(err, &syntax), errors.Is(err, io.ErrUnexpectedEOF):
		return domain.Validation("El cuerpo de la petición no es un JSON válido", nil)
	case errors.Is(err, io.EOF):
		return domain.Validation("El cuerpo de la petición está vacío", nil)
	}
	return nil
}

// fieldPath devuelve la ruta del campo sin el nombre de la estructura raíz, por
// ejemplo "familiares[0].cargo".
func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	if i := strings.Index(ns, "."); i >= 0 {
		return ns[i+1:]
	}
	return ns
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Campo obligatorio"
	case "email":
		return "Correo electrónico no válido"
	case "min":
		return "Valor demasiado corto o pequeño (mínimo " + fe.Param() + ")"
	case "max":
		return "Valor demasiado largo o grande (máximo " + fe.Param() + ")"
	case "oneof":
		return "Valor no permitido"
	default:
		return "Valor no válido"
	}
}
