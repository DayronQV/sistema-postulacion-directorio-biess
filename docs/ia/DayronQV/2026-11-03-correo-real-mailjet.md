# 2026-11-03: correo real con Mailjet

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** #52 (correo real en producción)

## Prompt
```
Actúa como desarrollador backend sénior (Go, Gin, arquitectura en capas) del sistema del
concurso del Directorio del BIESS.

Quiero que los correos del sistema (recuperación de contraseña, postulación enviada y
resultado de la evaluación) lleguen de verdad, como pide el requerimiento del BIESS. Hoy el
correo solo se escribe en el log.

Antes de programar, comprueba en el servidor de producción (solo lectura) si puede abrir
conexiones SMTP salientes. Si no puede, usa un proveedor gratuito que envíe por su API
HTTPS y que permita mandar a cualquier destinatario con un remitente Gmail verificado, sin
dominio propio.

Requisitos:
- Un adaptador nuevo detrás de la misma interfaz Mailer, con adjuntos PDF, sin tocar los
  servicios que envían correos. En local se sigue usando el adaptador de log.
- Las claves y el remitente van en variables de entorno; si faltan, el backend no arranca
  y dice qué falta. Ninguna clave en el repositorio ni en los mensajes de error.
- Sin dependencias nuevas (solo net/http).
- Pruebas contra un servidor local que imite al proveedor, y mutaciones a mano para
  comprobar que las pruebas detectan cada fallo.
- Documenta las variables en .env.example, el compose de producción, el README y
  docs/arquitectura.md.
- Prueba un envío real en local con mis claves, sin mostrarlas.
- Commits en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla.
```

## Resultado
- Archivos generados o modificados: `internal/integration/mailer/mailjet.go` (nuevo, con sus pruebas), `factory.go` (opción `MAILER=mailjet`), `config.go` (`MAILJET_API_KEY`, `MAILJET_SECRET_KEY`, `MAIL_FROM`, `MAIL_FROM_NAME`), `cmd/api/integraciones.go` (`correoDesdeConfig`, con su prueba) y `main.go`; además `.env.example`, `docker-compose.prod.yml`, `README.md` y `docs/arquitectura.md`.
- Resumen de lo que hizo: comprobó en el servidor que los puertos 25, 465 y 587 están bloqueados y que el 443 sale, y escribió un adaptador que envía por la API HTTPS de Mailjet (autenticación básica, adjuntos en base64 con su tipo). `MAILER=smtp` ahora explica que el servidor bloquea SMTP y pide usar `mailjet`.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test ./...` del backend completo; mutaciones a mano sobre la autenticación, el destinatario, los adjuntos, la respuesta de error, la validación del mensaje y del remitente, la fábrica y el paso de la configuración, todas detectadas; `docker compose -f docker-compose.prod.yml config`; y un envío real en local: "¿Olvidaste tu contraseña?" para una cuenta de prueba con el correo de Dayron, que llegó a su Gmail.
- Qué se corrigió: primero se escribió un adaptador para Brevo, pero Brevo exige verificar el teléfono para crear la clave de la API y el SMS no llegó a un número de Ecuador; se pasó a Mailjet y se quitó Brevo. Una mutación sobrevivió al principio (el paso de la configuración al adaptador en `main.go` no tenía prueba): se sacó a `correoDesdeConfig` y se le agregó su prueba. La cuenta de prueba local volvió a su correo original después del envío.
