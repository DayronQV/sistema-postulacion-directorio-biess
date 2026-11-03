package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/registrocivil"
)

func TestObtener_InformacionPersonalYParejaSegunElEstadoCivil(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	svc := newService(pool)

	casos := map[string]struct {
		cedula string
		estado domain.EstadoCivil
		pareja *domain.Pareja
	}{
		"casada":         {registrocivil.CedulaCasada, domain.EstadoCasado, &domain.Pareja{NombresApellidos: "Luis Alberto Párraga Mera", Cedula: "0913579249"}},
		"unión de hecho": {registrocivil.CedulaUnionDeHecho, domain.EstadoUnionDeHecho, &domain.Pareja{NombresApellidos: "Ana Lucía Moreira Zambrano", Cedula: "1313579243"}},
		"soltera":        {registrocivil.CedulaFlujoFeliz, domain.EstadoSoltero, nil},
	}
	for nombre, c := range casos {
		id := crearPostulante(t, pool, c.cedula)

		p, err := svc.Obtener(ctx, id)
		if err != nil {
			t.Fatalf("%s: error inesperado: %v", nombre, err)
		}
		if p.Personal.Cedula != c.cedula || p.Personal.Nombres == "" || p.Personal.Apellidos == "" ||
			p.Personal.Genero == "" || p.Personal.FechaNacimiento.IsZero() ||
			p.Personal.Nacionalidad != domain.NacionalidadEcuatoriana || p.Personal.EstadoCivil != c.estado {
			t.Fatalf("%s: información personal incompleta: %+v", nombre, p.Personal)
		}
		switch {
		case c.pareja == nil && p.Personal.Pareja != nil:
			t.Fatalf("%s: no debía traer pareja: %+v", nombre, p.Personal.Pareja)
		case c.pareja != nil && (p.Personal.Pareja == nil || *p.Personal.Pareja != *c.pareja):
			t.Fatalf("%s: pareja = %+v, se esperaba %+v", nombre, p.Personal.Pareja, c.pareja)
		}
	}
}

// La cédula extranjera de prueba no puede continuar llamando directamente al servicio:
// el primer guardado responde 422 y cierra la postulación; los siguientes, 409. Lo que
// llegó en el cuerpo nunca se escribe.
func TestGuardar_ExtranjeroQuedaBloqueadoParaSiempre(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaExtranjera)
	svc := newService(pool)

	// El cuerpo es inválido a propósito: no debe ocultar el cierre.
	invalido := domain.Formulario{Cargo: "gerente", TieneFamiliares: false, Familiares: []domain.Familiar{{NombresApellidos: "Ana"}}}
	if _, err := svc.Guardar(ctx, id, invalido); !errors.Is(err, domain.ErrUnprocessable) {
		t.Fatalf("primer guardado: se esperaba 422, llegó %v", err)
	}
	for _, f := range []domain.Formulario{invalido, {Cargo: domain.CargoAfiliados}, {}} {
		if _, err := svc.Guardar(ctx, id, f); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("guardado posterior: se esperaba 409, llegó %v", err)
		}
	}

	p, err := svc.Obtener(ctx, id)
	if err != nil {
		t.Fatalf("la lectura de una postulación cerrada debe funcionar: %v", err)
	}
	if p.Estado != domain.EstadoCerrada || p.Personal.Nacionalidad != "Colombiana" {
		t.Fatalf("postulación = %+v, se esperaba cerrada por nacionalidad", p)
	}
	var cargo *string
	var familiares int
	if err := pool.QueryRow(ctx, `SELECT cargo FROM postulaciones WHERE postulante_id = $1`, id).Scan(&cargo); err != nil || cargo != nil {
		t.Fatalf("no debía guardarse el formulario: cargo=%v err=%v", cargo, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM postulacion_familiares WHERE postulacion_id = $1`, p.ID).Scan(&familiares); err != nil || familiares != 0 {
		t.Fatalf("no debían guardarse familiares: %d (%v)", familiares, err)
	}
}

// El Registro Civil puede escribir el gentilicio de otra forma: la nacionalidad se compara
// normalizada, no por texto exacto.
func TestGuardar_NacionalidadEnOtraEscrituraSigueSiendoEcuatoriana(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	id := crearPostulante(t, pool, registrocivil.CedulaFlujoFeliz)
	svc := newService(pool)
	if _, err := svc.Obtener(ctx, id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE postulaciones SET nacionalidad = '  ECUATORIANA ' WHERE postulante_id = $1`, id); err != nil {
		t.Fatal(err)
	}

	p, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: domain.CargoAfiliados})
	if err != nil {
		t.Fatalf("una persona ecuatoriana no debe bloquearse: %v", err)
	}
	if p.Estado != domain.EstadoBorrador || p.Formulario.Cargo != domain.CargoAfiliados {
		t.Fatalf("postulación = %+v", p)
	}
}
