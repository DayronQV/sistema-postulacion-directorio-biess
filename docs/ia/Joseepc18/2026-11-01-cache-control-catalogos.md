# 2026-11-01: cabecera Cache-Control en catálogos

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code, claude-sonnet-5-5
- **Issue:** #66

## Prompt
```
Implementar la issue #66: que GET /catalogos/paises, /catalogos/paises/{id}/provincias y
/catalogos/provincias/{id}/cantones respondan con Cache-Control: public, max-age=3600 solo en
las respuestas 200; las de error (400, 404, 500) no la llevan. Con prueba del handler.
```

## Resultado
- Archivos generados o modificados: `backend/internal/http/handler/catalogo_handler.go`, `backend/internal/http/handler/catalogo_handler_test.go`.
- Resumen de lo que hizo: añadió la constante `catalogoCacheControl` y la escribe en `respondItems` solo en el camino exitoso. Dos pruebas nuevas: la cabecera está en las tres listas con 200 y no está en los casos 400, 404 y 500.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...`, `go build ./...` y `go test -race -count=1 ./...` en `backend/`, todo en verde.
- Qué se corrigió: nada.
