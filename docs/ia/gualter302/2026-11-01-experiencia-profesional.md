# 2026-11-01: experiencia profesional

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex
- **Issue:** #28

## Prompt
```
Crear PR de las issues asignadas que estén listas y no dependan de trabajo pendiente,
ejecutando las pruebas antes de abrirlo.
```

## Resultado
- Archivos generados o modificados: dominio y calculadora de duración, validación del servicio, repositorio PostgreSQL, DTO HTTP, cableado, tipos y servicios del frontend, componente de experiencia, hook y pruebas.
- Resumen de lo que hizo: agregó experiencias con respaldo obligatorio, fechas válidas, descripción de hasta 1000 caracteres, cálculo por fila y total, y persistencia dentro del guardado transaccional.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go test ./...`, `go vet ./...`, `npm run lint`, `npm test` y `npm run build`.
- Qué se corrigió: se ajustaron las fixtures del formulario completo y los selectores accesibles de las pruebas nuevas.
