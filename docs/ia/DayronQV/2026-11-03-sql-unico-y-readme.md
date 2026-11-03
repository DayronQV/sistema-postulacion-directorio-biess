# 2026-11-03: script SQL único y README final

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** #53 (README final con instrucciones completas)

## Prompt
```
Actúa como desarrollador full stack sénior preparando la entrega de un examen de Frameworks
de Programación Web: el sistema del concurso del Directorio del BIESS (frontend React/Vite,
backend Go con Gin, PostgreSQL y Redis, todo levantado con Docker Compose).

El enunciado del examen pide, entre los entregables, "un archivo .sql con el script DDL de
creación de tablas y datos semilla de catálogos (división política)" y un README con
prerrequisitos, instrucciones de ejecución y variables de entorno de ejemplo. Hoy el DDL está
en db/schema.sql y la división política en db/seed_catalogos.sql, así que no hay un único
archivo.

Haz lo siguiente:
1. Deja un solo script con el DDL y los datos semilla de la división política, conservando el
   nombre db/schema.sql (lo citan el CI, los Docker Compose, AGENTS.md y el código). Elimina
   db/seed_catalogos.sql para no tener dos copias que puedan desincronizarse. Los usuarios de
   prueba del panel (db/seed_admin.sql) se quedan aparte: tienen contraseñas públicas y no
   deben cargarse en producción.
2. Actualiza todo lo que apuntaba al archivo eliminado: los dos Docker Compose, el workflow de
   CI, la documentación y los comentarios del código.
3. Demuestra con evidencia que el script único es equivalente al anterior: compara el
   volcado completo (estructura y datos) de una base construida con los archivos viejos contra
   una construida solo con el script nuevo, comprueba que un PostgreSQL vacío lo carga con el
   mismo mecanismo de Docker (initdb) sin errores, y corre las pruebas del backend, incluidas
   las de integración, contra esa base.
4. Corrige el README: versión de Node real (24), el paso de la base de datos exacto, un resumen
   de qué incluye cada módulo (EF-11 a EF-14), las URLs de cada servicio, cómo entrar al panel
   de administración (usuario y contraseña, ruta /admin/login y el botón "Acceso del personal
   del BIESS"), una sección del script SQL y la advertencia de que no hay migraciones.
5. Trabaja en una rama nueva desde develop. Haz los commits en mi nombre, en inglés y con
   Conventional Commits, sin mencionar IA ni agregar coautoría, y abre un PR con la plantilla
   del repositorio.
6. Deja constancia de la sesión en docs/ia/DayronQV/ con la plantilla del equipo.
```

## Resultado
- Archivos generados o modificados: `db/schema.sql` (ahora con la división política al final), `db/seed_catalogos.sql` (eliminado), `docker-compose.yml`, `docker-compose.prod.yml`, `.github/workflows/ci.yml`, `docs/arquitectura.md`, un comentario de `domicilio_integration_test.go` y `README.md`.
- Resumen de lo que hizo: unió el DDL y la división política en un único `db/schema.sql`, quitó el archivo de catálogos y ajustó todas las referencias, y reescribió las partes del README que estaban desactualizadas o incompletas.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): volcado completo de dos bases (la de los archivos viejos y la del script único) comparado byte a byte, con resultado idéntico (misma estructura y 420 filas de datos); un PostgreSQL nuevo y vacío que carga el script por el mecanismo `docker-entrypoint-initdb.d` sin ningún error (19 tablas, 194 países, 24 provincias y 199 cantones, con todos los cantones enlazados a Ecuador); `go vet` y `go test ./...` del backend contra esa base; validación de ambos Docker Compose.
- Qué se corrigió: una frase de `docs/arquitectura.md` que decía que las cédulas de prueba se cargaban desde el SQL, cuando están en el adaptador `mock`; el README decía Node 20 y el CI y el Dockerfile usan Node 24.
