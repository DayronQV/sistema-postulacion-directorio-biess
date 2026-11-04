package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

// mailerIntermitente es el fakeMailer que además puede fallar a pedido.
type mailerIntermitente struct {
	*fakeMailer
	err error
}

func (m *mailerIntermitente) Enviar(ctx context.Context, msg domain.Mensaje) error {
	if m.err != nil {
		return m.err
	}
	return m.fakeMailer.Enviar(ctx, msg)
}

// formularioCompleto tiene todas las secciones obligatorias y las 13 respuestas.
func formularioCompleto() domain.Formulario {
	pais, canton := 60, 1701
	respuestas := map[int]bool{}
	for n := 1; n <= domain.TotalPreguntasInhabilidades; n++ {
		respuestas[n] = n == 1
	}
	return domain.Formulario{
		Domicilio:      domain.Domicilio{PaisID: &pais, CantonID: &canton, TelefonoCelular: "0999999999"},
		Cargo:          domain.CargoAfiliados,
		Formacion:      []domain.FormacionAcademica{{Titulo: "Economista"}},
		Experiencia:    []domain.ExperienciaProfesional{{Cargo: "Analista"}},
		Capacitaciones: []domain.Capacitacion{{Nombre: "Curso"}},
		Referencias:    []domain.ReferenciaPersonal{{NombresApellidos: "Ana"}},
		Inhabilidades:  domain.Inhabilidades{Respuestas: respuestas},
		Elegible:       true,
		Consentimiento: true,
	}
}

// fixtureEnvio crea la postulación y le pone el formulario dado y un mailer falso.
func fixtureEnvio(t *testing.T, formulario domain.Formulario) (postFixture, *mailerIntermitente) {
	t.Helper()
	f := newPostFixture(registrocivil.CedulaFlujoFeliz, nil)
	mailer := &mailerIntermitente{fakeMailer: &fakeMailer{}}
	f.svc.ConMailer(mailer)
	if _, err := f.svc.Obtener(context.Background(), postulanteID); err != nil {
		t.Fatal(err)
	}
	f.store.p.Personal.Email = "ana@correo.com"
	f.store.p.Formulario = formulario
	return f, mailer
}

func TestEnviar_PostulacionCompletaQuedaEnviadaYMandaElCorreo(t *testing.T) {
	f, mailer := fixtureEnvio(t, formularioCompleto())

	p, err := f.svc.Enviar(context.Background(), postulanteID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.Estado != domain.EstadoEnviada || f.store.enviadas != 1 {
		t.Fatalf("estado = %s, enviadas = %d", p.Estado, f.store.enviadas)
	}
	if len(mailer.enviados) != 1 {
		t.Fatalf("correos = %d", len(mailer.enviados))
	}
	msg := mailer.enviados[0]
	if msg.Para != "ana@correo.com" || !strings.Contains(msg.Asunto, p.Codigo) || !strings.Contains(msg.Texto, p.Codigo) {
		t.Fatalf("correo = %+v", msg)
	}
	if err := msg.Validar(); err != nil {
		t.Fatalf("el correo debe ser válido para el mailer: %v", err)
	}
}

func TestEnviar_NoElegibleTambienSeEnvia(t *testing.T) {
	formulario := formularioCompleto()
	formulario.Inhabilidades.Respuestas[2] = true
	formulario.Elegible = false
	f, _ := fixtureEnvio(t, formulario)

	if _, err := f.svc.Enviar(context.Background(), postulanteID); err != nil {
		t.Fatalf("la elegibilidad no bloquea el envío: %v", err)
	}
}

func TestEnviar_IncompletaResponde422ConLasSeccionesQueFaltan(t *testing.T) {
	f, mailer := fixtureEnvio(t, domain.Formulario{})

	_, err := f.svc.Enviar(context.Background(), postulanteID)
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("error = %#v; se esperaba 422", err)
	}
	for _, campo := range []string{"domicilio", "cargo", "formacion", "experiencia", "capacitaciones", "referencias", "inhabilidades", "consentimiento"} {
		if de.Fields[campo] == "" {
			t.Errorf("falta %s en fields: %v", campo, de.Fields)
		}
	}
	if f.store.p.Estado != domain.EstadoBorrador || len(mailer.enviados) != 0 {
		t.Fatalf("no debe cambiar nada: estado = %s, correos = %d", f.store.p.Estado, len(mailer.enviados))
	}
}

