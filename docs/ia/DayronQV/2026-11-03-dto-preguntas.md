# 2026-11-03: DTO propio para las preguntas del banco

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #105

## Prompt (versión redactada)
```
En la revisión del PR del banco de preguntas me señalaron que domain.Pregunta y
domain.PreguntaPublica llevan etiquetas JSON y el handler las serializa directamente,
cuando en el resto del proyecto el dominio no conoce el formato JSON y cada handler arma
su propio DTO.

Corrígelo: crea un DTO en el handler del administrador para la respuesta de las
preguntas, quita las etiquetas JSON del dominio y ajusta las pruebas. Comprueba que las
pruebas fallan si el dominio vuelve a llevar etiquetas o el handler lo serializa directo.
```

## Para qué se usó
Mantener las capas del backend como en las demás secciones: el dominio sin etiquetas JSON y la forma de la respuesta definida en el handler.
