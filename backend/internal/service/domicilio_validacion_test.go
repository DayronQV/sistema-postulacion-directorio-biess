package service_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

// Ids del catálogo falso. Ecuador tiene dos provincias con cantones; Colombia una.
const (
	idEcuador, idColombia                    = 1, 2
	idPichincha, idGuayas, idAntioquia       = 10, 11, 20
	idQuito, idCayambe, idGuayaquil, idMedel = 100, 101, 110, 200
)

// fakeCatalogosDomicilio implementa domain.Catalogos en memoria y cuenta las consultas.
type fakeCatalogosDomicilio struct {
	err       error
	consultas int
}

func (f *fakeCatalogosDomicilio) Paises(context.Context) ([]domain.ItemCatalogo, error) {
	f.consultas++
	return []domain.ItemCatalogo{{ID: idEcuador, Nombre: "Ecuador"}, {ID: idColombia, Nombre: "Colombia"}}, f.err
}

func (f *fakeCatalogosDomicilio) Provincias(_ context.Context, paisID int) ([]domain.ItemCatalogo, error) {
	f.consultas++
	if f.err != nil {
		return nil, f.err
	}
	switch paisID {
	case idEcuador:
		return []domain.ItemCatalogo{{ID: idPichincha, Nombre: "Pichincha"}, {ID: idGuayas, Nombre: "Guayas"}}, nil
	case idColombia:
		return []domain.ItemCatalogo{{ID: idAntioquia, Nombre: "Antioquia"}}, nil
	}
	return nil, domain.NotFound("El país no existe")
}

func (f *fakeCatalogosDomicilio) Cantones(_ context.Context, provinciaID int) ([]domain.ItemCatalogo, error) {
	f.consultas++
	if f.err != nil {
		return nil, f.err
	}
	switch provinciaID {
	case idPichincha:
		return []domain.ItemCatalogo{{ID: idQuito, Nombre: "Quito"}, {ID: idCayambe, Nombre: "Cayambe"}}, nil
	case idGuayas:
		return []domain.ItemCatalogo{{ID: idGuayaquil, Nombre: "Guayaquil"}}, nil
	case idAntioquia:
		return []domain.ItemCatalogo{{ID: idMedel, Nombre: "Medellín"}}, nil
	}
	return nil, domain.NotFound("La provincia no existe")
}

func ptr(n int) *int { return &n }

// domicilioEcuador es un domicilio válido en Quito.
func domicilioEcuador() domain.Domicilio {
	return domain.Domicilio{
		PaisID: ptr(idEcuador), ProvinciaID: ptr(idPichincha), CantonID: ptr(idQuito),
		TelefonoCelular: "0999999999", EmailSecundario: "otro@correo.com",
		CallePrincipal: "Av. Amazonas", Numero: "N12-34", CalleSecundaria: "Colón",
		Sector: "La Mariscal", Referencia: "Frente al parque", TelefonoDomicilio: "022345678",
	}
}

func camposDe(t *testing.T, err error) map[string]string {
	t.Helper()
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba ErrValidation, llegó %v", err)
	}
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("el error no es *domain.Error: %v", err)
	}
	return de.Fields
}

func TestGuardar_Domicilio_ValidoSeGuarda(t *testing.T) {
	extranjero := domain.Domicilio{
		PaisID: ptr(idColombia), ProvinciaTexto: "Antioquia", CiudadTexto: "Medellín", TelefonoCelular: "0999999999",
	}
	casos := map[string]domain.Domicilio{
		"Ecuador con provincia y cantón por id":      domicilioEcuador(),
		"otro país con provincia y ciudad por texto": extranjero,
		"borrador sin ningún dato":                   {},
		"solo teléfono":                              {TelefonoCelular: "0912345678"},
		"otro país, otra ciudad":                     {PaisID: ptr(idColombia), ProvinciaTexto: "Antioquia", CiudadTexto: "Cali"},
	}
	for nombre, d := range casos {
		t.Run(nombre, func(t *testing.T) {
			f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
			f.svc.ConCatalogos(&fakeCatalogosDomicilio{})

			p, err := f.svc.Guardar(context.Background(), postulanteID, domain.Formulario{Domicilio: d})
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if f.store.actualizadas != 1 || !reflect.DeepEqual(p.Formulario.Domicilio, d) {
				t.Fatalf("no se guardó el domicilio: %+v", p.Formulario.Domicilio)
			}
		})
	}
}

