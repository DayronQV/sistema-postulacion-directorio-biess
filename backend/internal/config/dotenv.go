package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// findDotEnv busca un archivo .env en el directorio actual y hasta tres niveles
// arriba, para que `go run ./cmd/api` funcione tanto desde la raíz del repositorio
// como desde backend/.
func findDotEnv() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for i := 0; i < 4; i++ {
		candidate := filepath.Join(dir, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// applyDotEnv carga las líneas CLAVE=valor del archivo en el entorno del proceso.
// Las variables que ya existen en el entorno tienen prioridad y no se pisan.
func applyDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, unquote(strings.TrimSpace(value))); err != nil {
			return err
		}
	}
	return sc.Err()
}

func unquote(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
