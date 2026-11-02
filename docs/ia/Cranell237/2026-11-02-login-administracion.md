# 2026-11-02: inicio de sesión de administración con roles

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #36

## Prompt
```
analiza y mira si puedes realizar las 3 issues faltantes, que un compañero depende de esas
```

## Resultado
- Archivos generados o modificados:
  - Backend: `internal/domain/admin.go`, `internal/repository/postgres/usuarios_admin_repo.go`, `internal/service/admin_auth_service.go` y su prueba, `internal/http/handler/admin_auth_handler.go`, `internal/http/routes_admin.go` y su prueba, `cmd/api/wire_admin.go`, `cmd/api/modules.go`.
  - Datos: `db/seed_admin.sql` (analista y administrador con bcrypt) montado en `docker-compose.yml`.
  - Frontend: `services/adminAuth.ts`, `features/admin/` (formulario de login, layout y menú del panel), `pages/admin/` (login y panel) con su prueba, `routes/index.tsx` y `routes/rutas.ts` con sus pruebas.
  - Documentación: `docs/api.md`, `docs/arquitectura.md` (claves `auth:*:admin:{usuario}`) y `README.md` (usuarios de prueba).
- Resumen de lo que hizo: `POST /api/v1/admin/auth/login` con `{ usuario, password }` emite un JWT con el rol del usuario (`analista` o `administrador`) y aplica el mismo bloqueo que el login del postulante (3 fallos, 10 minutos) con su propio contador. Las rutas `/api/v1/admin/*` exigen token de analista o administrador (`401` sin token, `403` con token de postulante) y se registran por módulo (`ModuloAdmin`) en un grupo de lectura y otro de escritura solo para el administrador (`403` para el analista). En el frontend, `/admin/login` y el panel `/admin` con menú y rutas protegidas por rol; un analista o administrador no abre las pantallas del postulante.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...` y `go test ./...` en verde; las pruebas de rutas comprueban 401 sin token, 403 para el postulante, 403 para el analista en escritura y 200 para el administrador. `npm run lint`, `npm test` (51 archivos, 422 pruebas) y `npm run build` en verde. Los hashes de la semilla se generaron con bcrypt y se comprobaron contra sus contraseñas.
- Qué se corrigió: las pruebas de rutas que iniciaban sesión con un token sin rol para entrar a las pantallas del postulante ahora usan un token con rol `postulante`, porque esas rutas pasaron a exigirlo.
