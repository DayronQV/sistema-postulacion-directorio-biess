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

## Contexto, timeout y errores de la petición
- **Timeout:** el middleware pone un plazo en `c.Request.Context()` (`REQUEST_TIMEOUT_SECONDS`). Los handlers pasan siempre ese contexto al servicio, y el servicio a los repositorios (pgx y go-redis). Si se usa `context.Background()`, la consulta ignora el timeout. El middleware solo responde cuando el handler vuelve; no interrumpe código que no mira el contexto.
- **Cuerpo JSON:** los handlers lo leen con `middleware.BindJSON(c, &req)`. Un JSON mal formado, un tipo incorrecto o un campo `binding` inválido responde 400 `VALIDATION_ERROR` con `fields`, nunca 500.
- **Errores:** los servicios devuelven errores de `domain` (`domain.Validation`, `domain.NotFound`, ...) y el handler llama a `middleware.Fail(c, err)`. Una petición cancelada por el cliente (`context.Canceled`) no se registra como error interno.

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
- Tablas hijas de `postulaciones` (0..n): `postulacion_familiares`, `formacion_academica`, `experiencia_profesional`, `publicaciones`, `capacitaciones` y `postulacion_inhabilidades` (una fila por pregunta, con `detalle` JSON para la pregunta 7).
- `documentos`: PDF guardados como `BYTEA` en la base (los hosts gratuitos no tienen disco persistente). El servidor valida tipo PDF, tamaño por documento y 50 MB por postulante.
- `usuarios_admin`: analistas y administradores sembrados, con rol.
- `fases`: fases del concurso y bandera de habilitación del formulario.
- Evaluación: `eval_config`, `eval_preguntas` (con opciones y respuesta correcta), `eval_intentos` (preguntas sorteadas, hora de cierre, nota) y `eval_respuestas`.

## Archivos, exportaciones y evaluación
- Subida de PDF por `multipart/form-data`; el tamaño máximo se aplica con `http.MaxBytesReader` antes de leer el cuerpo.
- Exportaciones: Excel con `excelize` y PDF con `maroto` o `gofpdf` (la librería se confirma en su issue).
- Evaluación: el servidor guarda la hora de cierre y es la autoridad del tiempo; el reloj del navegador es solo informativo. El bloqueo de "Atrás" y de copiar/pegar es una ayuda del frontend, no una garantía de seguridad.
- Roles: `postulante`, `analista`, `administrador`. El middleware de autorización valida el rol en cada ruta.

## Autenticación
JWT firmado, expiración corta, enviado como `Authorization: Bearer`. Contraseñas con bcrypt.

## Errores y tiempos
- Formato único de error (ver `docs/api.md`): 400, 401, 403, 404, 409, 422, 429, 500. Un 500 nunca expone detalles internos.
- Timeout por petición en el servidor y en las consultas a la base; el cliente cancela con `AbortController`.

## Frontend
- Rutas públicas: registro, login, recuperar contraseña. Rutas protegidas: formulario de postulación.
- Formularios con `react-hook-form` y validación con `zod`.
- Cada petición expone `loading`, `success` y `error` a la interfaz.
- Sesión: el token JWT se guarda en `sessionStorage` (no en `localStorage`), así que vive solo en la pestaña: al cerrarla o abrir otra se pide login de nuevo. La sesión se cierra sola al vencer `expires_in` o ante un `401`. El rol se lee del claim `rol` solo para mostrar u ocultar pantallas; el servidor valida siempre.

## Datos de prueba (semilla)
Un puñado de cédulas válidas en el Registro Civil simulado: al menos una ecuatoriana, una casada y una **extranjera** (para probar el cierre de la postulación). Se cargan desde `db/seed_catalogos.sql` o desde el propio adaptador `mock`.
