// Package config lee la configuración del backend desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DayronQV/sistema-postulacion-directorio-biess/backend/internal/domain"
)

// Config agrupa toda la configuración del servicio.
type Config struct {
	Port               string        // API_PORT
	DatabaseURL        string        // DATABASE_URL (obligatoria)
	RedisAddr          string        // REDIS_ADDR (obligatoria)
	JWTSecret          string        // JWT_SECRET (obligatoria)
	JWTTTL             time.Duration // JWT_TTL_MINUTES
	RequestTimeout     time.Duration // REQUEST_TIMEOUT_SECONDS
	CORSAllowedOrigins []string      // CORS_ALLOWED_ORIGIN (uno o varios, separados por coma)
	Mailer             string        // MAILER: adaptador de correo ("log" por defecto)
	PadronArchivo      string        // PADRON_ARCHIVO: personas de los simuladores (vacío = ciudadanos de prueba)
}

// Load carga el archivo .env si existe (sin pisar variables ya definidas en el
// entorno) y luego lee la configuración. Devuelve un error que lista todo lo que
// falta o es inválido, no solo el primer problema.
func Load() (Config, error) {
	if path := findDotEnv(); path != "" {
		if err := applyDotEnv(path); err != nil {
			return Config{}, fmt.Errorf("leyendo %s: %w", path, err)
		}
	}
	return fromEnv(os.Getenv)
}

func fromEnv(get func(string) string) (Config, error) {
	var missing, invalid []string

	required := func(key string) string {
		v := strings.TrimSpace(get(key))
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}
	positiveInt := func(key string, def int) int {
		raw := strings.TrimSpace(get(key))
		if raw == "" {
			return def
		}
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			invalid = append(invalid, key+" debe ser un entero positivo (valor actual: "+raw+")")
			return def
		}
		return n
	}

	cfg := Config{
		DatabaseURL: required("DATABASE_URL"),
		RedisAddr:   required("REDIS_ADDR"),
		JWTSecret:   required("JWT_SECRET"),
	}

	if cfg.JWTSecret != "" && len(cfg.JWTSecret) < domain.MinJWTSecretLength {
		invalid = append(invalid, fmt.Sprintf("JWT_SECRET debe tener al menos %d caracteres (tiene %d); genere uno con: openssl rand -hex 32",
			domain.MinJWTSecretLength, len(cfg.JWTSecret)))
	}

	cfg.Port = strings.TrimSpace(get("API_PORT"))
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if n, err := strconv.Atoi(cfg.Port); err != nil || n < 1 || n > 65535 {
		invalid = append(invalid, "API_PORT debe ser un puerto entre 1 y 65535 (valor actual: "+cfg.Port+")")
	}

	cfg.JWTTTL = time.Duration(positiveInt("JWT_TTL_MINUTES", 60)) * time.Minute
	cfg.RequestTimeout = time.Duration(positiveInt("REQUEST_TIMEOUT_SECONDS", 10)) * time.Second

	cfg.Mailer = strings.ToLower(strings.TrimSpace(get("MAILER")))
	if cfg.Mailer == "" {
		cfg.Mailer = "log"
	}

	cfg.PadronArchivo = strings.TrimSpace(get("PADRON_ARCHIVO"))

	origins := strings.TrimSpace(get("CORS_ALLOWED_ORIGIN"))
	if origins == "" {
		origins = "http://localhost:5173"
	}
	for _, o := range strings.Split(origins, ",") {
		if o = strings.TrimRight(strings.TrimSpace(o), "/"); o != "" {
			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, o)
		}
	}

	var problems []string
	if len(missing) > 0 {
		problems = append(problems, "faltan variables obligatorias: "+strings.Join(missing, ", "))
	}
	problems = append(problems, invalid...)
	if len(problems) > 0 {
		return Config{}, errors.New("configuración inválida: " + strings.Join(problems, "; ") +
			" (copie .env.example a .env y complételo)")
	}
	return cfg, nil
}
