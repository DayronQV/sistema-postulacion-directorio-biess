# 2026-11-03: quitar inicia_en de la configuración de la evaluación

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #111

## Prompt (versión redactada)
```
Desde la issue #108 el horario del examen lo fija la fase Evaluación de la pantalla de
fases, así que el campo inicia_en de la configuración de la evaluación ya no hace nada,
pero sigue siendo obligatorio al configurar.

Quítalo: la columna de eval_config, el dominio, el repositorio, el servicio y el handler.
La configuración queda solo con la duración y el número de preguntas; si llega un
inicia_en, se ignora. Actualiza docs/api.md y las pruebas, y comprueba que el examen sigue
tomando el horario de la fase Evaluación.

La issue #111 del repositorio tiene el detalle completo (archivos a tocar y criterios de
terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Quitar el campo de fecha que quedó sin uso en la configuración del examen, con sus pruebas y la documentación del contrato.
