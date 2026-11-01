package domain_test

import (
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

func TestRolValido(t *testing.T) {
	for _, r := range []domain.Rol{domain.RolPostulante, domain.RolAnalista, domain.RolAdministrador} {
		if !r.Valido() {
			t.Errorf("%q debería ser un rol válido", r)
		}
	}
	for _, r := range []domain.Rol{"", "admin", "Postulante", "superusuario"} {
		if r.Valido() {
			t.Errorf("%q no debería ser un rol válido", r)
		}
	}
}
