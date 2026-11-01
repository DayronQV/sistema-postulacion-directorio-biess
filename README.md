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
4. Frontend:
   ```bash
   cd frontend && npm install && npm run dev
   ```

Si el puerto 5432 de tu máquina ya está ocupado por otro PostgreSQL, cambia `POSTGRES_PORT` y el puerto de `DATABASE_URL` en tu `.env` (por ejemplo a 5433).

## Datos de prueba (Registro Civil simulado)
El Registro Civil real no es accesible para este proyecto, así que se simula con 7 ciudadanos de prueba (`backend/internal/integration/registrocivil/mock.go`). Las cédulas son válidas por módulo 10 y no pertenecen a personas reales. Para comprobar la identidad (registro y recuperación de contraseña) basta con acertar **uno** de los dos datos: la fecha de expedición o el código dactilar.

| Cédula | Caso que permite probar | Nacionalidad | Estado civil | Fecha de expedición | Código dactilar |
| --- | --- | --- | --- | --- | --- |
| 1712345675 | Flujo feliz (afiliado) | Ecuatoriana | Soltero/a | 20/05/2019 | V1234V5678 |
| 0912345675 | Muestra datos de la pareja (jubilado) | Ecuatoriana | Casado/a | 11/08/2017 | E3344I5566 |
| 1312345679 | Unión de hecho (empleador) | Ecuatoriana | Unión de hecho | 03/02/2021 | V2468E1357 |
| 0103456786 | Bloqueo por nacionalidad | Colombiana | Soltero/a | 17/10/2022 | V9876V5432 |
| 1709876542 | Sin calidad de afiliado, jubilado ni empleador | Ecuatoriana | Soltero/a | 29/06/2020 | I1122V3344 |
| 0924681356 | Pruebas de bloqueo por intentos fallidos | Ecuatoriana | Soltero/a | 05/12/2018 | E5566V7788 |
| 1802468130 | Pruebas de recuperación de contraseña | Ecuatoriana | Divorciado/a | 22/03/2016 | V3141V5926 |

Cualquier otra cédula responde "no existe" en el simulador.

## Pruebas
- Backend: `go test ./...` dentro de `backend/`.
- Frontend: `npm test` dentro de `frontend/`.

## Integrantes
Dayron Quiñonez, Jose Ponce, Luis Anchundia, Estalin Jaen, Oscar Gonzabay.
