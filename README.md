# Sistema de Registro y Postulación al Directorio del BIESS

Módulo web del Concurso Público de Méritos y Oposición para la selección de los Miembros Principales y Alternos del Directorio del BIESS, en representación de los Afiliados, Jubilados y Empleadores. Examen práctico de Frameworks de Programación Web (trabajo en grupo).

**Sistema en línea:** https://codework-studio.com

## Integrantes
| Integrante |
| --- |
| Dayron Adrian Quiñonez Valencia |
| Jose Eduardo Ponce Carlos |
| Luis Alfredo Anchundia Cruz |
| Estalin Gabriel Jaen Garcia |
| Oscar Jordan Gonzabay Muñoz |

## El problema
El **BIESS** (Banco del Instituto Ecuatoriano de Seguridad Social) es dirigido por un Directorio en el que participan representantes de los **afiliados**, los **jubilados** y los **empleadores**. La **Resolución No. C.D. 701** del Consejo Directivo del IESS ordena que esos representantes se elijan mediante un **concurso público de méritos y oposición**, abierto a todo el país.

Ejecutar ese concurso a mano es inviable: recibir y verificar a cientos de postulantes, comprobar su identidad y su calidad ante el IESS, revisar requisitos e inhabilidades, calificar méritos y una evaluación teórica, y notificar cada etapa con hojas de cálculo y correos sueltos genera errores, demoras y poca trazabilidad. El requerimiento del BIESS pide un **módulo web** que lo haga de forma ordenada y auditable.

## Cómo lo solucionamos
Construimos una aplicación web que cubre el recorrido completo del concurso:

1. **Registro con identidad verificada (EF-11).** El postulante se registra con su cédula y la fecha de expedición o el código dactilar, que se comprueban contra el Registro Civil. Solo se aceptan ecuatorianos. Tras **3 intentos fallidos la cuenta se bloquea 10 minutos** (Redis), y la contraseña se recupera por correo.
2. **Formulario de postulación (EF-12).** El postulante llena sus datos en 11 secciones: domicilio con la cascada país, provincia y cantón; familiares; cargo, validado contra el Historial Laboral del IESS (si no cumple la calidad del sector, la postulación se cierra); formación; experiencia con el cálculo de su tiempo; publicaciones; capacitación; referencias; inhabilidades y consentimiento. **Guardar la postulación es una transacción** en PostgreSQL: todo o nada. Al enviarla recibe un correo de confirmación e imprime su formulario para firmarlo y entregarlo.
3. **Panel de administración (EF-13).** El personal del BIESS abre las **fases** del concurso con sus fechas, ve un **dashboard**, revisa la **nómina** con el detalle y los certificados de cada postulación, marca quién **superó la Fase de Méritos** y descarga la **matriz comparativa** en Excel y PDF. El analista solo consulta; el administrador decide.
4. **Evaluación teórica (EF-14).** El administrador carga el banco de preguntas y configura el examen. Cada postulante habilitado recibe preguntas sorteadas con un reloj; el sistema **califica solo**, guarda la nota y le envía el **reporte en PDF** por correo.

El Registro Civil y el Historial Laboral del IESS no tienen acceso público, así que se **simulan** detrás de una interfaz con un adaptador `mock`; cambiar a los servicios reales solo cambiaría el adaptador. El correo es real (Mailjet).

## Tecnologías
- **Frontend:** React, TypeScript, Vite y Tailwind CSS (SPA), con Vitest para las pruebas.
- **Backend:** Go con Gin, en capas (handler → service → repository) con inyección de dependencias.
- **Datos:** PostgreSQL 16 (transacciones) y Redis 7.
- **Correo:** Mailjet en producción; en local los correos se escriben en el log.
- **Producción:** Docker Compose con Caddy (HTTPS automático) en un servidor de DigitalOcean.

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
| Herramienta | Versión | Para qué |
| --- | --- | --- |
| Git | cualquiera reciente | Clonar el repositorio |
| Go | **1.26.1** o superior (`backend/go.mod`) | Backend |
| Node.js y npm | **Node 24** (la que usan el CI y el Dockerfile del frontend), con su npm | Frontend |
| PostgreSQL | **16** | Base de datos |
| Redis | **7** | Bloqueo por intentos fallidos y caché de catálogos |
| Docker y Docker Compose | Docker Desktop o Docker Engine con Compose v2 | Levantar PostgreSQL 16 y Redis 7 sin instalarlos |

