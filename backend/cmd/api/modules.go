package main

import (
	"log/slog"

	goredis "github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
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
}

// modules devuelve el registro de rutas de cada módulo.
//
// Para sumar un módulo: crear su función de cableado (por ejemplo authRoutes(c) en
// wire_auth.go, que construye repositorios, servicio y handler) y agregar una línea
// a la lista. Es el único lugar de composición, junto con main.go.
func modules(_ *container) []apihttp.RouteRegistrar {
	return []apihttp.RouteRegistrar{
		// authRoutes(c),
		// catalogoRoutes(c),
		// postulacionRoutes(c),
	}
}
