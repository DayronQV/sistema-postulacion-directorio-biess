# 2026-11-03: diagrama de arquitectura y variables de entorno en el README

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** #53 (README final), con el alcance ampliado en un comentario

## Prompt
```
Actúa como desarrollador del equipo preparando el README final del examen.

La issue 53 pide dos cosas que al README le faltan: un diagrama simple de la arquitectura y
las variables de entorno explicadas. Agrégalas, tocando solo README.md:

1. Una sección "Arquitectura" con un diagrama (Mermaid, que GitHub dibuja solo) que muestre
   el navegador, las capas del backend, PostgreSQL, Redis y las integraciones simuladas.
2. Una sección "Variables de entorno" con una tabla de todas las variables de .env.example:
   para qué sirve cada una, si es obligatoria y un valor de ejemplo.

Requisitos:
- Todo lo que diga el README debe poder comprobarse en el código: verifica las variables y
  sus valores por defecto contra backend/internal/config/config.go, y qué capa usa Redis.
- Parte de la rama del PR #123, porque también reescribe el README.
- No agregues dependencias ni toques otros archivos.
- Commit en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla del
  repositorio. Deja constancia de la sesión en docs/ia/DayronQV/.
```

## Resultado
- Archivos generados o modificados: `README.md`.
- Resumen de lo que hizo: agregó la sección "Arquitectura" (diagrama Mermaid con navegador, handler, service, repository, PostgreSQL, Redis y los adaptadores simulados, más una viñeta por pieza) y la sección "Variables de entorno" (tabla de las 13 variables de `.env.example` con su función, si son obligatorias y un ejemplo, y una nota de cuáles pueden omitirse y cuáles arma el compose de producción).

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): los valores por defecto y las variables obligatorias contra `backend/internal/config/config.go`, el bloqueo de 3 intentos y 10 minutos contra `service/auth_service.go`, y que Redis lo usa el repository (`repository/redis`). Es un cambio de solo documentación, sin pruebas que correr.
- Qué se corrigió: el diagrama mostraba a Redis conectado al service; se cambió para que salga del repository, como en el código. El diagrama de Mermaid no se pudo dibujar en el entorno de la sesión, así que falta verlo en la vista previa de GitHub.