func TestGuardar_Domicilio_RechazaDatosInvalidos(t *testing.T) {
	con := func(cambia func(d *domain.Domicilio)) domain.Domicilio {
		d := domicilioEcuador()
		cambia(&d)
		return d
	}
	casos := []struct {
		nombre    string
		domicilio domain.Domicilio
		campos    []string
	}{
		{"cantón de otra provincia", con(func(d *domain.Domicilio) { d.CantonID = ptr(idGuayaquil) }), []string{"domicilio.canton_id"}},
		{"provincia de otro país", con(func(d *domain.Domicilio) { d.ProvinciaID, d.CantonID = ptr(idAntioquia), ptr(idMedel) }), []string{"domicilio.provincia_id"}},
		{"cantón que no existe", con(func(d *domain.Domicilio) { d.CantonID = ptr(999999) }), []string{"domicilio.canton_id"}},
		{"provincia que no existe", con(func(d *domain.Domicilio) { d.ProvinciaID = ptr(999999) }), []string{"domicilio.provincia_id"}},
		{"país que no existe", con(func(d *domain.Domicilio) { d.PaisID = ptr(999999) }), []string{"domicilio.pais_id"}},
		{"Ecuador sin provincia ni cantón", con(func(d *domain.Domicilio) { d.ProvinciaID, d.CantonID = nil, nil }), []string{"domicilio.provincia_id", "domicilio.canton_id"}},
		{"Ecuador sin cantón", con(func(d *domain.Domicilio) { d.CantonID = nil }), []string{"domicilio.canton_id"}},
		{"Ecuador con provincia y ciudad en texto", con(func(d *domain.Domicilio) { d.ProvinciaTexto, d.CiudadTexto = "Pichincha", "Quito" }), []string{"domicilio.provincia_texto", "domicilio.ciudad_texto"}},
		{"otro país con ids", domain.Domicilio{PaisID: ptr(idColombia), ProvinciaID: ptr(idAntioquia), CantonID: ptr(idMedel), ProvinciaTexto: "A", CiudadTexto: "B"}, []string{"domicilio.provincia_id", "domicilio.canton_id"}},
		{"otro país sin texto", domain.Domicilio{PaisID: ptr(idColombia)}, []string{"domicilio.provincia_texto", "domicilio.ciudad_texto"}},
		{"otro país con texto en blanco", domain.Domicilio{PaisID: ptr(idColombia), ProvinciaTexto: "  ", CiudadTexto: "Cali"}, []string{"domicilio.provincia_texto"}},
		{"provincia sin país", domain.Domicilio{ProvinciaID: ptr(idPichincha)}, []string{"domicilio.pais_id"}},
		{"texto sin país", domain.Domicilio{CiudadTexto: "Cali"}, []string{"domicilio.pais_id"}},
		{"celular que no empieza en 09", con(func(d *domain.Domicilio) { d.TelefonoCelular = "0899999999" }), []string{"domicilio.telefono_celular"}},
		{"celular de 9 dígitos", con(func(d *domain.Domicilio) { d.TelefonoCelular = "099999999" }), []string{"domicilio.telefono_celular"}},
		{"celular de 11 dígitos", con(func(d *domain.Domicilio) { d.TelefonoCelular = "09999999999" }), []string{"domicilio.telefono_celular"}},
		{"celular con letras", con(func(d *domain.Domicilio) { d.TelefonoCelular = "09abcdefgh" }), []string{"domicilio.telefono_celular"}},
		{"celular con +593", con(func(d *domain.Domicilio) { d.TelefonoCelular = "+593999999999" }), []string{"domicilio.telefono_celular"}},
		{"celular con espacios", con(func(d *domain.Domicilio) { d.TelefonoCelular = " 0999999999" }), []string{"domicilio.telefono_celular"}},
		{"correo sin dominio", con(func(d *domain.Domicilio) { d.EmailSecundario = "otro@correo" }), []string{"domicilio.email_secundario"}},
		{"correo con nombre", con(func(d *domain.Domicilio) { d.EmailSecundario = "Ana <ana@correo.com>" }), []string{"domicilio.email_secundario"}},
		{"correo sin arroba", con(func(d *domain.Domicilio) { d.EmailSecundario = "otrocorreo.com" }), []string{"domicilio.email_secundario"}},
		{"correo demasiado largo", con(func(d *domain.Domicilio) { d.EmailSecundario = strings.Repeat("a", 250) + "@b.com" }), []string{"domicilio.email_secundario"}},
		{"teléfono de domicilio con letras", con(func(d *domain.Domicilio) { d.TelefonoDomicilio = "02-234567" }), []string{"domicilio.telefono_domicilio"}},
		{"teléfono de domicilio corto", con(func(d *domain.Domicilio) { d.TelefonoDomicilio = "12345" }), []string{"domicilio.telefono_domicilio"}},
		{"calle principal demasiado larga", con(func(d *domain.Domicilio) { d.CallePrincipal = strings.Repeat("a", 101) }), []string{"domicilio.calle_principal"}},
		{"número demasiado largo", con(func(d *domain.Domicilio) { d.Numero = strings.Repeat("1", 21) }), []string{"domicilio.numero"}},
		{"referencia demasiado larga", con(func(d *domain.Domicilio) { d.Referencia = strings.Repeat("a", 201) }), []string{"domicilio.referencia"}},
		{"formato inválido: no se consulta la geografía", con(func(d *domain.Domicilio) { d.TelefonoCelular, d.CantonID = "1", ptr(idGuayaquil) }), []string{"domicilio.telefono_celular"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
			f.svc.ConCatalogos(&fakeCatalogosDomicilio{})

			_, err := f.svc.Guardar(context.Background(), postulanteID, domain.Formulario{Domicilio: c.domicilio})

			campos := camposDe(t, err)
			for _, esperado := range c.campos {
				if campos[esperado] == "" {
					t.Errorf("falta el error de %q; llegaron %v", esperado, campos)
				}
			}
			if len(campos) != len(c.campos) {
				t.Errorf("campos con error = %v, se esperaban %v", campos, c.campos)
			}
			if f.tx.ejecutadas != 0 || f.store.actualizadas != 0 {
				t.Errorf("un domicilio inválido no debe abrir la transacción ni escribir (tx=%d, escrituras=%d)", f.tx.ejecutadas, f.store.actualizadas)
			}
		})
	}
}

