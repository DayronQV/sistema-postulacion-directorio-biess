# 2026-10-31: validar identidad para el registro con bloqueo por intentos

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #12

## Prompt
```
Revisa el proyecto (todos los .md y el README), genera un plan de implementación
para mis issues (Cranell237), indica cuál va primero y de quién dependo. Las ramas se
crean por issue y siempre se traen los cambios de develop antes y después.

(después) Sí, empezar #12
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/auth.go`: puertos `Intentos`, `TicketsRegistro`, `Postulantes`.
  - `backend/internal/service/registro_service.go` y `_test.go`: `RegistroService.ValidarIdentidad`.
  - `backend/internal/repository/redis/intentos.go`, `tickets.go` y `intentos_test.go`.
  - `backend/internal/repository/postgres/postulante_repo.go`: solo `ExisteCedula` (la #13 lo amplía).
  - `backend/internal/http/handler/auth_handler.go` y `_test.go`, `backend/internal/http/routes_auth.go`.
  - `backend/cmd/api/wire_auth.go` y `modules.go` (se activa `authRoutes`).
- Resumen de lo que hizo: `POST /api/v1/auth/identidad`. Valida la cédula (módulo 10) y que venga exactamente uno de `fecha_expedicion` (`AAAA-MM-DD` o `DD/MM/AAAA`) o `codigo_dactilar`. Con bloqueo vigente responde 429 con el TTL restante. Dato incorrecto o cédula inexistente → 401 "la información ingresada es incorrecta" y cuenta como fallo; al tercero crea `auth:lock:registro:{cedula}` (600 s) y responde 429 "registro fallido, inténtelo nuevamente en 10 minutos". Si coincide, reinicia el contador; si la cédula ya tiene cuenta → 409; si no, guarda `registro:ticket:{id}` (10 min) y responde `200 { ticket }`.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...` y `go test ./...` en verde. Las pruebas del servicio cubren éxito, fallo, cédula inexistente, 3 fallos → bloqueo, bloqueo vigente con dato correcto, cédula ya registrada y errores de formato. Las pruebas del repositorio Redis se saltan si Redis no está levantado.
- Qué se corrigió:
