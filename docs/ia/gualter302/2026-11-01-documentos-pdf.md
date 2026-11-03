# Sesión de IA: documentos PDF

## Objetivo

Implementar la issue #25 para subir, descargar y eliminar documentos PDF de la postulación con sus límites de tamaño y pertenencia.

## Solicitud realizada

Avanzar las issues asignadas a `gualter302` que ya tengan sus dependencias fusionadas.

## Trabajo realizado

- Se actualizaron las referencias remotas y se revisaron las issues asignadas.
- Se confirmó que #20 y #8 ya están cerradas, por lo que #25 está desbloqueada.
- Se implementaron los endpoints para subir, descargar y eliminar documentos propios.
- Se validan el contenido `%PDF`, el tipo, el límite por documento y el máximo total de 50 MB.
- Se conectó el selector reutilizable con la API y se añadieron progreso, descarga, reemplazo y eliminación.
- Se agregaron pruebas unitarias y una prueba de integración para impedir descargas de otro postulante.

## Validaciones

- `go test ./...`: correcto (la integración con PostgreSQL se omite si `DATABASE_URL` no está definida).
- `go vet ./...`: correcto.
- `npm run lint`: correcto.
- `npm test`: 14 archivos y 103 pruebas correctas.
- `npm run build`: correcto.

## Decisiones y observaciones

- Las issues #27, #28, #29 y #30 esperan a que #25 sea fusionada.
- La issue #31 está desbloqueada, pero se trabajará en una rama y PR separados.
- Las tareas P2 y de entrega final se mantienen en espera hasta completar P1.
