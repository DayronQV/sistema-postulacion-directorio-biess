# 2026-11-01: domicilio con cascada país → provincia → cantón

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code, Sonnet 5.5
- **Issue:** #23 [EF-12] Sección Información domiciliaria con cascada País → Provincia → Cantón

## Prompt
```
Implementa la issue #23 (backend y frontend): lee su texto con gh issue view y sigue su alcance.
Parte de develop en la rama feat/domicilio-cascada.

Backend: validación del domicilio en el servicio, en un archivo nuevo. Si el país es Ecuador,
provincia y cantón son obligatorios por id y coherentes (la provincia pertenece al país y el
cantón a la provincia; hoy el PUT acepta combinaciones incoherentes). Si no es Ecuador,
provincia y ciudad van por texto. Celular de 10 dígitos que empieza en 09, correos válidos,
longitudes máximas. Errores con el formato de docs/api.md. La coherencia se comprueba con el
repositorio de catálogos a través de un puerto de domain/. Pruebas con falsos y de integración
contra PostgreSQL real.

Frontend: Domicilio.tsx y services/catalogos.ts. Select de país; con Ecuador, provincia (24) y
cantón dependientes; con otro país, texto libre. Al cambiar el país se limpian provincia y
cantón; al cambiar la provincia se limpia el cantón. Loader en cada lista. Campos de contacto de
la issue, correo 1 de solo lectura, validaciones con zod y botón Guardar. Componente de
presentación más un hook mínimo que lo conecta a la API. Pruebas con Vitest y Testing Library.

PUT /postulacion reemplaza el estado completo: el frontend lee la postulación, mezcla su sección
y envía el formulario entero. Verificarlo con una prueba de integración. No tocar docs/api.md ni
db/schema.sql: describir en el informe lo que haya que cambiar.
```

## Resultado
- Archivos generados o modificados:
  - Backend: `service/domicilio_validacion.go` y su prueba, `repository/postgres/domicilio_integration_test.go`, y cambios mínimos en `service/postulacion_service.go`, `cmd/api/wire_postulacion.go`, `domain/postulacion.go`, `http/handler/postulacion_handler.go` y `repository/postgres/postulacion_repo.go`.
  - Frontend: `features/postulacion/secciones/Domicilio.tsx` y `DomicilioSeccion.tsx`, `features/postulacion/domicilio.ts`, `useCatalogosGeo.ts`, `useDomicilio.ts`, `services/catalogos.ts`, `services/postulacion.ts` y sus pruebas.
- Resumen de lo que hizo: valida en el servidor que la provincia pertenezca al país y el cantón a la provincia (400 con `fields`), exige provincia y cantón por id con Ecuador y por texto con otro país, y comprueba celular, correos, teléfono de domicilio y longitudes. `GET /postulacion` devuelve además el correo del registro. En el frontend, la sección en cascada con loaders, limpieza de campos dependientes, validaciones con zod y guardado que lee la postulación, cambia solo el domicilio y envía el formulario entero.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet`, `go build`, `go test` y `go test -race` con PostgreSQL 16 y Redis 7 reales y sin pruebas saltadas; `npm run lint`, `npm test` y `npm run build`; recorrido real contra la API (registro, login, `GET` y `PUT /postulacion` con Ecuador válido, cantón de otra provincia, otro país con texto) y comprobación en el navegador de la cascada con la semilla.
- Qué se corrigió: la prueba de integración del rechazo no creaba la postulación antes de leer la base (se corrigió); los selectores de etiqueta de las pruebas del frontend se ajustaron al asterisco de obligatorio. Queda la revisión de quien apruebe el PR.
