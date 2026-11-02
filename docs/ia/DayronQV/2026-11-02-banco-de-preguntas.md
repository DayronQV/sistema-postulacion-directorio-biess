# 2026-11-02: banco de preguntas y configuración de la evaluación

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #42

## Prompt (versión redactada)
```
Actúa como desarrollador backend en Go.

Necesito que el administrador pueda cargar el banco de preguntas de opción múltiple y
configurar la evaluación teórica (EF-14).

Crea, en las capas del proyecto (handler, servicio y repositorio):
1. Crear preguntas (enunciado, opciones e índice de la respuesta correcta), una a una y
   en bloque desde JSON o CSV. Una carga en bloque guarda todas o ninguna y reporta los
   errores por fila.
2. Listar (con búsqueda y paginación), editar y eliminar preguntas, solo para el
   administrador.
3. Configurar fecha y hora de inicio, duración y número de preguntas por examen. El
   número de preguntas no puede superar las del banco.
4. Las respuestas correctas nunca se envían al postulante.

Agrega pruebas, incluida una carga masiva de al menos 20 preguntas, y comprueba que
fallan si se quita cada regla.

La issue #42 del repositorio tiene el detalle completo (archivos a tocar y criterios de
terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Construir el banco de preguntas y la configuración de la evaluación teórica con sus pruebas, y actualizar el contrato en `docs/api.md`.
