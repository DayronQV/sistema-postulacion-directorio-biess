package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// historialFalso devuelve siempre la misma respuesta y cuenta las consultas.
type historialFalso struct {
	calidad   domain.Calidad
	err       error
	consultas int
}

func (h *historialFalso) ConsultarCalidad(context.Context, string) (domain.Calidad, error) {
	h.consultas++
	return h.calidad, h.err
}

// Mensajes del requerimiento (docs/requerimiento-biess.md, 1.4.1 a 1.4.3), copiados aquí
// para que un cambio accidental en el servicio rompa la prueba.
const (
	msgNoAfiliado  = "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS"
	msgNoJubilado  = "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación"
	msgNoEmpleador = "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador"
)

type casoCargo struct {
	cargo    string
	valido   string
	noValido string
	cedula   string // cédula de prueba que cumple la calidad del cargo
}

var casosCargo = []casoCargo{
	{domain.CargoAfiliados, "Postulante afiliado al IESS", msgNoAfiliado, historiallaboral.CedulaAfiliado},
	{domain.CargoJubilados, "Postulante acredita jubilación", msgNoJubilado, historiallaboral.CedulaJubilado},
	{domain.CargoEmpleador, "Postulante es empleador", msgNoEmpleador, historiallaboral.CedulaEmpleador},
}

func formularioConCargo(cargo string) domain.Formulario {
	f := formularioValido()
	f.Cargo = cargo
	return f
}

func mensajeDe(t *testing.T, err error) string {
	t.Helper()
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("se esperaba un error de dominio, llegó %v", err)
	}
	return de.Message
}

func TestGuardar_CadaCedulaDePruebaPasaEnSuCargo(t *testing.T) {
	for _, c := range casosCargo {
		t.Run(c.cargo, func(t *testing.T) {
			f := newPostFixture(c.cedula, nil)

			p, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(c.cargo))
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if p.Estado != domain.EstadoBorrador || p.Formulario.Cargo != c.cargo {
				t.Fatalf("estado = %q, cargo = %q", p.Estado, p.Formulario.Cargo)
			}
		})
	}
}

func TestGuardar_CedulaSinCalidadSeBloqueaEnLosTresCargos(t *testing.T) {
	for _, c := range casosCargo {
		t.Run(c.cargo, func(t *testing.T) {
			f := newPostFixture(historiallaboral.CedulaNinguna, nil)

			_, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(c.cargo))
			if !errors.Is(err, domain.ErrUnprocessable) || mensajeDe(t, err) != c.noValido {
				t.Fatalf("se esperaba 422 %q, llegó %v", c.noValido, err)
			}
			if f.store.p.Estado != domain.EstadoCerrada {
				t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
			}
			if f.store.actualizadas != 0 || f.tx.rollbacks != 0 || len(*f.llamadas) != 0 {
				t.Fatalf("el cierre debe confirmarse sin escribir el formulario: actualizadas=%d rollbacks=%d llamadas=%v",
					f.store.actualizadas, f.tx.rollbacks, *f.llamadas)
			}

			// Ya cerrada: solo hay conflicto, con cualquier cargo.
			if _, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(c.cargo)); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("se esperaba 409, llegó %v", err)
			}
		})
	}
}

func TestGuardar_CedulaConOtraCalidadNoPasaEnElCargoAjeno(t *testing.T) {
	// Un jubilado no es afiliado ni empleador, y así con cada par.
	for _, c := range casosCargo {
		for _, otro := range casosCargo {
			if otro.cargo == c.cargo {
				continue
			}
			t.Run(c.cargo+" con la cédula de "+otro.cargo, func(t *testing.T) {
				f := newPostFixture(otro.cedula, nil)

				_, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(c.cargo))
				if !errors.Is(err, domain.ErrUnprocessable) || mensajeDe(t, err) != c.noValido {
					t.Fatalf("se esperaba 422 %q, llegó %v", c.noValido, err)
				}
				if f.store.p.Estado != domain.EstadoCerrada {
					t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
				}
			})
		}
	}
}

// Un cuerpo inválido no debe ocultar el cierre por cargo no elegible.
func TestGuardar_CargoNoElegibleConCuerpoInvalidoSeCierraIgual(t *testing.T) {
	f := newPostFixture(historiallaboral.CedulaNinguna, nil)
	fm := formularioConCargo(domain.CargoAfiliados)
	fm.TieneFamiliares = false // inválido: hay filas de familiares

	_, err := f.svc.Guardar(context.Background(), postulanteID, fm)
	if !errors.Is(err, domain.ErrUnprocessable) || mensajeDe(t, err) != msgNoAfiliado {
		t.Fatalf("se esperaba 422 %q, llegó %v", msgNoAfiliado, err)
	}
	if f.store.p.Estado != domain.EstadoCerrada {
		t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
	}
}

