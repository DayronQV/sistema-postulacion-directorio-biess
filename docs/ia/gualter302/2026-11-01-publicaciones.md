# 2026-11-01: publicaciones e investigaciones

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex
- **Issue:** #29

## Prompt
```
Crear PR de las issues asignadas que estén listas y no dependan de trabajo pendiente,
ejecutando las pruebas antes de abrirlo.
```

## Resultado
- Archivos generados o modificados: dominio, validación, repositorio PostgreSQL, DTO HTTP, cableado, tipos del frontend, componente, hook y pruebas de Publicaciones.
- Resumen de lo que hizo: implementó la pregunta Sí/No y el alta, edición, eliminación y guardado de publicaciones e investigaciones con respaldo PDF obligatorio.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go test ./...`, `go vet ./...`, `npm run lint`, `npm test` y `npm run build`.
- Qué se corrigió: se conservaron los datos de Formación en el formulario completo y se añadieron las nuevas listas a las fixtures existentes.
