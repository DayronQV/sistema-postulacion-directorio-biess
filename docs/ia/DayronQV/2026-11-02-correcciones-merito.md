# 2026-11-02: correcciones de la revisión de la marca de méritos

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #38 (PR #99)

## Prompt (versión redactada)
```
Revisé el PR #99 (marcar quién superó la Fase de Méritos) y dejé cuatro comentarios.
Resuélvelos y haz los commits a mi nombre en la rama del PR:

1. La validación del UUID del servicio no está probada: quitarla no hace fallar ninguna
   prueba, y con PostgreSQL real daría un 500 en lugar de un 404.
2. La condición de estado del UPDATE del repositorio no tiene prueba: agrega una prueba
   de integración con PostgreSQL real, como las de los demás repositorios.
3. El componente AccionMerito no sigue a su propiedad si el padre recarga la fila, y
   deja un espacio doble en la confirmación cuando no hay código.
4. El aviso al postulante promete una fecha y hora que nadie envía: cambia el texto.

Agrega las pruebas y comprueba que fallan si se quita cada corrección.
```

## Para qué se usó
Corregir las observaciones de la revisión del PR #99 con sus pruebas.
