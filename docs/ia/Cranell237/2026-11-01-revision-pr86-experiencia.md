# 2026-11-01: revisión y correcciones del PR #86 (experiencia profesional)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #28 (revisión del PR #86)

## Prompt
```
revisa el pr 86
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/service/experiencia_duracion.go` y su prueba (nuevos), `backend/internal/service/postulacion_service.go`.
  - `backend/internal/repository/postgres/experiencia_repo.go`, `backend/internal/domain/experiencia.go`.
  - `docs/api.md` (total de experiencia en `GET /postulacion`).
  - `frontend/src/features/postulacion/experiencia.ts` y su prueba, `secciones/Experiencia.tsx` y su prueba, `useExperiencia.ts`.
- Resumen de lo que hizo:
  - La calculadora de duración se ejecutaba dentro del repositorio; ahora el repositorio solo lee y escribe, y el servicio calcula la duración de cada experiencia y el total en `Obtener` (capas handler → service → repository).
  - `docs/api.md` documenta `duracion_experiencia_total`, que el PR agregaba a la respuesta sin documentar.
  - Frontend: singular y plural en las duraciones ("1 año, 1 mes y 1 día"), `role="alert"` en el error del respaldo, máximo de 50 experiencias como el servidor, formato con Prettier y pruebas que pegan el texto para no cortarse por tiempo.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test ./internal/...`; `npm run lint`, `npx tsc -b`, `npm test` (292, dos corridas) y `npm run build`. Las pruebas contra PostgreSQL las corre el CI.
- Qué se corrigió: lo listado arriba.
- Pendiente:
