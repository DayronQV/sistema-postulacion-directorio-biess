# 2026-11-03: README final para la entrega

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** #53 (README final)

## Prompt
```
Actúa como desarrollador del equipo preparando la entrega del examen práctico de Frameworks
de Programación Web.

Actualiza el README.md para que cumpla el formato del entregable: prerrequisitos de
instalación (versiones de Go, Node.js, PostgreSQL y Redis), instrucciones de compilación y
ejecución local (o docker-compose para la base de datos y Redis) y las variables de entorno de
ejemplo (.env.example). Que alguien pueda clonar el repositorio y levantarlo siguiendo los
pasos al pie de la letra.

Antes de eso, al inicio: los integrantes, de qué trata el sistema, cuál es el problema y cómo
lo solucionamos.

Comprueba cada versión y cada comando contra el código y el CI, sin inventar. Commit en mi
nombre, en inglés y con Conventional Commits, y PR con la plantilla hacia develop.
```

## Resultado
- Archivos generados o modificados: `README.md`.
- Resumen de lo que hizo: al inicio, la URL del sistema en línea, los integrantes, el problema (el concurso que ordena la Resolución C.D. 701, inviable a mano) y la solución (EF-11 a EF-14, con las integraciones simuladas); "Qué incluye" pasó a "Tecnologías" para no repetir. "Prerrequisitos" es ahora una tabla con las versiones de Git, Go, Node, PostgreSQL, Redis y Docker. "Ejecución local" suma clonar el repositorio, el paso obligatorio de reemplazar `JWT_SECRET` (con el comando para generarlo, también en PowerShell) y cómo compilar. Se agregó una tabla "Entregables del examen" que señala dónde está cada uno. Los integrantes pasaron del final al inicio.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): versiones contra `backend/go.mod` (Go 1.26.1), el CI (Node 24, PostgreSQL 16, Redis 7) y los Dockerfile; los scripts que carga `docker-compose.yml` (`01-schema.sql` y `02-seed_admin.sql`); el comando de PowerShell para la clave (64 caracteres hexadecimales); `go build -o bin/api ./cmd/api` compila y `backend/bin/` queda fuera de git.
- Qué se corrigió: el README no decía que el `JWT_SECRET` de ejemplo impide arrancar el backend, que era el primer tropiezo de quien clona; se agregó como paso explícito. Se anotó que el repositorio es privado y hay que pedir acceso.
