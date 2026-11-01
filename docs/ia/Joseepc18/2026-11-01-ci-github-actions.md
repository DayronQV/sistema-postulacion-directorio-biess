# 2026-11-01: workflow de CI con GitHub Actions

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #2

## Prompt
```
Implementa la issue #2 "ci: add GitHub Actions workflow for tests, lint and build". Un workflow
.github/workflows/ci.yml que se dispare en pull_request y en push a develop y a main, con dos trabajos
independientes. Backend: Go según go.mod, go vet, go build, go test y go test -race, con PostgreSQL 16 y
Redis 7 como services (healthchecks), esquema y semilla cargados con psql y DATABASE_URL / REDIS_ADDR
exportadas, para que las pruebas de integración no se salten. Frontend: Node LTS, npm ci, lint, test y
build con caché de npm. Permisos mínimos, concurrency que cancele ejecuciones antiguas, timeout en cada
trabajo y acciones fijadas a una versión mayor que exista de verdad. Sin despliegue ni cambios en el código
de la aplicación. Una sección corta "Integración continua" en el README. Verificar el YAML y correr los
mismos comandos en local con PostgreSQL y Redis reales.

Ajuste posterior tras la revisión: cancel-in-progress también cancelaba la ejecución anterior en develop y
main cuando entraban dos merges seguidos. Cambiarlo a `${{ github.event_name == 'pull_request' }}`, volver
a pasar actionlint, subir la rama y abrir el PR hacia develop.
```

## Resultado
- Archivos generados o modificados:
  - `.github/workflows/ci.yml` (nuevo)
  - `README.md` (sección "Integración continua")
- Resumen de lo que hizo: el workflow tiene un trabajo `backend` (PostgreSQL 16 y Redis 7 como servicios, carga de `db/schema.sql` y `db/seed_catalogos.sql` con `psql`, luego `go vet`, `go build`, `go test` y `go test -race`) y un trabajo `frontend` (Node 24, `npm ci`, lint, pruebas y build). Usa `actions/checkout@v7`, `actions/setup-go@v7` y `actions/setup-node@v7`, cuyas etiquetas mayores se comprobaron en GitHub.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): el YAML pasa `actionlint` sin avisos. Los mismos comandos se corrieron en local con PostgreSQL y Redis reales (`docker compose` con proyecto y puertos propios): backend con `go vet`, `go build`, `go test` y `go test -race` en verde, 261 pruebas ejecutadas y 0 saltadas (sin `DATABASE_URL` ni Redis eran 250 ejecutadas y 11 saltadas); frontend con `npm ci`, lint, `npm test` (46 pruebas) y build en verde. El workflow en sí solo se puede comprobar al subirlo a GitHub.
- Qué se corrigió: la concurrencia ahora cancela las ejecuciones antiguas solo en pull requests; en develop y main cada push termina su ejecución. `actionlint` volvió a pasar sin avisos.
