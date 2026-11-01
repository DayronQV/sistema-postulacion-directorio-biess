# 2026-10-31: simulador del Historial Laboral del IESS

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #11

## Prompt
Se pidió implementar la issue #11: la interfaz `HistorialLaboral` en `internal/domain` y su adaptador `mock` en `internal/integration/historiallaboral`, con las mismas cédulas de prueba del simulador del Registro Civil (afiliado, jubilado, empleador, ninguna condición y el resto como afiliado), con pruebas, siguiendo el estilo del simulador del Registro Civil y sin cambiar `docs/api.md` ni `db/schema.sql`.

## Resultado
- Archivos generados o modificados: `backend/internal/domain/historiallaboral.go`, `backend/internal/integration/historiallaboral/mock.go`, `backend/internal/integration/historiallaboral/mock_test.go`, `backend/cmd/api/main.go`, `backend/cmd/api/modules.go` y este registro.
- Resumen de lo que hizo: definió `Calidad` y el puerto `HistorialLaboral`, escribió el mock con las cuatro cédulas fijas (cualquier otra se trata como afiliada), sus pruebas, e inyectó el mock en el contenedor de `main.go`.

## Revisión humana
- Se ejecutaron `go vet ./...`, `go build ./...`, `go test ./...` y `go test -race -count=1 ./...` desde `backend/`, todos en verde.
- Una prueba comprueba que las cédulas del mock coinciden con las del Registro Civil simulado.
- El mock no consulta ninguna base de datos ni servicio externo, por lo que no se probó con Docker.
