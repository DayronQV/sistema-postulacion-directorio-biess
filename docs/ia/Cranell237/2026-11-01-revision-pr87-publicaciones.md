# 2026-11-01: revisión y correcciones del PR #87 (publicaciones e investigaciones)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #29 (revisión del PR #87)

## Prompt
```
revisa la 87 ahora
```

## Resultado
- Archivos generados o modificados:
  - Merge con `develop`: `wire_postulacion.go`, `domain/postulacion.go`, `postulacion_handler.go`, `postulacion_service.go`, `services/postulacion.ts` y los fixtures de `DomicilioSeccion.test.tsx` e `InformacionPersonalSeccion.test.tsx`.
  - `backend/internal/domain/publicacion.go`, `backend/internal/repository/postgres/publicaciones_repo.go` (comentarios).
  - `frontend/src/features/postulacion/secciones/Publicaciones.tsx` y su prueba, `usePublicaciones.ts`.
- Resumen de lo que hizo:
  - Resolvió los conflictos con Experiencia profesional (#86) conservando las dos secciones, con Publicaciones después de Experiencia como en la especificación.
  - Las preguntas Sí/No usan el componente `RadioSiNo` del proyecto en lugar de radios sueltos y un `Select`.
  - Responder "No" oculta las filas sin borrarlas hasta guardar (como Familiares), para no perder lo cargado si fue un error.
  - "¿Guarda relación…?" empieza sin responder y es obligatoria; antes quedaba en "No" por omisión.
  - El error del respaldo tiene `role="alert"`, hay un máximo de 50 filas como en el servidor y el archivo tiene el formato de Prettier.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test ./internal/...`; `npm run lint`, `npx tsc -b`, `npm run build` y las pruebas de Publicaciones (6). En la suite completa fallan por tiempo, de forma intermitente, pruebas de otras secciones que solas pasan.
- Qué se corrigió: lo listado arriba.
- Pendiente: subir el límite de tiempo de las pruebas del frontend (5 s por defecto) en un cambio aparte.