// El formato se valida aunque el servicio no tenga catálogos, y sus errores se suman a
// los de los demás campos de la postulación.
func TestGuardar_Domicilio_FormatoSinCatalogosYJuntoAOtrosErrores(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)

	_, err := f.svc.Guardar(context.Background(), postulanteID, domain.Formulario{
		Domicilio: domain.Domicilio{TelefonoCelular: "123"},
		Cargo:     "presidente",
	})

	campos := camposDe(t, err)
	if campos["domicilio.telefono_celular"] == "" || campos["cargo"] == "" || len(campos) != 2 {
		t.Fatalf("campos = %v, se esperaban el celular y el cargo", campos)
	}
}

func TestGuardar_Domicilio_ErrorDelCatalogoNoEsDeValidacion(t *testing.T) {
	boom := errors.New("redis y postgres caídos")
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	f.svc.ConCatalogos(&fakeCatalogosDomicilio{err: boom})

	_, err := f.svc.Guardar(context.Background(), postulanteID, domain.Formulario{Domicilio: domicilioEcuador()})

	if !errors.Is(err, boom) || errors.Is(err, domain.ErrValidation) {
		t.Fatalf("se esperaba el error del catálogo sin convertirlo en 400, llegó %v", err)
	}
	if f.store.actualizadas != 0 {
		t.Fatal("no debe escribir si no pudo validar el domicilio")
	}
}

func TestGuardar_Domicilio_NoConsultaCatalogosSinPais(t *testing.T) {
	catalogos := &fakeCatalogosDomicilio{}
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	f.svc.ConCatalogos(catalogos)

	if _, err := f.svc.Guardar(context.Background(), postulanteID, domain.Formulario{}); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if catalogos.consultas != 0 {
		t.Fatalf("consultas al catálogo = %d, se esperaban 0", catalogos.consultas)
	}
}
