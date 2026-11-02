# 2026-11-01: sección Selección del cargo con validación del Historial Laboral

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #26

## Prompt
```
Implementa la issue #26 "[EF-12] Sección Selección del cargo con validación del Historial Laboral",
backend y frontend, repitiendo el patrón de las secciones anteriores (componente de presentación con
props y un hook que lo conecta a la API, montable por el contenedor del formulario).
Backend: validar la elegibilidad por cargo (afiliados, jubilados, empleador) con el puerto del Historial
Laboral. Resultado válido: "Postulante afiliado al IESS", "Postulante acredita jubilación" o "Postulante
es empleador". Resultado inválido: 422 con el mensaje literal del requerimiento y la postulación pasa a
cerrada, igual que el cierre por nacionalidad. El bloqueo debe cumplirse en el servidor llamando a la API
directamente: se valida siempre que se guarde un cargo en el PUT. Organización de respaldo: si responde Sí,
nombre y documento (PDF, tipo organizacion, 4 MB, de la misma postulación) son obligatorios. Las cédulas de
prueba de afiliado, jubilado y empleador pasan en su cargo y 1709876542 se bloquea en los tres. Pruebas con
falsos y de integración contra PostgreSQL real.
Frontend: Cargo.tsx con opción única entre los tres cargos con su texto completo, resultado de la
validación, bloqueo del formulario si no cumple (mismo patrón que Información personal), pregunta de
respaldo según el cargo con nombre y subida del PDF si responde Sí, y botón "Guardar Postulación". Pruebas
con Vitest y Testing Library.
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/service/cargo_service.go` (+ `cargo_service_test.go`), `postulacion_service.go`
  - `backend/internal/domain/cargo.go`
  - `backend/internal/http/handler/cargo_handler.go` (+ prueba) y `routes_postulacion.go`
  - `backend/cmd/api/wire_postulacion.go` y `modules.go`
  - `backend/internal/repository/postgres/cargo_integration_test.go` y ajustes de los helpers de integración existentes
  - `frontend/src/features/postulacion/reglasCargo.ts`, `useCargo.ts`, `secciones/Cargo.tsx` y `secciones/CargoSeccion.tsx` (+ pruebas)
  - `frontend/src/services/postulacion.ts` (`validarCargo`)
- Resumen de lo que hizo: `PUT /postulacion` consulta el Historial Laboral cada vez que trae un cargo; si no se cumple la calidad, cierra la postulación (sin guardar el formulario) y responde 422 con el mensaje literal del requerimiento. Se agregó `POST /postulacion/cargo/validar` para mostrar el resultado antes de guardar: no escribe nada si el cargo es elegible y cierra la postulación si no lo es. El nombre y el documento de respaldo se exigen juntos; el repositorio ya comprobaba que el documento sea de tipo organizacion y de la misma postulación. En el frontend, la sección valida al elegir el cargo, pregunta por la organización según el cargo, sube el PDF con el componente existente y pasa a la pantalla de bloqueo cuando el servidor responde 422.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet`, `go build`, `go test` y `go test -race` en verde con PostgreSQL 16 y Redis 7 reales (504 pruebas y subpruebas, 0 saltadas). En el frontend, `npm run lint`, `npm test` y `npm run build`. Recorrido real contra la API con las cédulas de prueba: cada una pasa en su cargo, se bloquea en uno que no le corresponde (422 con el mensaje literal y estado cerrada) y 1709876542 se bloquea en los tres cargos por el `PUT` y por la validación.
- Qué se corrigió: varias pruebas de integración existentes guardaban `jubilados` o `empleador` con la cédula afiliada; ahora usan un cargo elegible para esa cédula, porque el servidor ya valida el cargo.

## Actualización: sincronización con develop y ajustes de la revisión
- Se integró `develop` (se habían mergeado la formación académica y la sesión del frontend). El conflicto estaba en `Guardar` y `validarFormularioEn` de `postulacion_service.go`: se conservaron `limpiarFormacion` y `validarFormacion` junto a `limpiarRespaldo` y `validarRespaldo`, y el bloque del cargo ahora valida con `validarFormularioEn(f, s.ahora())`. Se agregó `formacion: []` al fixture de `CargoSeccion.test.tsx`, que rompía `tsc -b`.
- El error del guardado muestra el nombre legible de cada campo en lugar de la clave técnica. Al responder "No", el PDF ya subido se borra con `DELETE /postulacion/documentos/:id` (y el que tenía guardada la postulación, tras guardar el "No"); un 409 se ignora. Con sus pruebas.
- Verificado de nuevo con PostgreSQL y Redis reales: `go vet`, `go build`, `go test` y `go test -race` en verde (513 pruebas y subpruebas, 0 saltadas); `lint`, `build` y `npm test` (339) en verde.

## Actualización: segunda sincronización con develop
- Se integró `develop` otra vez (experiencia, publicaciones y capacitación). El conflicto estaba de nuevo en `Guardar` y `validarFormularioEn`: se conservaron los `limpiar*` y `validar*` de las cuatro secciones nuevas junto a `limpiarRespaldo` y `validarRespaldo`. Se agregaron `experiencia`, `duracion_experiencia_total`, `publicaciones` y `capacitaciones` al fixture de `CargoSeccion.test.tsx`.
- Verificado de nuevo con PostgreSQL y Redis reales: `go vet`, `go build`, `go test` y `go test -race` en verde (540 pruebas y subpruebas, 0 saltadas); `lint`, `build` y `npm test` (357) en verde.

