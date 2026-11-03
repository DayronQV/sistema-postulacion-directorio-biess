# 2026-11-02: dashboard de postulantes

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex (GPT-5)
- **Issue:** #39

## Prompt
```
Implementar el dashboard de postulantes por cargo y estado, con actualización automática y pruebas, y crear su pull request.
```

## Resultado
- Archivos generados o modificados: dominio, repositorio, servicio, handler y pruebas del dashboard; servicio y pantalla del frontend; rutas, menú y documentación de la API.
- Resumen de lo que hizo: agregó los conteos agregados por cargo y estado, el total de postulaciones y una pantalla que se actualiza cada 30 segundos.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...`, `go test ./...`, `npm run lint`, `npm run build` y 437 pruebas del frontend.
- Qué se corrigió: pendiente de revisión humana en el pull request.
