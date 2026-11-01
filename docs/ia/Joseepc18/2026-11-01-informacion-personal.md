# 2026-11-01: sección Información personal y bloqueo por nacionalidad

- **Autor (usuario de GitHub):** Joseepc18
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #22

## Prompt
```
Implementa la issue #22 "[EF-12] Sección Información personal y bloqueo por nacionalidad", backend y
frontend, repitiendo el patrón de las secciones de domicilio y familiares (componente de presentación
con props y un hook que lo conecta a la API, montable por el contenedor del formulario).
Backend: la regla de nacionalidad de la issue #20 debe cumplirse en el servidor: la cédula extranjera de
prueba 0103456786 queda bloqueada y no puede continuar llamando a la API directamente (PUT 422 la primera
vez y 409 después; GET devuelve la postulación cerrada). Hacer robusta la comparación de la nacionalidad
(espacios, mayúsculas y tildes). Si el estado civil es casado/a o unión de hecho, la respuesta incluye los
datos de la pareja. Pruebas con falsos y de integración contra PostgreSQL real.
Frontend: InformacionPersonal.tsx de solo lectura (nombres, apellidos, cédula, género, fecha de nacimiento
en DD/MM/AAAA, nacionalidad y estado civil), pareja solo si corresponde, botón "Guardar" que envíe la
postulación entera (el PUT reemplaza el estado completo), pantalla de bloqueo cuando está cerrada y estados
de carga y error. Pruebas con Vitest y Testing Library.
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/nacionalidad.go` y `nacionalidad_test.go` (comparación normalizada)
  - `backend/internal/service/postulacion_service.go` y `postulacion_nacionalidad_test.go`
  - `backend/internal/repository/postgres/postulacion_repo.go` y `informacion_personal_integration_test.go`
  - `backend/internal/http/handler/postulacion_handler.go` y `postulacion_personal_test.go`
  - `frontend/src/features/postulacion/informacionPersonal.ts` (+ test), `useInformacionPersonal.ts`
  - `frontend/src/features/postulacion/secciones/InformacionPersonal.tsx`, `InformacionPersonalSeccion.tsx`, `PostulacionBloqueada.tsx` (+ pruebas)
  - `frontend/src/services/postulacion.ts` (campo `cedula`) y una línea en `DomicilioSeccion.test.tsx`
- Resumen de lo que hizo: la nacionalidad se compara sin distinguir mayúsculas, tildes ni espacios sobrantes (y acepta "Ecuatoriano"). Si la persona no es ecuatoriana, `PUT /postulacion` omite la validación del cuerpo para que un formulario inválido no oculte el cierre: la primera llamada responde 422 y cierra la postulación; las siguientes, 409. La respuesta incluye ahora la `cedula` en `informacion_personal`; los datos de la pareja ya los exponían el simulador y `GET /postulacion`. En el frontend, la sección es de solo lectura, muestra la pareja solo si es casado/a o unión de hecho, guarda la postulación entera releída en el momento del guardado y muestra la pantalla de bloqueo cuando la postulación está cerrada o la nacionalidad no es ecuatoriana.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet`, `go build`, `go test` y `go test -race` en verde con PostgreSQL 16 y Redis 7 reales (397 pruebas, 0 saltadas). En el frontend, `npm run lint`, `npm test` (222 pruebas) y `npm run build`. Recorrido real contra la API: registro y login con 0103456786 (extranjera) y 0912345675 (casada). La extranjera recibe 422 en el primer `PUT` (con un cuerpo inválido y sin `GET` previo), 409 en el siguiente y `GET` devuelve `estado: cerrada`; la casada recibe los datos de la pareja y puede guardar.
- Qué se corrigió: la regla de la issue #20 validaba el cuerpo antes de comprobar la nacionalidad, así que un extranjero con un formulario inválido recibía 400 y no se cerraba; ahora se comprueba primero. La respuesta no traía la cédula, que la sección necesita mostrar.
