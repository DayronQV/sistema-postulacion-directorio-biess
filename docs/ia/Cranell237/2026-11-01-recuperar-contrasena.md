# 2026-11-01: recuperar y restablecer contraseña

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #15

## Prompt
```
entonces podemos continuar con la #15, correcto?
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/service/auth_service.go` y `recuperacion_service_test.go`: `RecuperacionService.Recuperar` y `Restablecer`.
  - `backend/internal/service/registro_service.go`: la comprobación contra el Registro Civil pasa a `comprobarIdentidad`, compartida por el registro y la recuperación.
  - `backend/internal/domain/auth.go`: puerto `TokensRestablecer`, `Intentos.Desbloquear`, `Postulantes.ActualizarPassword` y `Cuenta.Email`.
  - `backend/internal/repository/redis/restablecer.go`, `intentos.go` y `restablecer_test.go`; `backend/internal/repository/postgres/postulante_repo.go`.
  - `backend/internal/http/handler/auth_handler.go` y `_test.go`, `backend/internal/http/routes_auth.go`, `backend/cmd/api/wire_auth.go`.
  - `docs/api.md` (nuevo `POST /auth/restablecer`) y `docs/arquitectura.md` (claves `auth:reset:{token}` y `auth:*:recuperar:{cedula}`).
- Resumen de lo que hizo: `POST /api/v1/auth/recuperar` valida la identidad con el mismo bloqueo que el registro (contador propio); si coincide responde `202` tenga o no cuenta la cédula, y si la tiene guarda `auth:reset:{token}` (30 min) y envía por el mailer simulado el enlace `<frontend>/restablecer?token=...`. `POST /api/v1/auth/restablecer` consume el token con `GETDEL` (un solo uso), guarda el nuevo hash bcrypt, borra los bloqueos de login y responde `204`; token inválido o vencido → `401`.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...` y `go test ./...` en verde. Las pruebas del servicio sacan el enlace del correo enviado y lo usan para cambiar la contraseña, comprueban que no sirva dos veces, que se borren los bloqueos de login, que una cédula sin cuenta no reciba correo y el bloqueo al tercer fallo. Las pruebas de Redis se saltan si Redis no está levantado.
- Qué se corrigió: en la revisión del PR (DayronQV), que probó el flujo con PostgreSQL y Redis reales, se agregaron dos pruebas que faltaban (el acierto reinicia el contador de fallos y el enlace se emite con TTL de 30 minutos); el desbloqueo del login tras cambiar la contraseña pasó a ser de mejor esfuerzo, con su prueba, y `.env.example` explica que el primer origen de `CORS_ALLOWED_ORIGIN` es la URL de los enlaces.
