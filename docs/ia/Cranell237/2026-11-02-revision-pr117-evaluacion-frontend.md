# 2026-11-02: revisión y correcciones del PR #117 (pantalla de evaluación del postulante)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #46 (revisión del PR #117)

## Prompt
```
revisa el pr 117
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/features/evaluacion/useExamen.ts`.
  - `frontend/src/pages/Evaluacion.tsx` y `Evaluacion.test.tsx`.
  - `frontend/src/services/evaluacion.ts` y `evaluacion.test.ts`.
- Resumen de lo que hizo:
  - **Última respuesta fuera de la nota:** si se pulsaba "Finalizar" (o el reloj llegaba a 00:00) mientras se guardaba una respuesta, la calificación podía llegar antes al servidor y esa respuesta quedaba fuera. Ahora finalizar espera a que termine el guardado en curso.
  - **Avance incorrecto al recargar:** al retomar un examen en curso solo se conocía la respuesta de la pregunta 1, así que el contador, el navegador de preguntas y el aviso "faltan N preguntas" decían que el resto estaba sin responder. `iniciarEvaluacion` indica si el examen ya estaba en curso (200) y, en ese caso, se leen las demás preguntas para recuperar lo respondido.
  - **Error invisible en "Ver mi resultado":** si fallaba, el mensaje solo se mostraba dentro del examen, no en la pantalla de inicio. Ahora se muestra.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `tsc -b`, `eslint`, `prettier`, `npm run build` y `vitest` completo (64 archivos, 498 pruebas). Pruebas nuevas: finalizar espera al guardado, examen retomado con respuestas previas, error de "Ver mi resultado" y la marca de examen retomado en el servicio.
- Qué se corrigió: lo listado arriba.
