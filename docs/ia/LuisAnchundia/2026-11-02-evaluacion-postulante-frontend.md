# 2026-11-02: pantalla de evaluación del postulante

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #46

## Prompt
```
sigue con el que puedas
(Contexto: con el backend de iniciar, responder, finalizar y calificar (#43 y #44) ya unido, la issue #46
pide la pantalla del examen: instrucciones, preguntas una por una guardando cada respuesta, reloj en
cuenta regresiva sincronizado con el servidor, bloqueos de Atrás, copiar, cortar, pegar y menú
contextual, finalizar con confirmación, cierre automático a 00:00 y pantalla de resultado.)
```

## Resultado
- Archivos generados o modificados: `frontend/src/pages/Evaluacion.tsx`, `frontend/src/features/evaluacion/` (`useExamen.ts`, `useReloj.ts`, `useBloqueosExamen.ts`, `InstruccionesExamen.tsx`, `PreguntaActual.tsx`, `ResultadoExamen.tsx`), `frontend/src/services/evaluacion.ts`, `frontend/src/routes/` (ruta `/evaluacion`) y el enlace "Ir a la evaluación" en `frontend/src/pages/FormularioPostulacion.tsx`, con sus pruebas.
- Resumen de lo que hizo: el enlace aparece en la pantalla de la postulación enviada solo cuando el postulante superó la Fase de Méritos ("Evaluación Técnica"). La pantalla muestra las instrucciones y "Iniciar Evaluación"; ya en el examen, las preguntas van una por una, cada respuesta se guarda al elegirla y se puede volver a una pregunta. El reloj cuenta con `performance.now` desde el tiempo restante que manda el servidor y se vuelve a sincronizar en cada pregunta; al llegar a 00:00 llama a finalizar. Se bloquean Atrás (`popstate`), copiar, cortar, pegar, menú contextual y se avisa al cerrar la pestaña. Finalizar pide confirmación y avisa de las preguntas sin responder; el resultado muestra la nota y el aviso del correo. Un 409 del servidor (examen terminado o tiempo agotado) se califica con lo guardado.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `tsc`, `eslint`, `vitest` (494 pruebas) y `npm run build`; en el navegador contra la API real de `develop`: postulación enviada, mérito superado por el administrador, fase Evaluación publicada, banco de 3 preguntas y examen de 2 minutos. Se comprobó el enlace, las instrucciones, el reloj, guardar una respuesta, que "Atrás" y pegar quedan bloqueados, el cierre automático a 00:00 con la nota (33.33 con 1 acierto de 3) y el correo del reporte en el log.
- Qué se corrigió: nada todavía.
- Pendiente: revisión del código por la persona autora; no se probó en el navegador "Finalizar y Calificar" con confirmación ni la recarga de la página a mitad del examen (solo con pruebas unitarias).
