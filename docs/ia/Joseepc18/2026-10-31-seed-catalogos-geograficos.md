# 2026-10-31: datos semilla de la división política

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #10

## Prompt
Se dio la instrucción de revisar qué de las issues asignadas se podía hacer ya y encargarse de ello. La issue #10 es la única sin dependencias, así que se trabajó esa: completar `db/seed_catalogos.sql` con los países, provincias y cantones del requerimiento.

## Resultado
- Archivos generados o modificados: `db/seed_catalogos.sql`.
- Resumen de lo que hizo: un script temporal (no versionado) leyó los ítems 1.2.1, 1.2.2 y 1.2.3 de `docs/requerimiento-biess.md` y generó los `INSERT`. Los ids se resuelven por nombre con subconsultas y los cantones se asocian a su provincia.

## Revisión humana
- Conteos obtenidos de la fuente: 194 países, 24 provincias y 199 cantones, que coinciden con la issue.
- Se debe verificar que el script cargue sin errores en una base limpia con `docker compose down -v && docker compose up -d`.
