package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func validEnv() map[string]string {
	return map[string]string{
		"DATABASE_URL": "postgres://u:p@localhost:5432/db",
		"REDIS_ADDR":   "localhost:6379",
		"JWT_SECRET":   "secreto-de-prueba-con-mas-de-32-caracteres",
	}
}

func TestFromEnvAppliesDefaults(t *testing.T) {
	cfg, err := fromEnv(env(validEnv()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.JWTTTL != 60*time.Minute || cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("valores por defecto incorrectos: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:5173" {
		t.Fatalf("CORS por defecto incorrecto: %v", cfg.CORSAllowedOrigins)
	}
}

func TestFromEnvReadsValues(t *testing.T) {
	e := validEnv()
	e["API_PORT"] = "9000"
	e["JWT_TTL_MINUTES"] = "15"
	e["REQUEST_TIMEOUT_SECONDS"] = "3"
	e["CORS_ALLOWED_ORIGIN"] = "https://a.com/, https://www.a.com"
	cfg, err := fromEnv(env(e))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "9000" || cfg.JWTTTL != 15*time.Minute || cfg.RequestTimeout != 3*time.Second {
		t.Fatalf("valores leídos incorrectos: %+v", cfg)
	}
	if len(cfg.CORSAllowedOrigins) != 2 || cfg.CORSAllowedOrigins[0] != "https://a.com" {
		t.Fatalf("CORS incorrecto: %v", cfg.CORSAllowedOrigins)
	}
}

func TestFromEnvListsEverythingMissing(t *testing.T) {
	_, err := fromEnv(env(map[string]string{}))
	if err == nil {
		t.Fatal("se esperaba error")
	}
	for _, key := range []string{"DATABASE_URL", "REDIS_ADDR", "JWT_SECRET"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("el error no menciona %s: %v", key, err)
		}
	}
}

func TestFromEnvRejectsInvalidNumbers(t *testing.T) {
	cases := map[string]string{
		"JWT_TTL_MINUTES":         "abc",
		"REQUEST_TIMEOUT_SECONDS": "0",
		"API_PORT":                "70000",
	}
	for key, value := range cases {
		e := validEnv()
		e[key] = value
		_, err := fromEnv(env(e))
		if err == nil || !strings.Contains(err.Error(), key) {
			t.Errorf("%s=%s: se esperaba error que mencione la variable, got %v", key, value, err)
		}
	}
}

func TestFromEnvRejectsShortJWTSecret(t *testing.T) {
	e := validEnv()
	e["JWT_SECRET"] = "corto"
	_, err := fromEnv(env(e))
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") || !strings.Contains(err.Error(), "32") {
		t.Fatalf("se esperaba un error que explique el mínimo de 32 caracteres, got %v", err)
	}

	e["JWT_SECRET"] = strings.Repeat("a", 32)
	if _, err := fromEnv(env(e)); err != nil {
		t.Fatalf("un secreto de exactamente 32 caracteres debe aceptarse: %v", err)
	}
}

func TestFromEnvMailer(t *testing.T) {
	cfg, err := fromEnv(env(validEnv()))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mailer != "log" {
		t.Errorf("por defecto MAILER debe ser \"log\", got %q", cfg.Mailer)
	}

	e := validEnv()
	e["MAILER"] = "  SMTP "
	cfg, err = fromEnv(env(e))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mailer != "smtp" {
		t.Errorf("MAILER debe leerse sin espacios y en minúsculas, got %q", cfg.Mailer)
	}
}

func TestApplyDotEnv(t *testing.T) {
	// Registra la restauración de las variables para no contaminar otras pruebas.
	t.Setenv("DOTENV_KEEP", "del-entorno")
	for _, k := range []string{"DOTENV_NEW", "DOTENV_QUOTED", "DOTENV_EXPORTED"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	path := filepath.Join(t.TempDir(), ".env")
	content := "# comentario\n\nDOTENV_KEEP=del-archivo\nDOTENV_NEW = nuevo \nDOTENV_QUOTED=\"con espacios\"\nexport DOTENV_EXPORTED=ok\nlinea-sin-igual\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := applyDotEnv(path); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"DOTENV_KEEP":     "del-entorno", // el entorno tiene prioridad
		"DOTENV_NEW":      "nuevo",
		"DOTENV_QUOTED":   "con espacios",
		"DOTENV_EXPORTED": "ok",
	}
	for k, v := range want {
		if got := os.Getenv(k); got != v {
			t.Errorf("%s = %q, se esperaba %q", k, got, v)
		}
	}
}
