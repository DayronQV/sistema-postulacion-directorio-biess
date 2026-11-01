# 2026-11-01: servicio de correo simulado

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #6

## Prompt (versión redactada)
```
Actúa como desarrollador backend en Go.

Necesito un servicio de correo para recuperar contraseñas, enviar la postulación y
notificar el resultado del examen, sin depender de un servidor de correo mientras se
desarrolla.

Crea:
1. Una interfaz Mailer para enviar un correo con destinatario, asunto, cuerpo en texto
   o HTML y adjuntos opcionales (por ejemplo un PDF).
2. Un adaptador que, en vez de enviar, escriba el correo en el log: destinatario,
   asunto, cuerpo y nombre de los adjuntos.
3. La elección del adaptador con la variable MAILER ("log" por defecto), dejando listo
   el punto para agregar un adaptador SMTP en producción, y MAILER en .env.example.

Valida el correo antes de enviarlo y agrega pruebas.

La issue #6 del repositorio tiene el detalle completo (archivos a tocar y criterios de
terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Construir el servicio de correo simulado (interfaz, adaptador de log y selección por MAILER) con sus pruebas.
