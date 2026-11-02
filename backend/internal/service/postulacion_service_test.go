package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

const postulanteID = "p-1"

// fakePostulaciones guarda una sola postulación en memoria. Las escrituras dentro de la
// transacción pasan por fakeTx, que las deshace si la función falla.
type fakePostulaciones struct {
	cedula       string
	p            *domain.Postulacion
	creadas      int
	anioUsado    int
	actualizadas int
}

func (f *fakePostulaciones) CedulaDePostulante(context.Context, string) (string, error) {
	return f.cedula, nil
}

func (f *fakePostulaciones) Crear(_ context.Context, id string, per domain.InformacionPersonal, anio int) (bool, error) {
	f.creadas++
	f.anioUsado = anio
	f.p = &domain.Postulacion{ID: "post-1", Codigo: "BIESS-2026-000001", PostulanteID: id, Estado: domain.EstadoBorrador, Personal: per}
	return true, nil
}

func (f *fakePostulaciones) ObtenerPorPostulante(context.Context, string) (domain.Postulacion, error) {
	if f.p == nil {
		return domain.Postulacion{}, domain.NotFound("no existe")
	}
	return *f.p, nil
}

func (f *fakePostulaciones) BloquearPorPostulante(ctx context.Context, _ domain.Tx, id string) (domain.Postulacion, error) {
	return f.ObtenerPorPostulante(ctx, id)
}

func (f *fakePostulaciones) Actualizar(_ context.Context, _ domain.Tx, _ string, fm domain.Formulario) error {
	f.actualizadas++
	f.p.Formulario = fm
	return nil
}

func (f *fakePostulaciones) Cerrar(context.Context, domain.Tx, string) error {
	f.p.Estado = domain.EstadoCerrada
	return nil
}

// fakeTx imita la transacción: si fn falla, restaura la postulación al estado previo.
type fakeTx struct {
	store      *fakePostulaciones
	ejecutadas int
	rollbacks  int
}

func (t *fakeTx) Ejecutar(ctx context.Context, fn func(context.Context, domain.Tx) error) error {
	t.ejecutadas++
	var antes domain.Postulacion
	if t.store.p != nil {
		antes = *t.store.p
	}
	if err := fn(ctx, struct{}{}); err != nil {
		t.rollbacks++
		if t.store.p != nil {
			*t.store.p = antes
		}
		return err
	}
	return nil
}

// fakeSeccion registra las llamadas y puede fallar al reemplazar.
type fakeSeccion struct {
	nombre   string
	err      error
	llamadas *[]string
}

func (s fakeSeccion) Cargar(context.Context, string, *domain.Formulario) error {
	*s.llamadas = append(*s.llamadas, "cargar:"+s.nombre)
	return nil
}

func (s fakeSeccion) Reemplazar(context.Context, domain.Tx, string, domain.Formulario) error {
	*s.llamadas = append(*s.llamadas, "reemplazar:"+s.nombre)
	return s.err
}

type postFixture struct {
	svc      *service.PostulacionService
	store    *fakePostulaciones
	tx       *fakeTx
	llamadas *[]string
}

func newPostFixture(cedula string, errSegunda error) postFixture {
	store := &fakePostulaciones{cedula: cedula}
	tx := &fakeTx{store: store}
	llamadas := &[]string{}
	svc := service.NewPostulacionService(store, registrocivil.NewMock(), tx,
		fakeSeccion{nombre: "familiares", llamadas: llamadas},
		fakeSeccion{nombre: "formacion", err: errSegunda, llamadas: llamadas},
	).ConHistorialLaboral(historiallaboral.NewMock())
	return postFixture{svc: svc, store: store, tx: tx, llamadas: llamadas}
}

func TestObtener_CreaBorradorConDatosDelRegistroCivil(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaCasada, nil)

	p, err := f.svc.Obtener(context.Background(), postulanteID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Estado != domain.EstadoBorrador || p.Codigo == "" {
		t.Fatalf("postulación = %+v", p)
	}
	if p.Personal.Nombres == "" || p.Personal.Nacionalidad != domain.NacionalidadEcuatoriana {
		t.Fatalf("no se copió la información personal: %+v", p.Personal)
	}
	if p.Personal.Pareja == nil {
		t.Fatalf("la cédula casada debe copiar los datos de la pareja")
	}
	if f.store.anioUsado != time.Now().Year() {
		t.Fatalf("año del código = %d", f.store.anioUsado)
	}

	// La segunda lectura no vuelve a crearla y carga las secciones.
	if _, err := f.svc.Obtener(context.Background(), postulanteID); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if f.store.creadas != 1 {
		t.Fatalf("se creó %d veces, se esperaba 1", f.store.creadas)
	}
	if got := len(*f.llamadas); got != 4 {
		t.Fatalf("Cargar debía llamarse en las dos secciones por lectura, llamadas = %v", *f.llamadas)
	}
}

