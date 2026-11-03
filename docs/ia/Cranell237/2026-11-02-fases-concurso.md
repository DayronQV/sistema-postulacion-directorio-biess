# 2026-11-02: fases del concurso y habilitación del formulario

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #37

## Prompt
```
esta mergeado, trae los ultimos cambios y continua con la siguiente issue
```

## Resultado
- Archivos generados o modificados:
  - `db/schema.sql` (columnas `fecha_inicio`/`fecha_fin` con `CHECK`, semilla de las tres fases), `docs/api.md`.
  - Backend: `domain/fase.go`, `repository/postgres/fases_repo.go`, `service/fase_service.go`, `http/handler/admin_fases_handler.go`, regla del periodo en `postulacion_service.go`, `postulacion_envio.go` y `cargo_service.go`, `periodo_abierto` en `postulacion_handler.go`, cableado en `cmd/api/wire_admin.go` y `wire_postulacion.go`.
  - Frontend: `services/adminFases.ts`, `features/admin/TarjetaFase.tsx`, `pages/admin/Fases.tsx`, ruta `/admin/fases` y su ítem de menú, aviso de periodo cerrado en `pages/FormularioPostulacion.tsx`.
  - Pruebas: `domain/fase_test.go`, `service/fase_service_test.go`, `http/routes_admin_fases_test.go`, `repository/postgres/fases_integration_test.go`, `adminFases.test.ts`, `TarjetaFase.test.tsx`, `Fases.test.tsx` y dos casos nuevos en `FormularioPostulacion.test.tsx`.
- Resumen de lo que hizo:
  - `GET /admin/fases` (analista y administrador) y `PUT /admin/fases/{id}` (solo administrador). El servicio valida que solo la fase Postulación habilite el formulario, que para habilitarlo esté publicada y que el fin no sea anterior al inicio (`400` con `fields`).
  - Fuera del periodo (fase Postulación sin publicar, sin habilitar o fuera de fechas) crear, guardar, validar el cargo y enviar la postulación responden `403` "El periodo de postulación no está habilitado"; leer una postulación existente sigue funcionando y la respuesta trae `periodo_abierto`.
  - La semilla deja la fase Postulación publicada y habilitada sin fechas, para que el sistema funcione al levantarlo.
  - Pantalla del panel con una tarjeta por fase (casillas, fechas y Guardar); el analista la ve en solo lectura. El postulante ve un aviso en el borrador cuando el periodo está cerrado.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...` y `go test ./...` en verde (la prueba de integración de fases se salta sin `DATABASE_URL`); `npm run lint`, `npx vitest run` (446 pruebas) y `npm run build` en verde.
- Qué se corrigió: una fecha mal formada en `PUT /admin/fases/{id}` respondía `500`; ahora el handler lee las fechas como texto y responde `400` con el campo. Se reutilizó el helper `camposDe` que ya existía en las pruebas del servicio.
