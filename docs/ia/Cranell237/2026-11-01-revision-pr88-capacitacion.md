# 2026-11-01: revisión y correcciones del PR #88 (capacitación)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #30 (revisión del PR #88)

## Prompt
```
revisa la 88, si esta correcto apruebas, para mergear
```

## Resultado
- Archivos generados o modificados:
  - Merge con `develop`: `wire_postulacion.go`, `domain/postulacion.go`, `postulacion_handler.go`, `postulacion_service.go`, `services/postulacion.ts` y los fixtures de tres pruebas.
  - `backend/internal/domain/capacitacion.go`, `backend/internal/service/capacitaciones_validacion.go` y su prueba, `backend/internal/repository/postgres/capacitaciones_repo.go`.
  - `frontend/src/features/postulacion/secciones/Capacitacion.tsx` y su prueba, `useCapacitacion.ts`.
- Resumen de lo que hizo:
  - Resolvió los conflictos con Experiencia (#86) y Publicaciones (#87), conservando todas las secciones con Capacitación al final, como en la especificación.
  - Tope de 10000 horas por evento: sin él, un número fuera del rango de INTEGER hacía fallar el INSERT con 500 en lugar de 400.
  - Repositorio y frontend con el formato del proyecto (SQL legible, Prettier) y comentarios en lo exportado.
  - La tabla muestra "Curso/Seminario/Taller" en lugar del valor interno; máximo de 50 eventos como el servidor.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test ./internal/...`; `npm run lint`, `npx tsc -b`, `npm run build` y `npx vitest run --no-file-parallelism` (302 pruebas). En paralelo fallan por tiempo pruebas de otras secciones (issue #89).
- Qué se corrigió: lo listado arriba.
- Pendiente:
