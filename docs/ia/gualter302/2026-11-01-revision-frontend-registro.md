# 2026-11-01: revisión del registro frontend

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex
- **Issue:** #16

## Prompt
```
Revisar el siguiente PR asignado, corregir directamente los hallazgos y dejar el merge para el usuario.
```

## Resultado
- Archivos generados o modificados: `frontend/src/features/auth/esquemas.ts`, `frontend/src/features/auth/esquemas.test.ts` y este registro.
- Resumen de lo que hizo: se revisó el registro en dos pasos y se corrigió la validación de la fecha de expedición para comparar contra el día local del navegador y rechazar fechas inexistentes.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` y `npm run build`.
- Qué se corrigió: el uso de `toISOString()` podía aceptar el día siguiente durante la noche en Ecuador porque compara en UTC.
