# 2026-11-03: correcciones de la revisión de las fases del concurso

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #37 (PR #102)

## Prompt (versión redactada)
```
Revisa el PR #102 (gestión de las fases del concurso y periodo de postulación), pruébalo
contra PostgreSQL real y en el navegador, y corrige en la misma rama lo que haga falta,
con los commits a mi nombre.

Encontré dos cosas para corregir:
1. PUT /admin/fases/{id} con un cuerpo vacío responde 200 y cierra el periodo de
   postulación (publicada y habilita_formulario pasan a falso). Los dos interruptores deben
   ser obligatorios.
2. Un postulante que aún no tiene postulación y entra con el periodo cerrado ve "No se
   pudo cargar la postulación" con un botón de reintentar; debe ver un aviso claro, sin
   reintentar.

Agrega las pruebas y comprueba que fallan si se quita cada corrección.
```

## Para qué se usó
Corregir las observaciones de la revisión del PR #102 con sus pruebas.
