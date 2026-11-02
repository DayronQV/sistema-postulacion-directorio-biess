# 2026-11-02: backend de inhabilidades y elegibilidad

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #33 (parte de backend)

## Prompt
```
si se puede hacer ya la 33?
(Contexto: la issue #33 pide guardar las 13 preguntas de inhabilidades dentro de la transacción de
PUT /postulacion y calcular `elegible` en el servidor: pregunta 1 en Sí y de la 2 a la 13 en No.)
```

## Resultado
- Archivos generados o modificados: `backend/internal/domain/elegibilidad.go`, `backend/internal/domain/postulacion.go`, `backend/internal/service/inhabilidades_validacion.go`, `backend/internal/service/postulacion_service.go`, `backend/internal/repository/postgres/inhabilidades_repo.go`, `backend/internal/http/handler/postulacion_handler.go`, `backend/cmd/api/wire_postulacion.go`, `db/schema.sql`, `docs/api.md` y las pruebas de dominio, servicio, handler y PostgreSQL.
- Resumen de lo que hizo: nueva sección `InhabilidadesRepo` (una fila por pregunta respondida, `detalle` JSONB para la 7) que se agrega a la lista de secciones de la transacción; el servicio calcula `elegible` con `domain.Inhabilidades.Elegible()` y el repositorio lo guarda en la columna nueva `postulaciones.elegible`. El DTO lee las claves "1" a "13" y `removido_detalle`, y responde solo las preguntas contestadas más `elegible` (solo lectura). La 7 en Sí exige institución y motivo (400 con `fields`); que las 13 estén respondidas se exigirá al enviar (#34).

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet`, `go build` y `go test -race ./...`, con la prueba de integración contra PostgreSQL 16 real (base recreada con el esquema nuevo).
- Qué se corrigió: una respuesta que no es booleana devolvía 500 desde el decodificador propio; ahora es un 400 con el campo en `fields`.
- Pendiente: revisión del código por la persona autora; las bases de datos ya creadas necesitan la columna `elegible` (se recrea con `docker compose down -v`).