func TestObtener_SolteroNoCopiaPareja(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	p, err := f.svc.Obtener(context.Background(), postulanteID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Personal.Pareja != nil {
		t.Fatalf("una persona soltera no tiene pareja: %+v", p.Personal.Pareja)
	}
}

func formularioValido() domain.Formulario {
	return domain.Formulario{
		Cargo:           domain.CargoAfiliados,
		TieneFamiliares: true,
		Familiares: []domain.Familiar{
			{NombresApellidos: "Ana Pérez", Parentesco: "hermano/a", Institucion: "IESS", Area: "Finanzas", Cargo: "Analista"},
		},
		Consentimiento: true,
	}
}

func TestGuardar_EscribeLaPostulacionYCadaSeccionEnUnaTransaccion(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)

	p, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if f.tx.ejecutadas != 1 || f.tx.rollbacks != 0 {
		t.Fatalf("transacciones = %d, rollbacks = %d", f.tx.ejecutadas, f.tx.rollbacks)
	}
	if p.Formulario.Cargo != domain.CargoAfiliados || !p.Formulario.Consentimiento {
		t.Fatalf("no se guardó el formulario: %+v", p.Formulario)
	}
	want := []string{"reemplazar:familiares", "reemplazar:formacion", "cargar:familiares", "cargar:formacion"}
	if got := *f.llamadas; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("llamadas = %v, se esperaba %v", got, want)
	}
}

func TestGuardar_ErrorEnUnaSeccionDeshaceTodo(t *testing.T) {
	boom := errors.New("falla la segunda tabla")
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, boom)
	if _, err := f.svc.Obtener(context.Background(), postulanteID); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	*f.llamadas = nil

	_, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido())
	if !errors.Is(err, boom) {
		t.Fatalf("se esperaba el error de la sección, llegó %v", err)
	}
	if f.tx.rollbacks != 1 {
		t.Fatalf("rollbacks = %d, se esperaba 1", f.tx.rollbacks)
	}
	if f.store.p.Formulario.Cargo != "" || f.store.p.Formulario.Consentimiento {
		t.Fatalf("quedó una escritura parcial: %+v", f.store.p.Formulario)
	}
}

func TestGuardar_SoloEnBorrador(t *testing.T) {
	for _, estado := range []string{domain.EstadoEnviada, domain.EstadoCerrada} {
		t.Run(estado, func(t *testing.T) {
			f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
			if _, err := f.svc.Obtener(context.Background(), postulanteID); err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			f.store.p.Estado = estado
			*f.llamadas = nil

			_, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido())
			if !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("se esperaba 409, llegó %v", err)
			}
			if f.store.actualizadas != 0 || len(*f.llamadas) != 0 {
				t.Fatalf("no debía escribir nada: actualizadas=%d llamadas=%v", f.store.actualizadas, *f.llamadas)
			}
		})
	}
}

func TestGuardar_NacionalidadNoEcuatorianaCierraLaPostulacion(t *testing.T) {
	f := newPostFixture(registrocivil.CedulaExtranjera, nil)

	_, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido())
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrUnprocessable) || de.Message != service.MsgNacionalidadNoEcuatoriana {
		t.Fatalf("se esperaba 422 %q, llegó %v", service.MsgNacionalidadNoEcuatoriana, err)
	}
	if f.store.p.Estado != domain.EstadoCerrada {
		t.Fatalf("estado = %q, se esperaba cerrada", f.store.p.Estado)
	}
	if f.tx.rollbacks != 0 {
		t.Fatalf("el cierre debe confirmarse, no deshacerse")
	}
	if f.store.actualizadas != 0 {
		t.Fatalf("no debía guardar el formulario")
	}

	// Ya cerrada: ahora es un conflicto.
	if _, err := f.svc.Guardar(context.Background(), postulanteID, formularioValido()); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("se esperaba 409, llegó %v", err)
	}
}

func TestGuardar_Validacion(t *testing.T) {
	casos := map[string]struct {
		mutar func(*domain.Formulario)
		campo string
	}{
		"cargo no válido":         {func(f *domain.Formulario) { f.Cargo = "gerente" }, "cargo"},
		"familiares sin declarar": {func(f *domain.Formulario) { f.TieneFamiliares = false }, "familiares"},
		"institución no válida":   {func(f *domain.Formulario) { f.Familiares[0].Institucion = "ISSFA" }, "familiares[0].institucion"},
		"familiar sin nombre":     {func(f *domain.Formulario) { f.Familiares[0].NombresApellidos = " " }, "familiares[0].nombres_apellidos"},
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
