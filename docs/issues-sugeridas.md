# Issues sugeridas

Propuesta de reparto para que Dayron la ajuste y cree las issues. Cada una debe seguir la plantilla `tarea.md` y ser lo bastante pequeña para una sola persona y su agente. El orden respeta las dependencias.

## Base (bloquea al resto; hacer primero)
| # | Issue | Depende de |
| --- | --- | --- |
| 1 | Andamio del backend: Go + Gin, capas, configuración por entorno, `/health`, middleware de errores y timeout, conexión a PostgreSQL y Redis | |
| 2 | Andamio del frontend: Vite + React + TypeScript + Tailwind, rutas, componentes base (`ui/`), cliente de API con `AbortController` | |
| 3 | Datos semilla de la división política: países, 24 provincias y cantones en `db/seed_catalogos.sql` | |
| 4 | CI con GitHub Actions: pruebas, lint y compilación de backend y frontend en cada PR y push a `main` | 1, 2 |

## P1: EF-11 y EF-12 (lo que califica la rúbrica)
| # | Issue | Depende de |
| --- | --- | --- |
| 5 | Adaptadores simulados: Registro Civil (con cédulas de prueba, una extranjera), Historial Laboral y envío de correo a log | 1 |
| 6 | Backend: verificación de identidad y registro, con bloqueo de 10 min tras 3 fallos en Redis | 1, 5 |
| 7 | Backend: login, recuperar contraseña y JWT con middleware, con bloqueo por intentos | 1, 6 |
| 8 | Backend: catálogos con caché en Redis (TTL) | 1, 3 |
| 9 | Backend: postulación núcleo (datos personales, domicilio, familiares, cargo y elegibilidad), nacionalidad ecuatoriana y guardado transaccional | 1, 5, 7 |
| 10 | Backend: formación, experiencia (con duración en años, meses y días), publicaciones, capacitación, inhabilidades y subida de PDF | 9 |
| 11 | Frontend: registro, login, recuperar contraseña, estado de sesión y ruta protegida | 2 |
| 12 | Frontend: formulario secciones 1 a 4 (cascada País → Provincia → Cantón, familiares, cargo) | 2, 11 |
| 13 | Frontend: formulario secciones 5 a 9, carga de PDF, consentimiento, guardar y enviar | 12 |
| 14 | Pruebas de integración y flujo completo, README final y datos de demostración para la sustentación | 10, 13 |

## P2 y P3 (opcionales, solo si P1 está terminado)
Se detallan en issues propias cuando el grupo decida abordarlos: administración (EF-14) y banco de preguntas con evaluación teórica.

## Pendientes de decisión (ver `docs/especificacion.md`)
Campos de "Formación adicional" y "Referencias personales", y cómo se determina que un postulante superó la Fase de Méritos.
