# 2026-11-02: revisión del PR #101 (banco de preguntas y configuración de la evaluación)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #42 (revisión del PR #101)

## Prompt
```
revisa el pr 101
```

## Resultado
- Archivos generados o modificados: `backend/internal/service/evaluacion_csv.go` y `evaluacion_csv_test.go`, `docs/api.md`.
- Resumen de lo que hizo:
  - Revisó el PR contra la issue #42: carga individual y en bloque (JSON y CSV, todo o nada), listado, edición y eliminación solo para el administrador, configuración que no acepta más preguntas que el banco y respuesta correcta fuera de lo que ve el postulante.
  - Encontró que un CSV guardado por Excel en Windows-1252 (con ¿, á o ñ) pasaba la validación y PostgreSQL rechazaba el INSERT con un 500. Ahora se valida que el archivo sea UTF-8 y se responde 400 con la indicación de guardarlo como "CSV UTF-8". Se documentó en `docs/api.md`.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go build`, `go vet`, `go test ./internal/...` y formato con `gofmt`; una prueba de sonda confirmó el problema antes de corregirlo. Las pruebas contra PostgreSQL las corre el CI.
- Qué se corrigió: lo listado arriba.
- Pendiente:
