# 2026-11-01: capacitación

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex
- **Issue:** #30

## Prompt
```
Crear PR de las issues asignadas que estén listas y no dependan de trabajo pendiente,
ejecutando las pruebas antes de abrirlo.
```

## Resultado
- Archivos generados o modificados: dominio, validación, repositorio PostgreSQL, DTO HTTP, cableado, tipos del frontend, componente, hook y pruebas de Capacitación.
- Resumen de lo que hizo: implementó cursos, seminarios y talleres con fechas, horas positivas, respaldo PDF opcional y guardado transaccional.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go test ./...`, `go vet ./...`, `npm run lint`, `npm test` y `npm run build`.
- Qué se corrigió: se aplicaron las reglas de fechas y horas tanto en frontend como en servidor y se preservó Formación en el formulario completo.
