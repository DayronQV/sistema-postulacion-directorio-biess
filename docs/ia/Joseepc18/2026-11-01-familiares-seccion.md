# 2026-11-01: sección Información adicional (familiares en el IESS o BIESS)

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #24

## Prompt
Se pidió implementar la issue #24 (EF-12, sección Información adicional) en backend y frontend, repitiendo los patrones ya aceptados. Backend: revisar qué dejó la #20 para familiares y completar solo lo que falte (al menos una fila si la respuesta es Sí, filas borradas si es No, valores de las listas validados, errores por fila con el formato de `docs/api.md`) y conservar el orden de captura sin cambiar `db/schema.sql`. Frontend: componente `Familiares.tsx` de presentación (valores iniciales y `onGuardar` por props) más un hook mínimo que lo conecta a la API, con pruebas. Se indicó que `PUT /postulacion` reemplaza el estado completo, por lo que el frontend debe leer la postulación, cambiar su sección y enviar el formulario entero, y que esto se demuestre con una prueba de integración.

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/familiar.go` (parentescos, instituciones y límites).
  - `backend/internal/service/familiares_validacion.go` y una llamada en `postulacion_service.go`.
  - `backend/internal/repository/postgres/familiares_repo.go` (ids ordenados por captura).
  - Pruebas: `familiares_service_test.go` y `familiares_integration_test.go`.
  - `frontend/src/services/postulacion.ts` (cliente de la API de postulación y sus tipos).
  - `frontend/src/features/postulacion/secciones/Familiares.tsx`, `reglasFamiliares.ts` y `useFamiliares.ts`.
  - Pruebas de Vitest del componente, las reglas, el hook y el servicio.
- Resumen de lo que hizo: el servicio exige al menos una fila si la respuesta es Sí, restringe parentesco e institución a sus listas, acota los textos y señala la fila (`familiares[i].campo`). El orden de captura se conserva sin tocar el esquema: cada fila se inserta con un UUID versión 7 construido con la hora y la posición de la fila, y `GET` ordena por `id`. El componente permite agregar, editar y eliminar filas en un modal y muestra los errores del servidor en su fila. El hook lee la postulación en cada guardado, cambia solo `tiene_familiares` y `familiares`, y envía el formulario entero.

## Revisión humana
- Backend: `go vet`, `go build`, `go test -count=1` y `go test -race -count=1` pasan con PostgreSQL 16 y Redis 7 reales (0 pruebas saltadas). Frontend: `npm run lint`, `npm test` y `npm run build` pasan.
- Recorrido real con la API levantada: registro y login con una cédula de prueba; `PUT` con cinco familiares no alfabéticos y `GET` repetido devuelven el mismo orden; Sí sin filas, valores fuera de lista y No con filas responden 400 con el campo de la fila y no cambian lo guardado; No sin filas borra las filas; enviar solo la sección borra el domicilio y el cargo, y enviar el formulario entero (`GET` más `PUT`) los conserva.
- A revisar: el orden se apoya en el `id` generado por el repositorio (si otra parte inserta filas con el id por defecto, el orden entre ellas no está garantizado), el tope de 20 familiares, que el `PUT` rechaza (400) un No acompañado de filas en lugar de borrarlas en silencio, y que `tiene_familiares = false` en un borrador nuevo se muestra como "No" porque la tabla no distingue "sin responder".
