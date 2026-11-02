# 2026-11-02: marcar postulantes que superaron la Fase de Méritos

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #38

## Prompt
```
ya esta mergeado, trae los ultimos cambios y continua con el otro issue

mergee un nuevo pr, traes los ultimos cambios para ver si hay conflictos
```
(Como la nómina y el detalle de postulaciones son de la #40, se eligió dejar la acción como componente reutilizable para que se coloque ahí.)

## Resultado
- Archivos generados o modificados:
  - Backend: `internal/domain/merito.go` (estado visible y puerto `Meritos`) con su prueba, `internal/repository/postgres/postulacion_merito_repo.go`, `internal/service/admin_postulacion_service.go` y su prueba, `internal/http/handler/admin_postulacion_handler.go`, `cmd/api/wire_admin.go`; `GET /postulacion` devuelve `supero_merito` y `estado_visible` (`postulacion_repo.go`, `postulacion_handler.go`, `domain/postulacion.go`) con su prueba; pruebas de rutas en `internal/http/routes_admin_test.go`.
  - Frontend: `services/adminPostulaciones.ts` y su prueba, `features/admin/AccionMerito.tsx` y su prueba, estado del postulante en `pages/FormularioPostulacion.tsx` y su prueba, tipos en `services/postulacion.ts`.
  - `docs/api.md`.
- Resumen de lo que hizo: `PUT /api/v1/admin/postulaciones/{id}/merito` con `{ supero }` marca o quita la Fase de Méritos; va en el grupo de escritura del panel (solo administrador, el analista recibe `403`) y solo acepta postulaciones `enviada` (`409` si no, `404` si no existe; la condición va en el mismo `UPDATE`). El postulante ve su estado como "Evaluación Técnica" cuando superó la Fase de Méritos. `AccionMerito` es el botón con confirmación para la nómina y el detalle de la #40; el analista solo ve el estado.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...` y `go test ./...` en verde (las rutas comprueban 401, 403 para postulante y analista y 200 para el administrador); `npm run lint`, `npm test` (53 archivos, 431 pruebas) y `npm run build` en verde, con el PR #97 ya unido.
- Qué se corrigió: una prueba del componente hacía clic en el botón equivocado del diálogo; ahora lo busca por su nombre.
