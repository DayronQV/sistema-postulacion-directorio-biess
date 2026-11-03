package domain_test

import (
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestEsNacionalidadEcuatoriana(t *testing.T) {
	casos := map[string]bool{
		"Ecuatoriana":         true,
		"ecuatoriana":         true,
		"ECUATORIANA":         true,
		"  Ecuatoriana  ":     true,
		"\tEcuatoriana\n":     true,
		"Ecuatoriano":         true,
		"ECUATORIANO":         true,
		"Écuatoriana":         true,
		"Ecuatorianá":         true,
		"Colombiana":          false,
		"":                    false,
		"   ":                 false,
		"Ecuador":             false,
		"No ecuatoriana":      false,
		"Ecuato riana":        false,
		"Ecuatoriana-Peruana": false,
	}
	for entrada, esperado := range casos {
		if got := domain.EsNacionalidadEcuatoriana(entrada); got != esperado {
			t.Errorf("EsNacionalidadEcuatoriana(%q) = %v, se esperaba %v", entrada, got, esperado)
		}
	}
}

func TestEsNacionalidadEcuatoriana_AceptaLaConstanteDelDominio(t *testing.T) {
	if !domain.EsNacionalidadEcuatoriana(domain.NacionalidadEcuatoriana) {
		t.Fatalf("la constante %q debe ser ecuatoriana", domain.NacionalidadEcuatoriana)
	}
}
