# 2026-11-01: sección Inhabilidades (parte frontend)

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #33 (solo el frontend)

## Prompt
```
Haz lo que puedas hacer sin dependencia de los demás.
(Contexto: la sección de las 13 preguntas de inhabilidades con el detalle de la pregunta 7 y el aviso
cuando una respuesta inhabilita. La parte de backend de #33 depende de #20 y no se hizo.)
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/features/postulacion/secciones/Inhabilidades.tsx` y su prueba.
  - `frontend/src/features/postulacion/inhabilidades.ts` y su prueba (texto de las preguntas, avisos y completitud).
  - `frontend/src/features/postulacion/secciones/index.ts` (registro de la sección).
- Resumen de lo que hizo: las 13 preguntas Sí/No con el texto del requerimiento, todas obligatorias; la 7 pide institución y motivo si es Sí; un aviso informativo lista las respuestas que inhabilitan sin bloquear el guardado. Al guardar convierte las respuestas a las claves "1" a "13" del contrato.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` y `npm run build`.
- Qué se corrigió: el aviso contaba como inhabilitante la pregunta 1 sin responder; ahora solo cuenta las ya respondidas. Se cambió `watch` por `useWatch` por la advertencia de lint.
- Pendiente: la parte de backend de #33 (`elegibilidad.go`, repositorio, columna `elegible` y cambios de `db/schema.sql` y `docs/api.md`) no se hizo porque depende de #20. Revisión del código por la persona autora.
