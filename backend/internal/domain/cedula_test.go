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
		// Cadenas de 10 bytes con un carácter que no es dígito: fallan por el chequeo
		// de dígitos y no por la longitud.
		{"con espacio al inicio", " 171234567", false},
		{"con guion al final", "171234567-", false},
		{"con carácter no ASCII (2 bytes)", "17123456é", false},

		// Provincia y tercer dígito: todas estas cédulas tienen el dígito verificador
		// CORRECTO, así que lo único que las hace inválidas es la regla que nombran.
		{"provincia 00", "0012345674", false},
		{"provincia 25", "2512345675", false},
		{"provincia 30", "3012345678", false},
		{"tercer dígito 6", "1762345674", false},
		{"tercer dígito 9", "1792345678", false},
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
