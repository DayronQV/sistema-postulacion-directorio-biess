# 2026-10-31: registrar correo y contraseña del postulante

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #13

## Prompt
```
si es posible trabajar ahora en el #13, adelante.
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/auth.go`: `NuevoPostulante`, `ConflictoCedula`/`ConflictoEmail`, puertos `TicketsRegistro.Cedula`/`Eliminar` y `Postulantes.ExisteEmail`/`Crear`.
  - `backend/internal/service/registro_service.go` y `_test.go`: `RegistroService.Registrar`.
  - `backend/internal/repository/postgres/postulante_repo.go`: `ExisteEmail` y `Crear` (violación UNIQUE → 409).
  - `backend/internal/repository/redis/tickets.go` e `intentos_test.go`: leer y borrar el ticket.
  - `backend/internal/http/handler/auth_handler.go` y `_test.go`, `backend/internal/http/routes_auth.go`.
  - `backend/go.mod`: `golang.org/x/crypto` pasa de indirecta a directa (bcrypt); no es una dependencia nueva.
- Resumen de lo que hizo: `POST /api/v1/auth/registro { ticket, email, password }`. Valida correo y contraseña (mínimo 8, letras y números, máximo 72 bytes por el límite de bcrypt) → 400 con `fields`; ticket vacío, inexistente o expirado → 401; cédula o correo ya registrados → 409 con `fields`; guarda la contraseña con bcrypt, borra el ticket y responde `201 { id }`.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...` y `go test ./...` en verde. Las pruebas del servicio cubren éxito (hash bcrypt verificado), ticket no reutilizable, ticket inválido, conflictos de cédula y correo y validaciones. Las pruebas del repositorio Redis se saltan si Redis no está levantado.
- Qué se corrigió: en la revisión del PR (gualter302) se detectó que el mínimo de 8 caracteres se contaba en bytes (`len`), así que `áááá1` (5 caracteres, 9 bytes) pasaba. Ahora el mínimo usa `utf8.RuneCountInString` y el máximo sigue en 72 bytes, con su mensaje corregido; se agregaron pruebas con contraseñas Unicode.
