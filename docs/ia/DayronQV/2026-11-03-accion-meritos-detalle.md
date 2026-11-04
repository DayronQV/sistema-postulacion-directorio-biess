# 2026-11-03: marcar la Fase de Méritos desde el detalle de la postulación

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (fallo detectado al probar el flujo completo en producción)

## Prompt
```
Actúa como desarrollador frontend sénior (React y TypeScript) del sistema del concurso del
Directorio del BIESS.

Probando en producción: un postulante ya envió su postulación, entro al panel como
administrador y no encuentro dónde marcar que superó la Fase de Méritos. Sin eso no puede
rendir la evaluación. Revisa por qué no aparece y corrígelo.

Requisitos: reutiliza lo que ya exista; solo el administrador puede marcar (el analista solo
ve el estado); pruebas del caso y mutaciones a mano; tipos, lint, todas las pruebas y el
build. Commit en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla
hacia develop.
```

## Resultado
- Archivos generados o modificados: `pages/admin/PostulacionDetalle.tsx` y su prueba.
- Resumen de lo que hizo: el componente `AccionMerito` (con su confirmación, la llamada `PUT /admin/postulaciones/{id}/merito` y sus pruebas) existía desde el PR que agregó la Fase de Méritos, pero no estaba montado en ninguna pantalla. Ahora el detalle de la postulación muestra una tarjeta "Fase de Méritos" con el estado y, para el administrador y solo en postulaciones enviadas, el botón "Marcar que superó".

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npx tsc -b`, `npm run lint`, `npx vitest run` (664 pruebas) y `npm run build`; pruebas nuevas del administrador (marca y ve "Superó la Fase de Méritos") y del analista (ve el estado y no tiene botón); 3 mutaciones (estado de méritos fijo, postulación tomada como no enviada, otro id), todas detectadas. No se probó en el navegador local porque ahí no hay ninguna postulación enviada.
- Qué se corrigió: se había agregado un `onMarcado` que actualizaba el estado de la página, pero la página no muestra ese estado y la acción ya actualiza su propia etiqueta; se quitó por innecesario.
