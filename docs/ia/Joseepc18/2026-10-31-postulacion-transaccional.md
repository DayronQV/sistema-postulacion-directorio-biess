# 2026-10-31: obtener y guardar la postulación en una transacción

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #20

## Prompt
Se pidió implementar la issue #20 (EF-12, núcleo de la postulación): `GET /postulacion` crea el borrador con código `BIESS-AAAA-NNNNNN` e información personal del Registro Civil simulado, y `PUT /postulacion` guarda el formulario en una sola transacción multi-tabla con `ROLLBACK` ante cualquier error. Con el mismo patrón de capas ya aceptado y con una prueba que demuestre el `ROLLBACK`. En una segunda instrucción, con `develop` ya trayendo el JWT y el control por rol, se pidió usarlos, agregar la secuencia del código a `db/schema.sql`, probar de punta a punta con servicios reales y abrir el PR.

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/postulacion.go` (entidades, `TxManager`, `Postulaciones` y `SeccionPostulacion`).
  - `backend/internal/repository/postgres/tx.go`, `postulacion_repo.go`, `familiares_repo.go`.
  - `backend/internal/service/postulacion_service.go`.
  - `backend/internal/http/handler/postulacion_handler.go` y `backend/internal/http/routes_postulacion.go`.
  - `backend/cmd/api/wire_postulacion.go` y una línea en `backend/cmd/api/modules.go`.
  - `db/schema.sql`: la secuencia `postulacion_codigo_seq` (autorizada para esta issue).
  - Pruebas del servicio, del handler y de integración contra PostgreSQL.
- Resumen de lo que hizo: el servicio abre la transacción con un `TxManager` (el dominio solo ve un `domain.Tx` opaco), bloquea la fila de la postulación, valida que esté en borrador, la actualiza y recorre una lista de repositorios de sección (`Reemplazar` dentro de la transacción). Familiares es la sección de referencia; las demás se agregan al cableado sin cambiar la transacción.

## Revisión humana
- `go vet`, `go build`, `go test -count=1` y `go test -race -count=1` pasan con PostgreSQL 16 y Redis 7 reales (0 pruebas saltadas). Las pruebas de `ROLLBACK` fallan si se cambia el `TxManager` para confirmar pese al error.
- Prueba de punta a punta con la API levantada: registro y login con una cédula de prueba, y luego `GET` y `PUT /postulacion` con el token. Se comprobó 401 sin token, 403 con rol de analista, código `BIESS-2026-000001` en una base nueva, 422 y cierre con la cédula extranjera, y 409 al editar una postulación enviada.
- A revisar: la secuencia en `db/schema.sql`, el orden de los familiares (la tabla no lo guarda), que `PUT` ignora las secciones aún no incluidas y el mensaje del 422.