func TestGuardar_SinCargoNoConsultaElHistorialLaboral(t *testing.T) {
	f := newPostFixture(historiallaboral.CedulaNinguna, nil)
	historial := &historialFalso{}
	f.svc.ConHistorialLaboral(historial)

	if _, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo("")); err != nil {
		t.Fatalf("un borrador sin cargo debe guardarse: %v", err)
	}
	if historial.consultas != 0 {
		t.Fatalf("consultas = %d, se esperaba 0", historial.consultas)
	}
}

func TestGuardar_ConsultaElHistorialEnCadaGuardadoConCargo(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	historial := &historialFalso{calidad: domain.Calidad{EsAfiliado: true}}
	f.svc.ConHistorialLaboral(historial)

	for i := 0; i < 2; i++ {
		if _, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(domain.CargoAfiliados)); err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
	}
	if historial.consultas != 2 {
		t.Fatalf("consultas = %d, se esperaba 2", historial.consultas)
	}
}

func TestGuardar_UnCargoYaGuardadoSeRevalidaYCierra(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	historial := &historialFalso{calidad: domain.Calidad{EsAfiliado: true}}
	f.svc.ConHistorialLaboral(historial)
	if _, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(domain.CargoAfiliados)); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	historial.calidad = domain.Calidad{} // el afiliado deja de serlo
	_, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(domain.CargoAfiliados))
	if !errors.Is(err, domain.ErrUnprocessable) || f.store.p.Estado != domain.EstadoCerrada {
		t.Fatalf("err = %v, estado = %q", err, f.store.p.Estado)
	}
}

func TestGuardar_ErrorDelHistorialNoCierraNiGuarda(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	boom := errors.New("historial caído")
	f.svc.ConHistorialLaboral(&historialFalso{err: boom})

	_, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(domain.CargoAfiliados))
	if !errors.Is(err, boom) {
		t.Fatalf("se esperaba el error del historial, llegó %v", err)
	}
	if f.store.p.Estado != domain.EstadoBorrador || f.store.actualizadas != 0 || f.tx.ejecutadas != 0 {
		t.Fatalf("estado = %q, actualizadas = %d, transacciones = %d", f.store.p.Estado, f.store.actualizadas, f.tx.ejecutadas)
	}
}

func TestGuardar_SinHistorialConfiguradoNoOmiteLaRegla(t *testing.T) {
	store := &fakePostulaciones{cedula: historiallaboral.CedulaNinguna}
	llamadas := &[]string{}
	svc := service.NewPostulacionService(store, registrocivil.NewMock(), &fakeTx{store: store},
		fakeSeccion{nombre: "familiares", llamadas: llamadas})

	_, err := svc.Guardar(context.Background(), postulanteID, formularioConCargo(domain.CargoAfiliados))
	if err == nil || errors.Is(err, domain.ErrUnprocessable) || store.actualizadas != 0 {
		t.Fatalf("debía fallar con un error interno sin guardar: err = %v, actualizadas = %d", err, store.actualizadas)
	}
}

func TestGuardar_NacionalidadTienePrioridadSobreElCargo(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaExtranjera, nil)

	_, err := f.svc.Guardar(context.Background(), postulanteID, formularioConCargo(domain.CargoAfiliados))
	if !errors.Is(err, domain.ErrUnprocessable) || mensajeDe(t, err) != service.MsgNacionalidadNoEcuatoriana {
		t.Fatalf("se esperaba el mensaje de nacionalidad, llegó %v", err)
	}
}

func TestGuardar_Respaldo(t *testing.T) {
	doc := "6f1d1f6e-39c4-4d59-a8e8-0a9f4d5a8f10"
	vacio := ""
	casos := map[string]struct {
		mutar  func(*domain.Formulario)
		campos []string // vacío si debe guardarse
	}{
		"responde No": {func(f *domain.Formulario) {}, nil},
		"Sí con nombre y PDF": {func(f *domain.Formulario) {
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &doc}
		}, nil},
		"Sí sin nombre": {func(f *domain.Formulario) { f.RespaldoOrganizacion = domain.RespaldoOrganizacion{DocumentoID: &doc} }, []string{"respaldo_organizacion.nombre"}},
		"Sí con nombre en blanco": {func(f *domain.Formulario) {
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{Nombre: "   ", DocumentoID: &doc}
		}, []string{"respaldo_organizacion.nombre"}},
		"Sí sin documento": {func(f *domain.Formulario) {
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{Nombre: "Federación"}
		}, []string{"respaldo_organizacion.documento_id"}},
		"Sí con documento vacío": {func(f *domain.Formulario) {
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &vacio}
		}, []string{"respaldo_organizacion.documento_id"}},
		"Sí sin nada útil": {func(f *domain.Formulario) {
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{DocumentoID: &vacio, Nombre: " "}
		}, nil},
		"nombre muy largo": {func(f *domain.Formulario) {
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{Nombre: string(make([]rune, 201)), DocumentoID: &doc}
		}, []string{"respaldo_organizacion.nombre"}},
		"respaldo sin cargo": {func(f *domain.Formulario) {
			f.Cargo = ""
			f.RespaldoOrganizacion = domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &doc}
		}, []string{"cargo"}},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
			fm := formularioValido()
			c.mutar(&fm)

			p, err := f.svc.Guardar(context.Background(), postulanteID, fm)
			if len(c.campos) == 0 {
				if err != nil {
					t.Fatalf("error inesperado: %v", err)
				}
				if nom := p.Formulario.RespaldoOrganizacion.Nombre; nom == "   " || nom == " " {
					t.Fatalf("el nombre debe guardarse sin espacios sobrantes: %q", nom)
				}
				return
			}
			var de *domain.Error
			if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("se esperaba 400, llegó %v", err)
			}
			for _, campo := range c.campos {
				if de.Fields[campo] == "" {
					t.Fatalf("falta el campo %q en %v", campo, de.Fields)
				}
			}
			if f.tx.ejecutadas != 0 {
				t.Fatalf("una validación fallida no debe abrir transacción")
			}
		})
	}
}

