# Contrato de la API (v1)

Base: `/api/v1`. JSON en ambos sentidos. Cambiar este archivo requiere una issue y un PR aprobado.

## Formato de error
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Datos inválidos",
    "fields": { "cedula": "Cédula no válida" },
    "retry_after_seconds": 600
  }
}
```
`fields` y `retry_after_seconds` aparecen solo cuando aplican. Códigos de `code`: `VALIDATION_ERROR` (400/422), `UNAUTHORIZED` (401), `FORBIDDEN` (403), `NOT_FOUND` (404), `CONFLICT` (409), `LOCKED` (429), `INTERNAL_ERROR` (500).

## Autenticación
| Método y ruta | Cuerpo | Respuesta |
| --- | --- | --- |
| `POST /auth/identidad` | `{ cedula, fecha_expedicion? , codigo_dactilar? }` (uno de los dos) | `200 { ticket }`; `401` si no coincide; `429 LOCKED` tras 3 fallos |
| `POST /auth/registro` | `{ ticket, email, password }` | `201 { id }`; `409` si la cédula o el correo ya existen |
| `POST /auth/login` | `{ cedula, password }` | `200 { token, expires_in, primer_ingreso }` (`expires_in` en segundos; `primer_ingreso` es `true` si aún no tiene postulación); `401` si es incorrecto; `429 LOCKED` tras 3 fallos |
| `POST /auth/recuperar` | `{ cedula, fecha_expedicion? , codigo_dactilar? }` (uno de los dos) | `202` sin cuerpo si la identidad coincide, tenga o no cuenta la cédula (no se revela quién está registrado); si tiene cuenta, se envía al correo registrado (simulado en el log) el enlace `<frontend>/restablecer?token=...`, válido 30 min y de un solo uso. `401` si no coincide; `429 LOCKED` tras 3 fallos |
| `POST /auth/restablecer` | `{ token, password }` | `204`; cambia la contraseña y borra el token y los bloqueos de login. `400` si la contraseña no cumple la regla del registro; `401` si el token no existe, venció o ya se usó |

## Catálogos (públicos, con caché en Redis)
| Método y ruta | Respuesta |
| --- | --- |
| `GET /catalogos/paises` | `[ { id, nombre } ]` |
| `GET /catalogos/paises/{id}/provincias` | `[ { id, nombre } ]` |
| `GET /catalogos/provincias/{id}/cantones` | `[ { id, nombre } ]` |

## Postulación (requiere `Authorization: Bearer <token>`)
| Método y ruta | Descripción |
| --- | --- |
| `GET /postulacion` | Devuelve la postulación del usuario, incluida la información personal precargada. Si no existe, la crea en estado `borrador`. |
| `PUT /postulacion` | Guarda el progreso. Persiste en **una transacción**. `422` si la nacionalidad no es ecuatoriana (cierra la postulación). |
| `POST /postulacion/enviar` | Valida que esté completa y la marca `enviada`. `409` si ya fue enviada. |

Las secciones con filas (familiares, formación, experiencia, publicaciones, capacitación) se guardan en el mismo `PUT /postulacion` como arreglos; `PUT` reemplaza el estado completo en una transacción. Los PDF se suben aparte y se referencian por `documento_id`:

| Método y ruta | Descripción |
| --- | --- |
| `POST /postulacion/documentos` | `multipart/form-data` con `archivo` (PDF) y `tipo` (`organizacion`, `formacion`, `experiencia`, `publicacion`, `capacitacion`). Devuelve `201 { documento_id, nombre, tamano_bytes }`. `422` si no es PDF o excede el límite del tipo (4 MB organización, 2 MB el resto) o los 50 MB totales. |
| `GET /postulacion/documentos/{id}` | Descarga el PDF propio. `404` si no existe o es de otro postulante. |
| `DELETE /postulacion/documentos/{id}` | Elimina un PDF propio únicamente si todavía no está referenciado por una sección. `204` si se eliminó, `409` si una sección lo usa, `404` si no existe o es de otro postulante. |

Forma de `PUT /postulacion`:
```json
{
  "domicilio": {
    "pais_id": 1, "provincia_id": 17, "canton_id": 170,
    "provincia_texto": null, "ciudad_texto": null,
    "telefono_celular": "0999999999", "email_secundario": "otro@correo.com",
    "calle_principal": "", "numero": "", "calle_secundaria": "",
    "sector": "", "referencia": "", "telefono_domicilio": ""
  },
  "tiene_familiares": true,
  "familiares": [
    { "nombres_apellidos": "", "parentesco": "hermano/a", "institucion": "IESS", "area": "", "cargo": "" }
  ],
  "cargo": "afiliados",
  "respaldo_organizacion": { "nombre": "", "documento_id": null },
  "formacion": [
    { "nivel": "tercer_nivel", "institucion": "", "tipo": "nacional", "titulo": "", "area": "finanzas",
      "senescyt_numero": "", "senescyt_fecha": "2020-01-31", "documento_id": "" }
  ],
  "experiencia": [
    { "tipo": "general", "cargo": "", "institucion": "", "fecha_inicio": "2018-01-01", "fecha_fin": "2020-01-01",
      "actividades": "finanzas", "descripcion": "", "documento_id": "" }
  ],
  "tiene_publicaciones": false,
  "publicaciones": [],
  "capacitaciones": [
    { "tipo": "curso", "nombre": "", "institucion": "", "fecha_inicio": "", "fecha_fin": "", "horas": 0, "documento_id": "" }
  ],
  "inhabilidades": {
    "1": false, "2": false, "3": false, "4": false, "5": false, "6": false, "7": false,
    "8": false, "9": false, "10": false, "11": false, "12": false, "13": false,
    "removido_detalle": { "institucion": "", "cargo": "", "fecha_fin": "", "motivo": "" }
  },
  "consentimiento": true
}
```
`cargo` ∈ `afiliados | jubilados | empleador`. En `GET /postulacion` cada experiencia incluye `duracion: { anios, meses, dias }` calculada en el servidor. `422` con el mensaje del requerimiento si el cargo no es elegible o la nacionalidad no es ecuatoriana.

## Administración (requiere rol `analista` o `administrador`)
| Método y ruta | Descripción |
| --- | --- |
| `POST /admin/auth/login` | `{ usuario, password }` → `{ token }` (mismo bloqueo por intentos) |
| `GET /admin/fases` · `PUT /admin/fases/{id}` | Ver y publicar las fases; habilitar el formulario EF-12 |
| `GET /admin/dashboard` | `{ por_cargo: { afiliados, jubilados, empleador } }` |
| `GET /admin/postulaciones?cargo=&estado=` | Nómina de postulantes, paginada |
| `GET /admin/postulaciones/{id}` | Formulario íntegro, con duración de experiencias |
| `GET /admin/postulaciones/{id}/documentos/{docId}` | Descarga de certificados |
| `GET /admin/matriz?formato=xlsx\|pdf` | Cuadro comparativo de postulantes |
| `PUT /admin/postulaciones/{id}/merito` | Marca si superó la Fase de Méritos |

## Evaluación teórica
| Método y ruta | Rol | Descripción |
| --- | --- | --- |
| `POST /admin/evaluacion/preguntas` | admin | Carga el banco (opción múltiple y respuesta correcta) |
| `PUT /admin/evaluacion/config` | admin | Fecha y hora, duración y número de preguntas |
| `POST /evaluacion/iniciar` | postulante | Sortea las preguntas únicas del postulante y fija la hora de cierre |
| `GET /evaluacion/pregunta/{n}` · `PUT /evaluacion/respuesta/{n}` | postulante | Mostrar y guardar respuesta, una por una |
| `POST /evaluacion/finalizar` | postulante | Califica, guarda la nota, genera el PDF y notifica |
| `GET /admin/evaluacion/resultados` | admin | Notas y reportes PDF |

El servidor es la autoridad del tiempo: al llegar a la hora de cierre, `finalizar` se ejecuta con las respuestas guardadas.

## Utilidad
`GET /health` → `200 { "status": "ok" }` (sin prefijo `/api/v1`).
