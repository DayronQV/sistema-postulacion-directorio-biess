# 2026-10-31: catálogos geográficos con caché en Redis

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #21

## Prompt
```
Implementa la issue #21 "[EF-12] Backend: catálogos geográficos con caché en Redis": endpoints
públicos GET /catalogos/paises, /catalogos/paises/{id}/provincias y /catalogos/provincias/{id}/cantones,
con patrón cache-aside en Redis (claves cache:catalogo:* y TTL de 1 h). Un fallo de Redis no debe tumbar
el endpoint (se consulta PostgreSQL y se registra el aviso); 404 si el id no existe y 400 si no es numérico.
Seguir el patrón de capas ya aceptado (puertos en domain, servicio sin HTTP ni SQL, repositorios como
implementaciones, cableado solo en la raíz de composición). Pruebas del servicio con falsos, del handler
y, si es viable, del repositorio de Redis contra Redis real.
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/catalogo.go` (puertos `Catalogos` y `CacheCatalogos`)
  - `backend/internal/service/catalogo_service.go` y `catalogo_service_test.go`
  - `backend/internal/repository/postgres/catalogo_repo.go` y `catalogo_repo_test.go`
  - `backend/internal/repository/redis/cache.go` y `cache_test.go`
  - `backend/internal/http/handler/catalogo_handler.go` y `catalogo_handler_test.go`
  - `backend/internal/http/routes_catalogos.go`
  - `backend/cmd/api/wire_catalogos.go` y una línea en `backend/cmd/api/modules.go`
- Resumen de lo que hizo: el servicio busca la lista en la caché y, si no está, la lee de PostgreSQL y la guarda con TTL de 1 hora. Si la caché falla, registra un aviso y responde desde PostgreSQL. El repositorio de PostgreSQL distingue "padre inexistente" (404) de "padre sin hijos" (lista vacía) con una sola consulta, y ordena los países con Ecuador primero. El handler responde 400 si el id no es un entero válido.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet`, `go build`, `go test` y `go test -race` en verde. Se levantó PostgreSQL y Redis reales con `docker compose` (proyecto y puertos propios) y se comprobó con `curl` los tres endpoints: 194 países, 24 provincias de Ecuador y los cantones de cada provincia coinciden con la base. La segunda lectura sale de Redis (clave `cache:catalogo:*` con TTL cercano a 3600 s sin reiniciarse). Con Redis detenido los endpoints siguen respondiendo 200 desde PostgreSQL, con aviso en el log.
- Qué se corrigió: `gofmt` convirtió los finales de línea de archivos ajenos a la issue; se restauraron antes de commitear. La prueba del orden de países se ajustó para no comparar bytes, porque la base ordena con su colación (Bélgica antes que Belice).
