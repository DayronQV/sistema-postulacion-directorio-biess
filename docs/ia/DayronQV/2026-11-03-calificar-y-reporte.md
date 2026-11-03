# 2026-11-03: guardar respuestas, finalizar, calificar y reporte PDF

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #44 y #114

## Prompt (versión redactada)
```
Actúa como desarrollador backend en Go.

Necesito terminar la evaluación teórica (EF-14) del lado del postulante y del administrador.

Crea, en las capas del proyecto (handler, servicio y repositorio):
1. Guardar la respuesta de cada pregunta mientras el examen siga abierto; después de la hora
   de cierre, 409.
2. Finalizar y calificar: comparar con las respuestas correctas y guardar la nota.
3. Cierre automático: si el tiempo se agotó, finalizar el examen con las respuestas guardadas
   al consultarlo y con una tarea periódica.
4. Un reporte PDF con diseño (datos del postulante, nota destacada, detalle por pregunta y
   código de verificación con QR) que se envía por correo al finalizar.
5. Resultados para el administrador con la descarga del reporte de cada examen.

Usa la librería maroto para el PDF (issue #114), sin mostrar la respuesta correcta en el reporte.
Agrega pruebas y comprueba que fallan si se quita cada regla.

Las issues #44 y #114 del repositorio tienen el detalle completo (archivos a tocar y criterios
de terminado): revísalas para cumplirlas exactamente.
```

## Para qué se usó
Construir el guardado de respuestas, la calificación, el cierre automático, el reporte PDF y los resultados del administrador, con sus pruebas y la documentación del contrato.
