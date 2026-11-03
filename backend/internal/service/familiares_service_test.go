package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

func familiarValido(nombre string) domain.Familiar {
	return domain.Familiar{NombresApellidos: nombre, Parentesco: "hijo/a", Institucion: "BIESS", Area: "Finanzas", Cargo: "Analista"}
}

func TestGuardar_FamiliaresValidacion(t *testing.T) {
	casos := map[string]struct {
		mutar func(*domain.Formulario)
		campo string
	}{
		"sí sin filas": {func(f *domain.Formulario) { f.Familiares = nil }, "familiares"},
		"no con filas": {func(f *domain.Formulario) { f.TieneFamiliares = false }, "familiares"},
		"demasiadas filas": {func(f *domain.Formulario) {
			f.Familiares = make([]domain.Familiar, domain.MaxFamiliares+1)
			for i := range f.Familiares {
				f.Familiares[i] = familiarValido("Ana")
			}
		}, "familiares"},
		"parentesco fuera de la lista": {func(f *domain.Formulario) { f.Familiares[0].Parentesco = "primo" }, "familiares[0].parentesco"},
		"parentesco vacío":             {func(f *domain.Formulario) { f.Familiares[0].Parentesco = "  " }, "familiares[0].parentesco"},
		"parentesco en mayúsculas":     {func(f *domain.Formulario) { f.Familiares[0].Parentesco = "Madre" }, "familiares[0].parentesco"},
		"institución vacía":            {func(f *domain.Formulario) { f.Familiares[0].Institucion = "" }, "familiares[0].institucion"},
		"institución fuera de lista":   {func(f *domain.Formulario) { f.Familiares[0].Institucion = "iess" }, "familiares[0].institucion"},
		"nombre largo": {func(f *domain.Formulario) {
			f.Familiares[0].NombresApellidos = strings.Repeat("a", domain.MaxLongitudNombreFamiliar+1)
		}, "familiares[0].nombres_apellidos"},
		"área vacía": {func(f *domain.Formulario) { f.Familiares[0].Area = "" }, "familiares[0].area"},
		"área larga": {func(f *domain.Formulario) {
			f.Familiares[0].Area = strings.Repeat("a", domain.MaxLongitudAreaFamiliar+1)
		}, "familiares[0].area"},
		"cargo vacío": {func(f *domain.Formulario) { f.Familiares[0].Cargo = "\t" }, "familiares[0].cargo"},
		"cargo largo": {func(f *domain.Formulario) {
			f.Familiares[0].Cargo = strings.Repeat("a", domain.MaxLongitudCargoFamiliar+1)
		}, "familiares[0].cargo"},
		"el error indica la fila": {func(f *domain.Formulario) {
			f.Familiares = append(f.Familiares, familiarValido("Beto"), familiarValido("Carla"))
			f.Familiares[2].Area = ""
		}, "familiares[2].area"},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
			fm := formularioValido()
			c.mutar(&fm)

			_, err := f.svc.Guardar(context.Background(), postulanteID, fm)
			var de *domain.Error
			if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("se esperaba 400, llegó %v", err)
			}
			if _, ok := de.Fields[c.campo]; !ok {
				t.Fatalf("falta el campo %q en %v", c.campo, de.Fields)
			}
			if f.tx.ejecutadas != 0 {
				t.Fatalf("una validación fallida no debe abrir transacción")
			}
		})
	}
}

func TestGuardar_FamiliaresNoReportaFilasValidas(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	fm := formularioValido()
	fm.Familiares = append(fm.Familiares, familiarValido("Beto"))
	fm.Familiares[1].Cargo = ""

	_, err := f.svc.Guardar(context.Background(), postulanteID, fm)
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("se esperaba un error de dominio, llegó %v", err)
	}
	if len(de.Fields) != 1 || de.Fields["familiares[1].cargo"] == "" {
		t.Fatalf("solo debía fallar la fila 1: %v", de.Fields)
	}
}

func TestGuardar_FamiliaresAceptaTodosLosParentescos(t *testing.T) {
	if len(domain.Parentescos) != 15 {
		t.Fatalf("la lista tiene %d parentescos, se esperaban 15", len(domain.Parentescos))
	}
	for _, p := range domain.Parentescos {
		f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
		fm := formularioValido()
		fm.Familiares[0].Parentesco = p
		if _, err := f.svc.Guardar(context.Background(), postulanteID, fm); err != nil {
			t.Errorf("parentesco %q rechazado: %v", p, err)
		}
	}
}

func TestGuardar_FamiliaresLimpiaEspaciosYConservaElOrden(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	fm := formularioValido()
	fm.Familiares = []domain.Familiar{
		{NombresApellidos: "  Zoila Mora ", Parentesco: " madre", Institucion: "IESS ", Area: " Legal ", Cargo: " Jefa "},
		familiarValido("Ana"),
	}

	if _, err := f.svc.Guardar(context.Background(), postulanteID, fm); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	got := f.store.p.Formulario.Familiares
	want := domain.Familiar{NombresApellidos: "Zoila Mora", Parentesco: "madre", Institucion: "IESS", Area: "Legal", Cargo: "Jefa"}
	if len(got) != 2 || got[0] != want || got[1].NombresApellidos != "Ana" {
		t.Fatalf("familiares guardados = %+v", got)
	}
	if fm.Familiares[0].NombresApellidos != "  Zoila Mora " {
		t.Fatalf("no debe modificar el formulario recibido")
	}
}

func TestGuardar_FamiliaresNoSinFilasEsValido(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	fm := formularioValido()
	fm.TieneFamiliares = false
	fm.Familiares = nil

	if _, err := f.svc.Guardar(context.Background(), postulanteID, fm); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if f.store.p.Formulario.TieneFamiliares || len(f.store.p.Formulario.Familiares) != 0 {
		t.Fatalf("formulario guardado = %+v", f.store.p.Formulario)
	}
}
