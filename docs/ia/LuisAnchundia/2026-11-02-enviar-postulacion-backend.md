# 2026-11-02: backend de enviar la postulación

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #34 (parte de backend)

## Prompt
```
Sí, empieza #34
(Contexto: la issue #34 pide POST /postulacion/enviar: validar que las secciones obligatorias estén
completas y el consentimiento aceptado, 409 si ya fue enviada, cambiar a `enviada`, guardar `enviada_at`
y mandar el formulario al correo del postulante, en una transacción.)
```

## Resultado
- Archivos generados o modificados: `backend/internal/service/postulacion_envio.go`, `backend/internal/service/postulacion_service.go`, `backend/internal/domain/postulacion.go`, `backend/internal/repository/postgres/postulacion_repo.go`, `backend/internal/http/handler/postulacion_handler.go`, `backend/internal/http/routes_postulacion.go`, `backend/cmd/api/wire_postulacion.go`, `docs/api.md` y las pruebas de servicio, handler y PostgreSQL.
- Resumen de lo que hizo: `PostulacionService.Enviar` lee la postulación, responde 409 si no es un borrador, 422 con `fields` por sección incompleta (domicilio, cargo, formación, experiencia, capacitaciones, referencias, inhabilidades y consentimiento) y, si todo está bien, dentro de una transacción con la fila bloqueada la marca `enviada` (`MarcarEnviada` guarda `enviada_at`) y manda el correo por el puerto `Mailer`. Si el correo falla, la transacción se deshace y la postulación sigue en borrador.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test -race ./...`, incluida la prueba contra PostgreSQL real de `MarcarEnviada`.
- Qué se corrigió: nada todavía.
- Pendiente: revisión del código por la persona autora; las reglas de "sección completa" repiten las del formulario (`secciones/index.ts`) y deben mantenerse iguales. Está apilado sobre #94 porque usa `Inhabilidades.Completa()`.
