# Sesión de IA: formación académica

## Objetivo

Implementar la issue #27 para registrar la formación académica y su respaldo PDF.

## Solicitud realizada

Avanzar las issues P1 desbloqueadas, creando directamente el PR de la primera y verificando todas las pruebas.

## Trabajo realizado

- Se confirmó que #20 y #25 están cerradas.
- Se inició la implementación desde el `develop` actualizado.
- Se añadió el modelo, validaciones y repositorio transaccional de formación académica.
- Se comprueba que el respaldo sea propio y de tipo `formacion`.
- Se creó la sección frontend con agregar, editar, eliminar, carga PDF y guardado sin borrar otras secciones.
- Los respaldos reemplazados o eliminados se limpian después de actualizar sus referencias.

## Validaciones

- `go test ./...`: correcto (las pruebas PostgreSQL se ejecutan cuando `DATABASE_URL` está definida).
- `go vet ./...`: correcto.
- `npm run lint`: correcto.
- `npm test`: 32 archivos y 255 pruebas correctas.
- `npm run build`: correcto.

## Decisiones y observaciones

- La fecha SENESCYT usa el formato `AAAA-MM-DD` y no puede ser futura.
- Las filas se reemplazan junto con el resto del formulario dentro de la misma transacción.
