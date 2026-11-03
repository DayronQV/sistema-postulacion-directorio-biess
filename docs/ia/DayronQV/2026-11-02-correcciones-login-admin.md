# 2026-11-02: correcciones de la revisión del login de administración

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #36 (PR #98)

## Prompt (versión redactada)
```
Revisé el PR #98 (inicio de sesión de administración) y dejé cuatro comentarios.
Resuelve los tres que se pueden cerrar dentro del PR y haz los commits a mi nombre
en la rama del PR:

1. Falta una prueba para el caso de un usuario cuyo rol no es analista ni
   administrador: no debe emitirse token ni contarse como intento fallido.
2. El usuario no tiene largo máximo y el contador de intentos crea claves de Redis
   de tamaño arbitrario: limítalo a 50 caracteres en el servidor y en el formulario.
3. Deja escrito en db/seed_admin.sql que no debe cargarse en producción.

Agrega las pruebas de cada cambio y comprueba que fallan si se quita la corrección.
```

## Para qué se usó
Corregir las observaciones de la revisión del PR #98 con sus pruebas.
