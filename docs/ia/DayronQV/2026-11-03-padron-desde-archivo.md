# 2026-11-03: padrón de personas en un archivo fuera del repositorio

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (preparación del despliegue en el VPS, acordada en el grupo)

## Prompt
```
Actúa como desarrollador backend sénior (Go, Gin, arquitectura en capas) preparando el
sistema del concurso del Directorio del BIESS para producción.

Quiero que en el servidor solo puedan crear cuenta las personas que consten en el Registro
Civil y tengan datos del IESS: los cinco integrantes del equipo (con sus cédulas reales) y
cinco personas más con nombres verosímiles. Hoy los simuladores solo conocen 7 cédulas de
prueba escritas en el código.

Requisitos:
- Las cédulas reales no pueden quedar en el repositorio ni en su historial: los datos
  viven en un archivo JSON fuera de git, y una variable de entorno opcional lo indica.
  Sin la variable, todo sigue igual y las pruebas existentes no cambian.
- Los simuladores del Registro Civil y del Historial Laboral leen el mismo archivo.
  Con archivo, una cédula que no consta no tiene ninguna calidad en el IESS.
- Valida el archivo al arrancar (cédula por módulo 10, fechas, código dactilar, estado
  civil con pareja, calidades) y que el error liste todos los problemas sin repetir datos
  personales. Un archivo ausente o inválido detiene el arranque.
- Sin dependencias nuevas ni cambios en los contratos (docs/api.md, db/schema.sql).
- Pruebas para el lector, los dos simuladores, la configuración y el cableado; luego rompe
  el código a propósito, una pieza a la vez, y comprueba que las pruebas lo detectan.
- Comprueba el flujo real (identidad, registro, login y validación del cargo) contra el
  backend, y deja la base local como estaba.
- Commits en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla.
```

## Resultado
- Archivos generados o modificados: paquete nuevo `internal/integration/padron` (lector y validador, con pruebas), constructores `NewDesdeArchivo` en los dos simuladores, `PADRON_ARCHIVO` en `config`, `cmd/api/integraciones.go` (un solo lugar que decide qué simulador crear) y el cableado en `main.go`, `modules.go`, `wire_auth.go` y `wire_postulacion.go`; además `docker-compose.prod.yml` (monta el archivo) y `.env.example`.
- Resumen de lo que hizo: el backend puede cargar las personas de un JSON indicado con `PADRON_ARCHIVO`. Sin él conserva los ciudadanos de prueba. Con él, el Registro Civil y el IESS simulados solo conocen a esas personas, y una cédula ausente no tiene calidad. Los datos reales (padrón y cuentas del panel) se generaron aparte, fuera de git.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go vet` y `go test ./...` del backend completo; siete mutaciones a mano (campos desconocidos, pareja en soltero, cédula repetida, calidad de jubilado, calidad por defecto del IESS, elección del simulador y limpieza de la variable), todas detectadas; el compose de producción con `docker compose config`; y el flujo real contra un segundo backend leyendo el archivo: identidad con fecha y con código dactilar, rechazo de una cédula de prueba vieja, registro, login y validación del cargo correcto, y el ingreso de las cuentas del panel.
- Qué se corrigió: una cédula de prueba antigua dejó de existir en modo archivo, que es lo buscado. Los datos creados para las pruebas se borraron de la base local. Una cuenta del panel parecía fallar con 400 solo por cómo el bucle de shell pasaba los datos; se repitió con un cliente en Python y las cuatro entraron.
