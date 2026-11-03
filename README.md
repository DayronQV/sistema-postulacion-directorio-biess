# Sistema de Registro y Postulación al Directorio del BIESS

Módulo de registro y postulación del Concurso Público de Méritos y Oposición del Directorio del BIESS. Examen práctico de Frameworks de Programación Web (trabajo en grupo).

## Qué incluye
Frontend en React, TypeScript, Vite y Tailwind CSS; backend en Go con Gin; PostgreSQL 16 y Redis 7.

- **Registro y acceso del postulante (EF-11):** validación de identidad con el Registro Civil simulado (fecha de expedición o código dactilar), bloqueo de 10 minutos tras 3 intentos fallidos con Redis y recuperación de contraseña.
- **Formulario de postulación (EF-12):** información personal, domicilio con la cascada país, provincia y cantón, familiares, cargo (con la validación de Historial Laboral simulada), formación, experiencia con su cálculo de tiempo, publicaciones, capacitación, referencias, inhabilidades, consentimiento y envío. Solo se admite nacionalidad ecuatoriana. Guardar la postulación es una transacción.
- **Panel de administración (EF-13):** fases del concurso, dashboard, nómina con el detalle de cada postulación y sus certificados, y la matriz comparativa en Excel y PDF.
- **Evaluación teórica (EF-14):** banco de preguntas, examen con sorteo y reloj, calificación automática y reporte en PDF por correo.

Las integraciones externas (Registro Civil, Historial Laboral y correo) están simuladas detrás de una interfaz con un adaptador `mock`.

## Arquitectura
```mermaid
flowchart LR
    Navegador["Navegador<br/>SPA React + TypeScript"] -->|"HTTP /api/v1 (JWT)"| Handler

    subgraph Backend["Backend Go + Gin"]
        Handler["Handler<br/>HTTP y validación de entrada"] --> Service["Service<br/>reglas de negocio"]
        Service --> Repository["Repository<br/>interfaces inyectadas"]
    end

    Repository --> PG[("PostgreSQL 16<br/>datos y transacciones")]
    Repository --> Redis[("Redis 7<br/>intentos fallidos, bloqueo 10 min<br/>y catálogos con TTL")]
    Service -.->|"adaptador mock"| Externos["Registro Civil<br/>Historial Laboral<br/>Correo"]
```
- **Frontend:** componentes en `components/`, llamadas a la API solo desde `services/`.
- **Backend:** el handler solo traduce HTTP; el service no conoce HTTP ni SQL; el repository es una interfaz que se inyecta, así que las integraciones externas se reemplazan por un adaptador `mock`.
- **PostgreSQL:** guarda la postulación completa en una transacción de varias tablas.
- **Redis:** cuenta los intentos fallidos de ingreso (bloqueo de 10 minutos tras 3) y cachea los catálogos con TTL.

Más detalle en [docs/arquitectura.md](docs/arquitectura.md).

## Documentación para el equipo
- [Reglas para agentes y personas](AGENTS.md): léelo primero.
- [Especificación condensada](docs/especificacion.md)
- [Arquitectura](docs/arquitectura.md)
- [Contrato de la API](docs/api.md)
- [Registro de uso de IA](docs/ia/README.md)

## Prerrequisitos
- Go 1.26.1 o superior
- Node.js 24 (la versión que usan el CI y el Dockerfile del frontend) y npm
- Docker y Docker Compose (para PostgreSQL 16 y Redis 7)

Si no usas Docker, necesitas PostgreSQL 16 y Redis 7 instalados; ver "Script SQL" más abajo.

## Ejecución local
1. Copiar las variables de entorno y ajustarlas:
   ```bash
   cp .env.example .env
   ```
2. Levantar la base de datos y Redis. La primera vez, con el volumen vacío, PostgreSQL ejecuta solo el script `db/schema.sql` (tablas y división política) y `db/seed_admin.sql` (usuarios de prueba del panel):
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
5. Abrir la aplicación: el frontend en http://localhost:5173 y la API en http://localhost:8080. Desde la pantalla de inicio, "Registrarse" crea una cuenta de postulante (usa las cédulas de prueba de más abajo) y "Acceso del personal del BIESS" lleva al panel de administración.

Si el puerto 5432 de tu máquina ya está ocupado por otro PostgreSQL, cambia `POSTGRES_PORT` y el puerto de `DATABASE_URL` en tu `.env` (por ejemplo a 5433).

El proyecto no tiene migraciones: el esquema solo se carga desde cero. Si ya tenías una base de una versión anterior, recréala (se borran sus datos) con `docker compose down -v` y `docker compose up -d`.

## Variables de entorno
Se definen en el `.env` de la raíz (copia de `.env.example`, que no se sube al repositorio). El backend lista todo lo que falta o es inválido al arrancar.

