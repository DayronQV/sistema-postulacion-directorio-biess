# Sistema de Registro y Postulación al Directorio del BIESS

Módulo de registro y postulación del Concurso Público de Méritos y Oposición del Directorio del BIESS. Examen práctico de Frameworks de Programación Web (trabajo en grupo).

## Documentación para el equipo
- [Reglas para agentes y personas](AGENTS.md): léelo primero.
- [Especificación condensada](docs/especificacion.md)
- [Arquitectura](docs/arquitectura.md)
- [Contrato de la API](docs/api.md)
- [Registro de uso de IA](docs/ia/README.md)

## Prerrequisitos
- Go 1.26 o superior
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
3. Backend (lee el `.env` de la raíz; si ya definiste una variable en tu entorno, esa tiene prioridad):
   ```bash
   cd backend && go run ./cmd/api
   ```
   Comprobar que responde: `curl http://localhost:8080/health` devuelve `{"status":"ok"}`.
4. Frontend (cuando exista `frontend/`):
   ```bash
   cd frontend && npm install && npm run dev
   ```

Si el puerto 5432 de tu máquina ya está ocupado por otro PostgreSQL, cambia `POSTGRES_PORT` y el puerto de `DATABASE_URL` en tu `.env` (por ejemplo a 5433).

La instrucción del paso 4 se confirma cuando se agregue el frontend.

## Pruebas
- Backend: `go test ./...` dentro de `backend/`.
- Frontend: `npm test` dentro de `frontend/`.

## Integrantes
Dayron Quiñonez, Jose Ponce, Luis Anchundia, Estalin Jaen, Oscar Gonzabay.
