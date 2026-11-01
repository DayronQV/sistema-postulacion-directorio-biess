package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/integration/historiallaboral"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/service"
)

// Estas pruebas usan las cédulas de prueba del Historial Laboral simulado, que también
// existen en el Registro Civil simulado: afiliado, jubilado, empleador y sin calidad.

var cargos = []struct {
	cargo, cedula, valido, noValido string
}{
	{domain.CargoAfiliados, historiallaboral.CedulaAfiliado, service.MsgCargoAfiliadosValido, service.MsgCargoAfiliadosNoElegible},
	{domain.CargoJubilados, historiallaboral.CedulaJubilado, service.MsgCargoJubiladosValido, service.MsgCargoJubiladosNoElegible},
	{domain.CargoEmpleador, historiallaboral.CedulaEmpleador, service.MsgCargoEmpleadorValido, service.MsgCargoEmpleadorNoElegible},
}

// estadoYCargo lee de la base el estado y el cargo guardados de un postulante.
func estadoYCargo(t *testing.T, pool *pgxpool.Pool, postulanteID string) (string, *string) {
	t.Helper()
	var estado string
	var cargo *string
	err := pool.QueryRow(context.Background(),
		`SELECT estado, cargo FROM postulaciones WHERE postulante_id = $1`, postulanteID).Scan(&estado, &cargo)
	if err != nil {
		t.Fatalf("leyendo la postulación: %v", err)
	}
	return estado, cargo
}

func mensajeDeError(t *testing.T, err error) string {
	t.Helper()
	var de *domain.Error
	if !errors.As(err, &de) {
		t.Fatalf("se esperaba un error de dominio, llegó %v", err)
	}
	return de.Message
}

func TestCargo_CadaCedulaDePruebaPasaEnSuCargo(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	for _, c := range cargos {
		t.Run(c.cargo, func(t *testing.T) {
			id := crearPostulante(t, pool, c.cedula)
			svc := newService(pool)

			res, err := svc.ValidarCargo(ctx, id, c.cargo)
			if err != nil || !res.Elegible || res.Mensaje != c.valido {
				t.Fatalf("validar = %+v, %v", res, err)
			}
			if estado, cargo := estadoYCargo(t, pool, id); estado != domain.EstadoBorrador || cargo != nil {
				t.Fatalf("validar no debe escribir: estado=%q cargo=%v", estado, cargo)
			}

			p, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: c.cargo})
			if err != nil {
				t.Fatalf("guardar: %v", err)
			}
			if estado, cargo := estadoYCargo(t, pool, id); estado != domain.EstadoBorrador || cargo == nil || *cargo != c.cargo || p.Formulario.Cargo != c.cargo {
				t.Fatalf("estado=%q cargo=%v", estado, cargo)
			}
		})
	}
}

func TestCargo_CedulaSinCalidadSeBloqueaEnLosTresCargosAlGuardar(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	for _, c := range cargos {
		t.Run(c.cargo, func(t *testing.T) {
			id := crearPostulante(t, pool, historiallaboral.CedulaNinguna)
			svc := newService(pool)

			_, err := svc.Guardar(ctx, id, domain.Formulario{
				Cargo: c.cargo, TieneFamiliares: true, Familiares: []domain.Familiar{familiar("Ana")},
			})
			if !errors.Is(err, domain.ErrUnprocessable) || mensajeDeError(t, err) != c.noValido {
				t.Fatalf("se esperaba 422 %q, llegó %v", c.noValido, err)
			}
			estado, cargo := estadoYCargo(t, pool, id)
			if estado != domain.EstadoCerrada || cargo != nil {
				t.Fatalf("estado=%q cargo=%v: debe quedar cerrada sin guardar el cargo", estado, cargo)
			}
			if got := nombresFamiliares(t, pool, id); len(got) != 0 {
				t.Fatalf("el cierre no debe guardar el formulario: familiares=%v", got)
			}
			if _, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: c.cargo}); !errors.Is(err, domain.ErrConflict) {
				t.Fatalf("una postulación cerrada no se edita: %v", err)
			}
		})
	}
}

func TestCargo_ValidarSinCalidadCierraLaPostulacion(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	for _, c := range cargos {
		t.Run(c.cargo, func(t *testing.T) {
			id := crearPostulante(t, pool, historiallaboral.CedulaNinguna)
			svc := newService(pool)

			_, err := svc.ValidarCargo(ctx, id, c.cargo)
			if !errors.Is(err, domain.ErrUnprocessable) || mensajeDeError(t, err) != c.noValido {
				t.Fatalf("se esperaba 422 %q, llegó %v", c.noValido, err)
			}
			if estado, _ := estadoYCargo(t, pool, id); estado != domain.EstadoCerrada {
				t.Fatalf("estado = %q, se esperaba cerrada", estado)
			}
		})
	}
}

