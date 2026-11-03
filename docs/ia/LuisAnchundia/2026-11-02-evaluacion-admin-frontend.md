# 2026-11-02: administración del banco de preguntas y configuración de la evaluación

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #45

## Prompt
```
creo que ya se puede avanzar
(Contexto: con el backend del banco de preguntas (#42) y el login de administración (#36) ya unidos, la
issue #45 pide en el panel: listado con buscar/editar/eliminar, formulario de pregunta con opciones
dinámicas, carga masiva desde archivo con vista previa y reporte de errores, configuración del examen y
vista de resultados con descarga de reportes PDF.)
```

## Resultado
- Archivos generados o modificados: `frontend/src/services/adminEvaluacion.ts`, `frontend/src/features/evaluacion/admin/` (`preguntas.ts`, `FormularioPregunta.tsx`, `CargaMasiva.tsx`), `frontend/src/pages/admin/EvaluacionBanco.tsx`, `EvaluacionConfiguracion.tsx`, `EvaluacionResultados.tsx`, `frontend/src/features/admin/menu.ts`, `frontend/src/routes/index.tsx` y sus pruebas.
- Resumen de lo que hizo: tres pantallas nuevas del panel, solo para el administrador. El banco lista con búsqueda y paginación, y permite crear, editar y eliminar (con confirmación); el formulario maneja de 2 a 8 opciones y la marca de la correcta. La carga masiva lee el CSV en el navegador (comillas, `;` o `,`, BOM), muestra una vista previa con los errores por línea y solo deja cargar si todas son válidas; si el servidor rechaza, lista los errores por línea. La configuración guarda duración y número de preguntas y avisa que el horario lo fija la fase Evaluación. Los resultados muestran la nota de cada examen y descargan el PDF de los finalizados.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `tsc`, `eslint`, `vitest` (527 pruebas) y `npm run build`; en el navegador contra la API real (login de administrador, alta de una pregunta, configuración con el error 422 por pedir más preguntas de las que hay) y el contrato de la carga CSV con `curl`.
- Qué se corrigió: el comentario de Dayron en la issue (backend sin fecha de inicio, horario en la fase Evaluación, aviso si la ventana es más corta que el examen) se leyó después de abrir el PR. El código ya seguía `docs/api.md` en todo menos el último punto, que se agregó: la pantalla de configuración avisa, con la duración que se escribe, si la fase Evaluación dura menos que el examen.
- Pendiente: revisión del código por la persona autora. La pantalla de resultados depende de `GET /admin/evaluacion/resultados` y del reporte PDF (PR #115 de Dayron, aún sin unir): con el backend actual responde 404; solo se probó con pruebas unitarias y el contrato de `docs/api.md`. La carga masiva con un archivo real no se probó a mano en el navegador.
