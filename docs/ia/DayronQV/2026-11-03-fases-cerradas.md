# 2026-11-03: las fases del concurso arrancan cerradas

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (pedido de Dayron tras las pruebas en producción)

## Prompt
```
Actúa como desarrollador backend sénior (Go, PostgreSQL) del sistema del concurso del
Directorio del BIESS.

Al levantar la base, la fase Postulación arranca publicada y con el formulario habilitado,
sin fechas. No quiero eso: quiero que arranque cerrada y abrirla yo mismo desde el panel de
administración con sus fechas.

Requisitos: cambia el script db/schema.sql; ajusta las pruebas que dependan del estado
inicial sin que fallen en las bases locales ya creadas; documenta el paso de abrir la fase en
el README; corre toda la suite, incluidas las pruebas de integración contra una base limpia.
Commit en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla hacia
develop.
```

## Resultado
- Archivos generados o modificados: `db/schema.sql`, `backend/internal/repository/postgres/fases_integration_test.go`, `backend/internal/repository/postgres/schema_test.go` (nuevo) y `README.md`.
- Resumen de lo que hizo: las tres fases se siembran sin publicar y sin formulario (`('Postulación', 1, FALSE, FALSE)`). La prueba de integración de fases deja de exigir el estado inicial (una base local creada con un script anterior no lo refleja) y prueba que las banderas y las fechas se guardan y se leen; el estado inicial lo prueba `schema_test.go` leyendo el propio script, sin base de datos. El README suma el paso de abrir la fase Postulación desde el panel.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet ./...`; la suite completa del backend, con las pruebas de integración, contra una base temporal y limpia de `postgres:16` creada con el script nuevo (todas en verde; las tres fases arrancaron cerradas); mutación: volver a sembrar la Postulación abierta hace fallar `schema_test.go`.
- Qué se corrigió: contra la base local de desarrollo las pruebas de integración fallaban porque esa base tiene cuentas de prueba y se creó con un script anterior (le falta el CHECK de fechas de `fases`); no era un fallo del cambio. Una primera corrida con `postgres:16-alpine` falló en el orden alfabético con tildes de los catálogos porque esa imagen ordena distinto; con `postgres:16`, la del proyecto, pasa.
- En la base de producción, además, se limpiaron a mano la cuenta y la postulación de prueba (con respaldo previo) y se cerraron las tres fases; ese cambio de datos no es parte de este PR.
