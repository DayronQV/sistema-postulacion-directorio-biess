# 2026-11-02: exportar matriz comparativa

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex
- **Issue:** #41

## Prompt
```
Revisa qué me falta y haz el PR.
```

## Resultado
- Archivos generados o modificados: servicio y consulta de la matriz administrativa, generadores Excel/PDF, endpoint protegido, botones de descarga, pruebas, contrato de API y dependencias de generación.
- Resumen de lo que hizo: se implementó la exportación de la matriz comparativa en `.xlsx` y PDF horizontal, respetando los filtros activos de cargo, estado y elegibilidad.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go test ./...`, `go vet ./...`, `npm run lint`, `npx tsc -b`, `npm run build` y 462 pruebas de Vitest.
- Qué se corrigió: se mantuvo la descarga segura de archivos temporales y se validaron formato y filtros antes de consultar PostgreSQL.
