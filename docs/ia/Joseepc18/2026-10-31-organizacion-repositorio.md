# 2026-10-31: organización del repositorio

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** ninguna (PR 1, preparación del repositorio)

## Prompts (resumen, en orden)
1. Pedí organizar el repositorio del grupo para que todos trabajen con sus agentes de forma estructurada, con buenas prácticas y sin salirse del proyecto.
2. Pasé las capturas del enunciado del examen y el `requerimiento.pdf`, y pedí transcribirlos tal cual a `.md` dentro de `docs/`.
3. Corregí el alcance: cumplir todo el PDF. Después decidí priorizar lo que califica la rúbrica (EF-11 y EF-12 obligatorios; administración y evaluación teórica opcionales).
4. Pedí verificar que el `.md` del PDF dijera todo lo que dice el PDF, y revisar si el PDF traía imágenes.
5. Pedí anotar una issue de CI con GitHub Actions para después de los andamios.
6. Pedí atender los comentarios de revisión de Dayron en el PR 1.

## Resultado
- Archivos generados o modificados: `AGENTS.md`, `CLAUDE.md`, `README.md`, `docker-compose.yml`, `.env.example`, `.gitignore`, `db/schema.sql`, `db/seed_catalogos.sql`, `docs/especificacion.md`, `docs/arquitectura.md`, `docs/api.md`, `docs/indicaciones-examen.md`, `docs/requerimiento-biess.md`, plantillas de issue y de PR, y este registro.
- Resumen de lo que hizo: redactó la documentación y las reglas, escribió el esquema SQL inicial, transcribió el PDF y las capturas, y abrió el PR 1 y la issue 2.

## Revisión humana
- Se comparó el `.md` del PDF contra el texto extraído del PDF palabra por palabra; las diferencias eran solo de formato.
- El esquema SQL se cargó en un PostgreSQL 16 temporal sin errores y `docker compose config` pasó. No se probó `docker compose up` completo.
- Dayron revisó el PR y marcó correcciones (rama `develop`, apellido, registro de IA, vacíos del contrato), que se están aplicando.
