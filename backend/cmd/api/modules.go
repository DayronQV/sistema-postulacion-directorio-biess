package main

import (
	"log/slog"

	goredis "github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
)

// container agrupa la infraestructura compartida (configuración, log, PostgreSQL y
// Redis) que main.go entrega a cada módulo para construir sus repositorios,
// servicios y handlers.
type container struct {
	cfg    config.Config
	logger *slog.Logger
	db     *pgxpool.Pool
	redis  *goredis.Client

	// historial es el Historial Laboral del IESS (hoy el simulador); lo usa la
	// selección del cargo de la postulación.
	historial domain.HistorialLaboral
}

// modules devuelve el registro de rutas de cada módulo.
//
// Para sumar un módulo: crear su función de cableado (por ejemplo authRoutes(c) en
// wire_auth.go, que construye repositorios, servicio y handler) y agregar una línea
// a la lista. Es el único lugar de composición, junto con main.go.
//
// Dónde va cada pieza de un módulo:
//   - Interfaces de repositorio (los puertos) y entidades: internal/domain. El servicio
//     depende solo de estas interfaces, nunca de pgx ni de go-redis.
//   - Implementaciones de los repositorios: internal/repository/postgres y
//     internal/repository/redis. El cableado les entrega c.db o c.redis.
//   - Servicios (casos de uso): internal/service. Sin HTTP ni SQL.
//   - Handlers y rutas: internal/http/handler y internal/http/routes_<modulo>.go.
//
// Los handlers pasan siempre c.Request.Context() al servicio, y este a los
// repositorios: así el timeout por petición cancela las consultas.
func modules(c *container) []apihttp.RouteRegistrar {
	return []apihttp.RouteRegistrar{
		authRoutes(c),
		// catalogoRoutes(c),
		// postulacionRoutes(c),
	}
}