| Variable | Para qué sirve | Valor de ejemplo |
| --- | --- | --- |
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Usuario, clave y nombre de la base que crea `docker compose` | `biess`, `cambiar_esta_clave`, `biess` |
| `POSTGRES_PORT` | Puerto de tu máquina para PostgreSQL | `5432` |
| `REDIS_PORT` | Puerto de tu máquina para Redis | `6379` |
| `DATABASE_URL` | Cadena de conexión a PostgreSQL que usa el backend (**obligatoria**; debe coincidir con las tres primeras y con `POSTGRES_PORT`) | `postgres://biess:cambiar_esta_clave@localhost:5432/biess?sslmode=disable` |
| `REDIS_ADDR` | Dirección de Redis que usa el backend (**obligatoria**; debe coincidir con `REDIS_PORT`) | `localhost:6379` |
| `API_PORT` | Puerto en el que escucha la API | `8080` |
| `JWT_SECRET` | Clave con la que se firman los tokens de sesión (**obligatoria**, mínimo 32 caracteres; genera una con `openssl rand -hex 32`) | `cambiar_esto` (se rechaza hasta que lo reemplaces) |
| `JWT_TTL_MINUTES` | Minutos que dura la sesión | `60` |
| `REQUEST_TIMEOUT_SECONDS` | Tiempo máximo de cada petición a la API | `10` |
| `CORS_ALLOWED_ORIGIN` | Orígenes del frontend autorizados, separados por coma; el primero también arma los enlaces de los correos | `http://localhost:5173` |
| `MAILER` | Adaptador de correo: `log` escribe los correos en el log del backend en vez de enviarlos (desarrollo); `mailjet` los envía de verdad por la API de Mailjet (producción) | `log` |
| `MAILJET_API_KEY`, `MAILJET_SECRET_KEY` | Claves de la API de Mailjet (**obligatorias con `MAILER=mailjet`**) | vacías |
| `MAIL_FROM` | Remitente de los correos, verificado en Mailjet (**obligatorio con `MAILER=mailjet`**) | `tu.correo@gmail.com` |
| `MAIL_FROM_NAME` | Nombre que ve el destinatario como remitente | `Concurso Directorio BIESS` |
| `PADRON_ARCHIVO` | Archivo JSON con las personas que reconocen los simuladores del Registro Civil y del IESS; vacío = las cédulas de prueba (ver `backend/internal/integration/padron/padron.example.json`) | vacía |
| `VITE_API_URL` | URL de la API que usa el frontend (Vite) | `http://localhost:8080/api/v1` |
| `DOMINIO` | Solo producción (`docker-compose.prod.yml`): dominio público sin `https://`, para el certificado HTTPS | `postulacion.example.com` |

Las variables con valor por defecto (`API_PORT`, `JWT_TTL_MINUTES`, `REQUEST_TIMEOUT_SECONDS`, `CORS_ALLOWED_ORIGIN`, `MAILER`, `MAIL_FROM_NAME`, `PADRON_ARCHIVO`) pueden omitirse. En producción, el compose arma `DATABASE_URL`, `REDIS_ADDR`, `CORS_ALLOWED_ORIGIN` y `VITE_API_URL` por su cuenta; ver los comentarios de `.env.example`.

## Script SQL
`db/schema.sql` es el script único de la base de datos: crea todas las tablas y, al final, carga la división política (194 países, 24 provincias y 199 cantones). Se puede ejecutar de una vez en una base vacía, sin Docker:
```bash
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/schema.sql
```
Los usuarios de prueba del panel van aparte, en `db/seed_admin.sql`, porque tienen contraseñas públicas y no deben cargarse en producción.

## Datos de prueba (Registro Civil simulado)
El Registro Civil real no es accesible para este proyecto, así que se simula con 7 ciudadanos de prueba (`backend/internal/integration/registrocivil/mock.go`). Las cédulas son válidas por módulo 10 y no pertenecen a personas reales. Para comprobar la identidad (registro y recuperación de contraseña) se ingresa **uno** de los dos datos, a elección: la fecha de expedición o el código dactilar (10 caracteres entre letras y números).

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

## Usuarios de administración
El panel no tiene registro público y tiene su propio acceso, con usuario y contraseña (no con cédula): http://localhost:5173/admin/login, o el botón "Acceso del personal del BIESS" de la pantalla de inicio. `db/seed_admin.sql` siembra estos usuarios al crear la base. Las contraseñas son solo para desarrollo y pruebas.

| Usuario | Contraseña | Rol |
| --- | --- | --- |
| `analista` | `Analista2026` | Analista (Talento Humano): solo lectura |
| `administrador` | `Admin2026` | Administrador (Comisión de Selección) |

## Pruebas
- Backend: `go test ./...` dentro de `backend/`.
- Frontend: `npm test` dentro de `frontend/`.

## Integración continua
El workflow `.github/workflows/ci.yml` corre en cada Pull Request y en cada push a `develop` y `main`, con dos trabajos independientes:
- **Backend** (con PostgreSQL 16 y Redis 7 reales, esquema y semilla cargados, para que las pruebas de integración no se salten): `go vet ./...`, `go build ./...`, `go test -count=1 ./...` y `go test -race -count=1 ./...`.
- **Frontend** (Node 24): `npm ci`, `npm run lint`, `npm test` y `npm run build`.

Para reproducirlo en local, con `.env` listo y `docker compose up -d` corriendo (el `DATABASE_URL` y el `REDIS_ADDR` del `.env` deben estar exportados en tu shell; sin ellos las pruebas de integración se saltan):
```bash
cd backend && go vet ./... && go build ./... && go test -count=1 ./... && go test -race -count=1 ./...
```
```bash
cd frontend && npm ci && npm run lint && npm test && npm run build
```

## Integrantes
Dayron Quiñonez, Jose Ponce, Luis Anchundia, Estalin Jaen, Oscar Gonzabay.
