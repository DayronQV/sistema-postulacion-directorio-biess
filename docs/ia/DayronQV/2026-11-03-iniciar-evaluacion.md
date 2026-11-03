# 2026-11-03: iniciar la evaluación con sorteo aleatorio

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #43

## Prompt (versión redactada)
```
Actúa como desarrollador backend en Go.

Necesito que el postulante pueda iniciar la evaluación teórica (EF-14) y ver sus preguntas
una por una.

Crea, en las capas del proyecto (handler, servicio y repositorio):
1. Iniciar la evaluación: solo si superó la Fase de Méritos y dentro de la fecha y hora
   configuradas; si no, 403.
2. Sorteo con crypto/rand de las preguntas del banco que pide la configuración, con orden
   aleatorio de preguntas y de opciones para cada postulante.
3. Crear el intento con hora de cierre igual al inicio más la duración. Un solo intento
   por postulante: si ya existe y sigue vigente, devolver el mismo.
4. Mostrar la pregunta n sin la respuesta correcta y con el tiempo restante calculado por
   el servidor.

Agrega pruebas (incluido que dos postulantes reciben exámenes distintos) y comprueba que
fallan si se quita cada regla.

La issue #43 del repositorio tiene el detalle completo (archivos a tocar y criterios de
terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Construir el inicio de la evaluación con su sorteo y la consulta de preguntas, con sus pruebas, y documentar el contrato en `docs/api.md`.

## Segundo prompt (versión redactada)
```
El PR de las fases del concurso trajo la fase Evaluación, con publicada, fecha de inicio y
fecha de fin, pero el examen no la usa: lee su propia fecha de inicio de la configuración.
Haz que el examen use la fase Evaluación: solo se puede iniciar con la fase publicada y entre
sus fechas, y la hora de cierre de cada examen es la menor entre su inicio más la duración y la
fecha de fin de la fase. La configuración conserva la duración y el número de preguntas.
Agrega las pruebas y comprueba que fallan si se quita cada regla.
```