Con Docker no hace falta instalar PostgreSQL ni Redis: `docker-compose.yml` los levanta. Sin Docker, instálalos en esas versiones y carga el script con `psql` (ver "Script SQL").

## Ejecución local
1. Clonar el repositorio y entrar a su carpeta (es privado: pide acceso a Dayron Quiñonez, `DayronQV` en GitHub):
   ```bash
   git clone https://github.com/DayronQV/sistema-postulacion-directorio-biess.git
   ```
   ```bash
   cd sistema-postulacion-directorio-biess
   ```
2. Crear el `.env` a partir del ejemplo:
   ```bash
   cp .env.example .env
   ```
   En PowerShell: `Copy-Item .env.example .env`.
3. **Reemplazar `JWT_SECRET` en el `.env`** por una clave de 32 caracteres o más (el valor de ejemplo es corto a propósito y **el backend no arranca hasta cambiarlo**). Para generarla:
   ```bash
   openssl rand -hex 32
   ```
   En PowerShell, sin openssl: `-join ((1..64) | ForEach-Object { '{0:x}' -f (Get-Random -Maximum 16) })`.
   El resto de valores del ejemplo sirven tal cual para desarrollo (`MAILER=log` escribe los correos en el log del backend en vez de enviarlos).
4. Levantar PostgreSQL y Redis. La primera vez, con el volumen vacío, PostgreSQL ejecuta solo `db/schema.sql` (tablas y división política) y `db/seed_admin.sql` (usuarios de prueba del panel):
   ```bash
   docker compose up -d
   ```
5. Backend (lee el `.env` de la raíz; una variable ya definida en tu entorno tiene prioridad):
   ```bash
   cd backend && go run ./cmd/api
   ```
   Comprobar que responde: `curl http://localhost:8080/health` devuelve `{"status":"ok"}`.
6. Frontend, en otra terminal:
   ```bash
   cd frontend && npm install && npm run dev
   ```
7. Abrir http://localhost:5173. "Registrarse" crea una cuenta de postulante (con las cédulas de prueba de más abajo) y "Acceso del personal del BIESS" lleva al panel de administración. La API queda en http://localhost:8080.
8. Abrir el periodo de postulación: la base arranca con todas las fases cerradas. Entra al panel como `administrador`, ve a **Fases** y, en "1. Postulación", marca "Fase publicada" y "Habilitar el formulario de postulación" (con fechas completas, día y hora, o sin ellas para dejarla abierta sin límite) y guarda. Sin esto, el formulario responde "El periodo de postulación no está habilitado".

Para compilar en lugar de ejecutar en modo desarrollo: `go build -o bin/api ./cmd/api` dentro de `backend/` y `npm run build` dentro de `frontend/` (genera `frontend/dist/`).

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

## Entregables del examen
| Entregable | Dónde está |
| --- | --- |
| Repositorio Git con historial y ramas | https://github.com/DayronQV/sistema-postulacion-directorio-biess (`main` estable, `develop` de integración, una rama por cambio y PR) |
| Prerrequisitos, compilación y ejecución local | Este README: "Prerrequisitos" y "Ejecución local" |
| `docker-compose.yml` | Raíz del repositorio: PostgreSQL 16 y Redis 7 |
| Variables de entorno de ejemplo | [`.env.example`](.env.example), explicado en "Variables de entorno" |
| Script `.sql` con DDL y datos semilla de la división política | [`db/schema.sql`](db/schema.sql) |

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
