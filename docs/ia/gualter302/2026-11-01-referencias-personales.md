# Sesión de IA: referencias personales

## Objetivo

Implementar la issue #31 dentro del guardado transaccional de la postulación.

## Solicitud realizada

Avanzar las issues asignadas a `gualter302` cuyas dependencias ya estén fusionadas.

## Trabajo realizado

- Se confirmó que la única dependencia, #20, está cerrada.
- Se añadió el contrato, el esquema, el dominio y el repositorio transaccional de referencias.
- Se añadieron validaciones de cantidad, teléfono y correo.
- Se creó una sección reutilizable del frontend para agregar, editar, eliminar y guardar.

## Validaciones

- `go test ./...`: correcto (la integración con PostgreSQL se omite si `DATABASE_URL` no está definida).
- `go vet ./...`: correcto.
- `npm run lint`: correcto.
- `npm test`: 14 archivos y 102 pruebas correctas.
- `npm run build`: correcto.

## Decisiones y observaciones

- La lista puede estar vacía mientras la postulación es borrador para respetar el guardado por partes.
- El envío final deberá exigir al menos una referencia.
