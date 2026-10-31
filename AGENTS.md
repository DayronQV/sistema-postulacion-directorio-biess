# Reglas para agentes y personas

Proyecto: examen práctico de Frameworks de Programación Web. Módulo de registro y postulación del Concurso de Méritos y Oposición del Directorio del BIESS. Trabajo en grupo; cada integrante usa su propio agente de IA. Estas reglas valen para todos.

## Orden de lectura (y nada más)
1. Este archivo.
2. `docs/especificacion.md`: qué se construye y qué NO.
3. `docs/arquitectura.md` y `docs/api.md`: cómo se construye.
4. La issue asignada.

No leas ni resumas documentos largos del repositorio fuera de esa lista. Todo lo necesario ya está condensado arriba; leer de más cuesta tokens y no mejora el resultado.

## Alcance
- **P1 es obligatorio:** EF-11 y EF-12 completos (`docs/especificacion.md`). Es lo único que califica la rúbrica del examen; se termina, se prueba y se pule antes de tocar otra cosa.
- **P2 y P3 son opcionales y van después de P1, en ese orden:** módulo de administración y banco de preguntas con evaluación teórica. Se arrancan solo si P1 está terminado y el grupo lo decide.
- Fuera de alcance: integraciones reales (Registro Civil, Historial Laboral, correo, firma electrónica); se simulan.
- Trabajá solo en lo que dice tu issue. Si ves algo que arreglar fuera de ella, abrí otra issue; no lo cambies.
- Sin dependencias nuevas, sin cambiar el stack y sin modificar los contratos (`docs/api.md`, `db/schema.sql`) sin una issue que lo apruebe.
- Si algo es ambiguo, preguntá a la persona antes de suponer.

## Stack fijo
- **Frontend:** React + TypeScript + Vite + Tailwind CSS (SPA).
- **Backend:** Go con Gin, en capas (handler → service → repository) con inyección de dependencias.
- **Datos:** PostgreSQL 16 (transacciones `BEGIN`/`COMMIT`/`ROLLBACK`) y Redis 7.
- **Entorno local:** `docker-compose.yml` levanta PostgreSQL y Redis.

## Reglas de código
- Backend: los handlers no contienen lógica de negocio; los servicios no conocen HTTP ni SQL; los repositorios son interfaces inyectadas.
- Validación siempre en el servidor. Respuestas de error con el formato de `docs/api.md`.
- Las integraciones externas (Registro Civil, correo, Historial Laboral) se simulan detrás de una interfaz con un adaptador `mock`.
- Frontend: componentes pequeños y reutilizables en `components/`; llamadas a la API solo desde `services/`; estados de carga, éxito y error visibles en cada formulario.
- Sin secretos en el repositorio: usar `.env` (ignorado) y `.env.example`.
- Cada tarea trae sus pruebas (backend con `go test`, frontend con Vitest).

## Flujo de trabajo
- Una rama por issue desde `main`: `feat/<tema>`, `fix/<tema>`, `docs/<tema>`, `test/<tema>`.
- Commits pequeños con Conventional Commits en inglés (`feat: add login endpoint`).
- No se hace push directo a `main`: todo entra por Pull Request con la plantilla del repo, enlazando la issue.
- Un PR toca solo los archivos de su issue. Si hay conflicto, se resuelve en la rama del PR.

## Registro de uso de IA (obligatorio)
Cada sesión con IA deja un archivo en `docs/ia/<usuario-github>/AAAA-MM-DD-tema.md`, con la plantilla de `docs/ia/plantilla.md`. Esa carpeta se versiona y nunca va en `.gitignore`.

## Barra de calidad (rúbrica del examen, 40 puntos)
| Criterio | Pts | Qué se evidencia |
| --- | --- | --- |
| Arquitectura backend e inyección de dependencias | 8 | Capas sin acoplamiento directo, errores bien manejados |
| PostgreSQL y transacciones | 8 | Modelo normalizado; guardado de la postulación en una transacción multi-tabla |
| Redis | 6 | Bloqueo de 10 min tras 3 intentos fallidos y caché de catálogos con TTL |
| Frontend SPA | 6 | Componentes modulares, estado sólido, responsive, loaders y validaciones |
| Consumo de API | 6 | `async/await`, rutas protegidas, estados loading/success/error |
| Reglas de negocio | 6 | Bloqueo de intentos, validación dactilar simulada, nacionalidad ecuatoriana, cascada país/provincia/cantón |

Además se entrega: `README.md` (prerrequisitos y ejecución), `docker-compose.yml`, `.env.example` y el script `.sql` con DDL y datos semilla de la división política. Habrá sustentación de 10 minutos: cada quien debe poder explicar el código que su agente escribió.
