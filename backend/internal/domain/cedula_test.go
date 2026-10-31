package domain_test

import (
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestCedulaValida(t *testing.T) {
	cases := []struct {
		name   string
		cedula string
		valida bool
	}{
		// Válidas
		{"flujo feliz", "1712345675", true},
		{"casada", "0912345675", true},
		{"unión de hecho", "1312345679", true},
		{"extranjero con cédula", "0103456786", true},
		{"verificador 0", "1802468130", true},
		{"provincia 24 (límite superior)", "2400000002", true}, // 4+4 = 8 → verificador 2

		// Dígito verificador incorrecto
		{"verificador alterado", "1712345676", false},
		{"un dígito cambiado", "1712345575", false},

		// Longitud y caracteres
		{"vacía", "", false},
		{"9 dígitos", "171234567", false},
		{"11 dígitos", "17123456750", false},
		{"con letras", "17123456A5", false},
		{"con espacio al inicio", " 171234567", false},
		{"con guion", "171234567-5", false},
		{"dígitos de ancho completo", "１７１２３４５６７５", false},

		// Provincia
		{"provincia 00", "0012345675", false},
		{"provincia 25", "2512345678", false},
		{"provincia 30", "3012345670", false},

		// Tercer dígito
		{"tercer dígito 6", "1762345670", false},
		{"tercer dígito 9", "1792345670", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := domain.CedulaValida(c.cedula); got != c.valida {
				t.Fatalf("CedulaValida(%q) = %v, se esperaba %v", c.cedula, got, c.valida)
			}
		})
	}
}

// Cédula con la provincia y el tercer dígito correctos, para confirmar que lo que
// la hace inválida es solo el dígito verificador.
func TestCedulaValidaUsaElDigitoVerificador(t *testing.T) {
	valida := "1712345675"
	for d := '0'; d <= '9'; d++ {
		cedula := valida[:9] + string(d)
		esperado := cedula == valida
		if got := domain.CedulaValida(cedula); got != esperado {
			t.Errorf("CedulaValida(%q) = %v, se esperaba %v", cedula, got, esperado)
		}
	}
}

func TestValidarCedulaUsaElFormatoDeLaAPI(t *testing.T) {
	if err := domain.ValidarCedula("1712345675"); err != nil {
		t.Fatalf("una cédula válida no debe dar error: %v", err)
	}

	err := domain.ValidarCedula("1712345676")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba ErrValidation, got %v", err)
	}
	var de *domain.Error
	if !errors.As(err, &de) || de.Fields["cedula"] != "Cédula no válida" {
		t.Fatalf("falta el campo cedula en fields: %+v", de)
	}
}