func TestCargo_NoPasaConLaCedulaDeOtroCargo(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	for _, c := range cargos {
		for _, otro := range cargos {
			if otro.cargo == c.cargo {
				continue
			}
			t.Run(c.cargo+" con la cédula de "+otro.cargo, func(t *testing.T) {
				id := crearPostulante(t, pool, otro.cedula)
				svc := newService(pool)

				_, err := svc.Guardar(ctx, id, domain.Formulario{Cargo: c.cargo})
				if !errors.Is(err, domain.ErrUnprocessable) || mensajeDeError(t, err) != c.noValido {
					t.Fatalf("se esperaba 422 %q, llegó %v", c.noValido, err)
				}
				if estado, cargo := estadoYCargo(t, pool, id); estado != domain.EstadoCerrada || cargo != nil {
					t.Fatalf("estado=%q cargo=%v", estado, cargo)
				}
			})
		}
	}
}

func TestCargo_RespaldoDeLaOrganizacion(t *testing.T) {
	pool := newPool(t)
	ctx := context.Background()
	idA := crearPostulante(t, pool, historiallaboral.CedulaAfiliado)
	idB := crearPostulante(t, pool, historiallaboral.CedulaJubilado)
	svc := newService(pool)
	documentos := service.NewDocumentoService(postgres.NewDocumentoRepo(pool))
	for _, id := range []string{idA, idB} {
		if _, err := svc.Obtener(ctx, id); err != nil {
			t.Fatal(err)
		}
	}
	propio, err := documentos.Subir(ctx, idA, domain.DocumentoOrganizacion, "respaldo.pdf", []byte("%PDF-1.7 respaldo"))
	if err != nil {
		t.Fatalf("subiendo el respaldo: %v", err)
	}
	ajeno, err := documentos.Subir(ctx, idB, domain.DocumentoOrganizacion, "ajeno.pdf", []byte("%PDF-1.7 ajeno"))
	if err != nil {
		t.Fatalf("subiendo el respaldo ajeno: %v", err)
	}
	otroTipo, err := documentos.Subir(ctx, idA, domain.DocumentoFormacion, "titulo.pdf", []byte("%PDF-1.7 titulo"))
	if err != nil {
		t.Fatalf("subiendo el título: %v", err)
	}

	t.Run("con nombre y documento propio se guarda", func(t *testing.T) {
		p, err := svc.Guardar(ctx, idA, domain.Formulario{
			Cargo:                domain.CargoAfiliados,
			RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: "  Federación Nacional  ", DocumentoID: &propio.ID},
		})
		if err != nil {
			t.Fatalf("guardar: %v", err)
		}
		r := p.Formulario.RespaldoOrganizacion
		if r.Nombre != "Federación Nacional" || r.DocumentoID == nil || *r.DocumentoID != propio.ID {
			t.Fatalf("respaldo guardado = %+v", r)
		}
	})

	t.Run("responder No lo borra", func(t *testing.T) {
		p, err := svc.Guardar(ctx, idA, domain.Formulario{Cargo: domain.CargoAfiliados})
		if err != nil {
			t.Fatalf("guardar: %v", err)
		}
		if r := p.Formulario.RespaldoOrganizacion; r.Nombre != "" || r.DocumentoID != nil {
			t.Fatalf("respaldo guardado = %+v, se esperaba vacío", r)
		}
	})

	t.Run("el documento de otra postulación responde 400", func(t *testing.T) {
		_, err := svc.Guardar(ctx, idA, domain.Formulario{
			Cargo:                domain.CargoAfiliados,
			RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &ajeno.ID},
		})
		var de *domain.Error
		if !errors.As(err, &de) || !errors.Is(err, domain.ErrValidation) || de.Fields["respaldo_organizacion.documento_id"] == "" {
			t.Fatalf("se esperaba 400 en respaldo_organizacion.documento_id, llegó %v", err)
		}
	})

	t.Run("un documento de otro tipo responde 400", func(t *testing.T) {
		_, err := svc.Guardar(ctx, idA, domain.Formulario{
			Cargo:                domain.CargoAfiliados,
			RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &otroTipo.ID},
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("se esperaba 400, llegó %v", err)
		}
	})

	t.Run("un identificador mal formado responde 400", func(t *testing.T) {
		malo := "no-es-un-uuid"
		_, err := svc.Guardar(ctx, idA, domain.Formulario{
			Cargo:                domain.CargoAfiliados,
			RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &malo},
		})
		if !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("se esperaba 400, llegó %v", err)
		}
	})

	t.Run("el nombre o el documento solos responden 400", func(t *testing.T) {
		for nombre, r := range map[string]domain.RespaldoOrganizacion{
			"solo nombre":    {Nombre: "Federación"},
			"solo documento": {DocumentoID: &propio.ID},
		} {
			_, err := svc.Guardar(ctx, idA, domain.Formulario{Cargo: domain.CargoAfiliados, RespaldoOrganizacion: r})
			if !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("%s: se esperaba 400, llegó %v", nombre, err)
			}
		}
	})

	t.Run("el documento referenciado ya no se puede eliminar", func(t *testing.T) {
		if _, err := svc.Guardar(ctx, idA, domain.Formulario{
			Cargo:                domain.CargoAfiliados,
			RespaldoOrganizacion: domain.RespaldoOrganizacion{Nombre: "Federación", DocumentoID: &propio.ID},
		}); err != nil {
			t.Fatalf("guardar: %v", err)
		}
		if err := documentos.Eliminar(ctx, idA, propio.ID); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("se esperaba 409, llegó %v", err)
		}
	})
}
