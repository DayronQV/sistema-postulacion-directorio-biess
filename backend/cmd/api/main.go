// Comando api: servicio HTTP del sistema de postulación al Directorio del BIESS.
//
// main.go es la raíz de composición: el único lugar que conoce las implementaciones
// concretas. Crea la infraestructura (PostgreSQL, Redis) y la inyecta en los módulos.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/config"
	apihttp "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/http/handler"
	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/postgres"
	redisrepo "github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/repository/redis"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("el servidor terminó con error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Se cancela con Ctrl+C o SIGTERM (docker stop) para apagar ordenadamente.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	rdb, err := redisrepo.NewClient(ctx, cfg.RedisAddr)
	if err != nil {
		return err
	}
	defer rdb.Close()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	c := &container{cfg: cfg, logger: logger, db: pool, redis: rdb}
	router := apihttp.NewRouter(apihttp.Deps{
		Config:  cfg,
		Logger:  logger,
		Health:  handler.NewHealthHandler(),
		Modules: modules(c),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("API escuchando", "addr", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
		logger.Info("apagando el servidor")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
