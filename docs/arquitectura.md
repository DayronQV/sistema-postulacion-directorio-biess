# Arquitectura (propuesta base)

Propuesta para que cinco personas trabajen en paralelo sin pisarse. Se ajusta por PR a este archivo.

## Estructura del repositorio
```
backend/
  cmd/api/main.go            # raíz de composición: crea dependencias y las inyecta
  internal/
    config/                  # lectura de variables de entorno
    domain/                  # entidades, errores de dominio, interfaces (puertos)
    service/                 # casos de uso; sin HTTP ni SQL
    repository/
      postgres/              # implementación de los puertos con pgx
      redis/                 # intentos, bloqueo y caché
    integration/
      registrocivil/         # interfaz + adaptador mock
      historiallaboral/      # interfaz + adaptador mock
      mailer/                # interfaz + adaptador que escribe en log
    http/
      router.go
      handler/               # un archivo por recurso
      middleware/            # auth, errores, timeout, request id
frontend/
  src/
    components/              # ui/ (botón, input, select, loader) y form/
    features/                # auth/, postulacion/ (componentes, hooks y tipos de cada flujo)
    pages/
    routes/                  # rutas y guardia de ruta protegida
    services/                # único lugar con llamadas a la API
    store/                   # estado global (sesión)
db/
  schema.sql                 # DDL
  seed_catalogos.sql         # países, provincias y cantones
docker-compose.yml
```

## Capas del backend
`handler → service → repository`. Cada capa depende de **interfaces** definidas en `domain/`. `main.go` es el único lugar que conoce las implementaciones concretas. Los servicios se prueban con repositorios falsos, sin base de datos.

## Transacciones
`PostulacionService.Guardar` abre una transacción en el repositorio (`BEGIN`), escribe `postulaciones`, `postulacion_familiares` y lo que se agregue, y hace `COMMIT`; ante cualquier error hace `ROLLBACK`. Ninguna escritura parcial es válida.

## Redis
| Clave | Uso | TTL |
| --- | --- | --- |
| `auth:fail:login:{cedula}` | contador de fallos de login | 10 min |
| `auth:lock:login:{cedula}` | bloqueo activo de login (se crea al 3.er fallo) | 10 min |
| `auth:fail:registro:{cedula}` | contador de fallos de registro | 10 min |
| `auth:lock:registro:{cedula}` | bloqueo activo de registro | 10 min |
| `registro:ticket:{id}` | identidad ya validada, pendiente de correo y contraseña | 10 min |
| `cache:catalogo:paises` | lista de países | 1 h |
| `cache:catalogo:provincias:{paisId}` | provincias | 1 h |
| `cache:catalogo:cantones:{provinciaId}` | cantones | 1 h |

Un login correcto borra los contadores de fallo de esa cédula.

## Modelo de datos (resumen; el DDL vive en `db/schema.sql`)
- `paises`, `provincias`, `cantones`: catálogos con claves foráneas encadenadas.
- `postulantes`: cédula única, correo único, hash de contraseña (bcrypt).
- `postulaciones`: una por postulante, con código único, estado (`borrador`, `enviada`, `cerrada`), cargo, datos personales y domiciliarios.
- `postulacion_familiares`: familiares en IESS/BIESS (0..n por postulación).

## Autenticación
JWT firmado, expiración corta, enviado como `Authorization: Bearer`. Contraseñas con bcrypt.

## Errores y tiempos
- Formato único de error (ver `docs/api.md`): 400, 401, 403, 404, 409, 422, 429, 500. Un 500 nunca expone detalles internos.
- Timeout por petición en el servidor y en las consultas a la base; el cliente cancela con `AbortController`.

## Frontend
- Rutas públicas: registro, login, recuperar contraseña. Rutas protegidas: formulario de postulación.
- Formularios con `react-hook-form` y validación con `zod`.
- Cada petición expone `loading`, `success` y `error` a la interfaz.

## Datos de prueba (semilla)
Un puñado de cédulas válidas en el Registro Civil simulado: al menos una ecuatoriana, una casada y una **extranjera** (para probar el cierre de la postulación). Se cargan desde `db/seed_catalogos.sql` o desde el propio adaptador `mock`.
