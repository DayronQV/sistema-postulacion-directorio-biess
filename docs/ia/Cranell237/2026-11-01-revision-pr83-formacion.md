# 2026-11-01: revisión y correcciones del PR #83 (formación académica)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #27 (revisión del PR #83)

## Prompt
```
quiero que revises el pr #83, para ver si cumple con su issue y no hay ciertos errores por corregir
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/formacion.go`, `backend/internal/service/formacion_validacion.go` y su prueba.
  - `backend/internal/repository/postgres/formacion_repo.go` y `formacion_integration_test.go`.
  - `frontend/src/features/postulacion/secciones/Formacion.tsx` y su prueba.
  - `frontend/src/components/form/DocumentoUpload.tsx` y su prueba, `frontend/src/components/ui/FileUpload.tsx`.
- Resumen de lo que hizo:
  - El respaldo PDF pasa a ser opcional, como en el contrato: `formacion_academica.documento_id` admite NULL y la especificación solo lo exige en Experiencia. El repositorio guarda NULL y lo lee como vacío (antes la lectura fallaba con NULL).
  - Validación: un mismo PDF no puede respaldar dos títulos; máximo de 20 formaciones y largo máximo de los textos, como familiares y referencias.
  - `DocumentoUpload` acepta `eliminarEnServidor={false}`: en Formación, Quitar o Reemplazar ya no intenta borrar un PDF que la versión guardada sigue usando (el servidor respondía 409 y no se podía quitar).
  - Formación borra los PDF subidos que ninguna fila usa al guardar o al salir sin guardar, para que no sumen al límite de 50 MB.
  - El error del servidor muestra también los campos; `FileUpload` no muestra un tamaño falso para los documentos ya guardados.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test ./internal/...`; `npm run lint`, `npx tsc -b`, `npm test` (263, dos corridas) y `npm run build`. Las pruebas contra PostgreSQL las corre el CI.
- Qué se corrigió: lo listado arriba.
- Pendiente:
