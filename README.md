# Sistema de Registro y Postulación al Directorio del BIESS

Módulo de registro y postulación del Concurso Público de Méritos y Oposición del Directorio del BIESS. Examen práctico de Frameworks de Programación Web (trabajo en grupo).

## Documentación para el equipo
- [Reglas para agentes y personas](AGENTS.md): léelo primero.
- [Especificación condensada](docs/especificacion.md)
- [Arquitectura](docs/arquitectura.md)
- [Contrato de la API](docs/api.md)

## Prerrequisitos
- Go 1.22 o superior
- Node.js 20 o superior
- Docker y Docker Compose (para PostgreSQL 16 y Redis 7)

## Ejecución local
1. Copiar las variables de entorno y ajustarlas:
   ```bash
   cp .env.example .env
   ```
2. Levantar la base de datos y Redis (carga `db/schema.sql` y `db/seed_catalogos.sql` la primera vez):
   ```bash
   docker compose up -d
   ```
3. Backend (cuando exista `backend/`):
   ```bash
   cd backend && go run ./cmd/api
   ```
4. Frontend (cuando exista `frontend/`):
   ```bash
   cd frontend && npm install && npm run dev
   ```

Las instrucciones de los pasos 3 y 4 se confirman cuando se agregue cada proyecto.

## Pruebas
- Backend: `go test ./...` dentro de `backend/`.
- Frontend: `npm test` dentro de `frontend/`.

## Integrantes
Dayron Quiñonez, Jose Ponce, Luis Anchundia, Estalin Jaen, Oscar Gonzabay.