func TestEnviar_SinConsentimientoNoSeEnvia(t *testing.T) {
	formulario := formularioCompleto()
	formulario.Consentimiento = false
	f, _ := fixtureEnvio(t, formulario)

	_, err := f.svc.Enviar(context.Background(), postulanteID)
	var de *domain.Error
	if !errors.As(err, &de) || len(de.Fields) != 1 || de.Fields["consentimiento"] == "" {
		t.Fatalf("error = %#v", err)
	}
}

func TestEnviar_NoSePuedeEnviarDosVeces(t *testing.T) {
	f, mailer := fixtureEnvio(t, formularioCompleto())
	if _, err := f.svc.Enviar(context.Background(), postulanteID); err != nil {
		t.Fatal(err)
	}

	_, err := f.svc.Enviar(context.Background(), postulanteID)
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v; se esperaba 409", err)
	}
	if len(mailer.enviados) != 1 || f.store.enviadas != 1 {
		t.Fatalf("correos = %d, enviadas = %d", len(mailer.enviados), f.store.enviadas)
	}
}

func TestEnviar_PostulacionCerradaResponde409(t *testing.T) {
	f, _ := fixtureEnvio(t, formularioCompleto())
	f.store.p.Estado = domain.EstadoCerrada

	if _, err := f.svc.Enviar(context.Background(), postulanteID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("error = %v; se esperaba 409", err)
	}
}

func TestEnviar_SiFallaElCorreoSeDeshaceYPuedeReintentarse(t *testing.T) {
	f, mailer := fixtureEnvio(t, formularioCompleto())
	mailer.err = errors.New("sin conexión")

	if _, err := f.svc.Enviar(context.Background(), postulanteID); err == nil {
		t.Fatal("se esperaba el error del correo")
	}
	if f.store.p.Estado != domain.EstadoBorrador || f.tx.rollbacks != 1 {
		t.Fatalf("estado = %s, rollbacks = %d", f.store.p.Estado, f.tx.rollbacks)
	}

	mailer.err = nil
	if p, err := f.svc.Enviar(context.Background(), postulanteID); err != nil || p.Estado != domain.EstadoEnviada {
		t.Fatalf("reintento: %v, %s", err, p.Estado)
	}
}

func TestEnviar_SiSinFilasEnFamiliaresOPublicacionesEstaIncompleta(t *testing.T) {
	formulario := formularioCompleto()
	formulario.TieneFamiliares = true
	formulario.TienePublicaciones = true
	f, mailer := fixtureEnvio(t, formulario)

	_, err := f.svc.Enviar(context.Background(), postulanteID)
	var de *domain.Error
	if !errors.As(err, &de) || !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("error = %#v; se esperaba 422", err)
	}
	if de.Fields["familiares"] == "" || de.Fields["publicaciones"] == "" {
		t.Fatalf("fields = %v; se esperaban familiares y publicaciones", de.Fields)
	}
	if len(mailer.enviados) != 0 {
		t.Fatal("no debe enviarse el correo")
	}
}

func TestEnviar_ElCorreoAgradeceYExplicaElProximoPaso(t *testing.T) {
	f, mailer := fixtureEnvio(t, formularioCompleto())

	p, err := f.svc.Enviar(context.Background(), postulanteID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	msg := mailer.enviados[0]
	if !strings.HasPrefix(msg.Texto, "Hola Carlos Andrés,") {
		t.Errorf("el correo debe saludar por los nombres del postulante: %q", msg.Texto)
	}
	for _, esperado := range []string{
		"Muchas gracias por postular", "Su código único es: " + p.Codigo,
		"imprima el formulario de postulación", "fírmelo y entréguelo", "Le deseamos muchos éxitos",
	} {
		if !strings.Contains(msg.Texto, esperado) {
			t.Errorf("el correo no incluye %q", esperado)
		}
	}
	if !strings.Contains(msg.Asunto, "Gracias por postular") || !strings.Contains(msg.Asunto, p.Codigo) {
		t.Errorf("asunto = %q", msg.Asunto)
	}
	// Es un mensaje corto: el formulario completo se imprime desde el sistema.
	for _, detalle := range []string{"INHABILIDADES", "FORMACIÓN ACADÉMICA", "Cédula", "Mendoza"} {
		if strings.Contains(msg.Texto, detalle) {
			t.Errorf("el correo no debe traer el detalle de la postulación (%q)", detalle)
		}
	}
}