func TestValidarCargo_CadaCedulaDePruebaEsElegibleEnSuCargo(t *testing.T) {
	for _, c := range casosCargo {
		t.Run(c.cargo, func(t *testing.T) {
			f := newPostFixture(c.cedula, nil)

			res, err := f.svc.ValidarCargo(context.Background(), postulanteID, c.cargo)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if !res.Elegible || res.Mensaje != c.valido || res.Cargo != c.cargo {
				t.Fatalf("resultado = %+v", res)
			}
			if f.store.p.Estado != domain.EstadoBorrador || f.tx.ejecutadas != 0 || f.store.actualizadas != 0 {
				t.Fatalf("validar un cargo elegible no escribe nada: estado = %q, transacciones = %d", f.store.p.Estado, f.tx.ejecutadas)
			}
		})
	}
}

func TestValidarCargo_CedulaSinCalidadCierraLaPostulacionEnLosTresCargos(t *testing.T) {
	for _, c := range casosCargo {
		t.Run(c.cargo, func(t *testing.T) {
			f := newPostFixture(historiallaboral.CedulaNinguna, nil)

			_, err := f.svc.ValidarCargo(context.Background(), postulanteID, c.cargo)
			if !errors.Is(err, domain.ErrUnprocessable) || mensajeDe(t, err) != c.noValido {
				t.Fatalf("se esperaba 422 %q, llegó %v", c.noValido, err)
			}
			if f.store.p.Estado != domain.EstadoCerrada {
				t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
			}

			// Ya cerrada: ni el cargo ni el guardado vuelven a abrirla.
			if _, err := f.svc.ValidarCargo(context.Background(), postulanteID, c.cargo); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("se esperaba 409, llegó %v", err)
			}
		})
	}
}

func TestValidarCargo_CargoNoValido(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)

	for _, cargo := range []string{"", "gerente", "Afiliados"} {
		_, err := f.svc.ValidarCargo(context.Background(), postulanteID, cargo)
		var de *domain.Error
		if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) || de.Fields["cargo"] == "" {
			t.Fatalf("cargo %q: se esperaba 400 con el campo cargo, llegó %v", cargo, err)
		}
	}
	if f.store.p != nil {
		t.Fatalf("un cargo inválido no debe crear ni cerrar la postulación")
	}
}

func TestValidarCargo_NacionalidadNoEcuatorianaCierra(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaExtranjera, nil)

	_, err := f.svc.ValidarCargo(context.Background(), postulanteID, domain.CargoAfiliados)
	if !errors.Is(err, domain.ErrUnprocessable) || mensajeDe(t, err) != service.MsgNacionalidadNoEcuatoriana {
		t.Fatalf("se esperaba el mensaje de nacionalidad, llegó %v", err)
	}
	if f.store.p.Estado != domain.EstadoCerrada {
		t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
	}
}

func TestValidarCargo_SoloEnBorrador(t *testing.T) {
	for _, estado := range []string{domain.EstadoEnviada, domain.EstadoCerrada} {
		t.Run(estado, func(t *testing.T) {
			f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
			if _, err := f.svc.Obtener(context.Background(), postulanteID); err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			f.store.p.Estado = estado

			if _, err := f.svc.ValidarCargo(context.Background(), postulanteID, domain.CargoAfiliados); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("se esperaba 409, llegó %v", err)
			}
		})
	}
}

func TestValidarCargo_ErrorDelHistorialNoCierra(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	boom := errors.New("historial caído")
	f.svc.ConHistorialLaboral(&historialFalso{err: boom})

	if _, err := f.svc.ValidarCargo(context.Background(), postulanteID, domain.CargoAfiliados); !errors.Is(err, boom) {
		t.Fatalf("se esperaba el error del historial, llegó %v", err)
	}
	if f.store.p.Estado != domain.EstadoBorrador {
		t.Fatalf("estado = %q, se esperaba borrador", f.store.p.Estado)
	}
}
