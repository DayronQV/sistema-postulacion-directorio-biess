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
| `POST /postulacion/enviar` | Valida que esté completa, la marca `enviada` (guarda `enviada_at`) y manda el formulario al correo del postulante, en una transacción. Devuelve la postulación. `422` con `fields` por sección incompleta (`domicilio`, `familiares` y `publicaciones` si se respondió Sí sin agregar filas, `cargo`, `formacion`, `experiencia`, `capacitaciones`, `referencias`, `inhabilidades`, `consentimiento`); `409` si ya fue enviada o está cerrada. |

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
  "referencias": [
    { "nombres_apellidos": "", "relacion": "", "telefono": "0999999999", "correo": "persona@correo.com" }
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
`cargo` ∈ `afiliados | jubilados | empleador`. En `GET /postulacion` cada experiencia incluye `duracion: { anios, meses, dias }` calculada en el servidor, y la postulación incluye `duracion_experiencia_total: { anios, meses, dias }` (suma de las experiencias; 30 días forman un mes). Ambos son de solo lectura: `PUT` los ignora. `422` con el mensaje del requerimiento si el cargo no es elegible o la nacionalidad no es ecuatoriana.

**Inhabilidades.** `inhabilidades` lleva las claves `"1"` a `"13"` con `true` (Sí) o `false` (No); una pregunta sin responder no aparece, así el borrador puede guardarse a medias (las 13 son obligatorias al enviar). Si la `7` es `true`, `removido_detalle` exige `institucion` y `motivo` (`cargo` y `fecha_fin` son opcionales; `fecha_fin` en `AAAA-MM-DD`) y, si no, se descarta; con error `400` y `fields` como `inhabilidades.removido_detalle.motivo`. La respuesta incluye `elegible` (solo lectura, `PUT` lo ignora): `true` solo cuando las 13 están respondidas, la `1` es Sí y de la `2` a la `13` son No. Se guarda en la columna `postulaciones.elegible` y no descalifica automáticamente: la revisión la hace el analista (EF-13).

**Fase de Méritos.** La respuesta incluye `supero_merito` (lo marca el administrador con `PUT /admin/postulaciones/{id}/merito`) y `estado_visible`, el estado que ve el postulante: `Borrador`, `Enviada`, `Cerrada` o `Evaluación Técnica` (enviada y con la Fase de Méritos superada). Ambos son de solo lectura: `PUT` los ignora.

## Administración (requiere rol `analista` o `administrador`)
| Método y ruta | Descripción |
| --- | --- |
| `POST /admin/auth/login` | `{ usuario, password }` → `{ token, expires_in, usuario, rol }` (`expires_in` en segundos; mismo bloqueo por intentos: `401` "sus credenciales de ingreso son incorrectas", `429 LOCKED` al tercer fallo). Público |
| `GET /admin/sesion` | `{ usuario, rol }` del token. Todas las rutas `/admin/*` salvo el login exigen token de `analista` o `administrador` (`401` sin token, `403` con token de postulante); las de escritura, solo `administrador` (`403` para el analista) |
| `GET /admin/fases` · `PUT /admin/fases/{id}` | Ver y publicar las fases; habilitar el formulario EF-12 |
| `GET /admin/dashboard` | `{ por_cargo: { afiliados, jubilados, empleador }, por_estado: { borrador, enviada, cerrada }, total }` |
| `GET /admin/postulaciones?cargo=&estado=` | Nómina de postulantes, paginada |
| `GET /admin/postulaciones/{id}` | Formulario íntegro, con duración de experiencias |
| `GET /admin/postulaciones/{id}/documentos/{docId}` | Descarga de certificados |
| `GET /admin/matriz?formato=xlsx\|pdf` | Cuadro comparativo de postulantes |
| `PUT /admin/postulaciones/{id}/merito` | `{ supero: bool }` → `{ id, supero_merito, estado_visible }`. Marca si superó la Fase de Méritos. Solo `administrador` (`403` para el analista) y solo postulaciones `enviada` (`409` si no; `404` si no existe). |

## Evaluación teórica
| Método y ruta | Rol | Descripción |
| --- | --- | --- |
| `POST /admin/evaluacion/preguntas` | admin | Carga el banco (opción múltiple y respuesta correcta): una pregunta, o en bloque con JSON o CSV |
| `GET /admin/evaluacion/preguntas?buscar=&pagina=&por_pagina=` | admin | Lista el banco, con la respuesta correcta |
| `PUT /admin/evaluacion/preguntas/{id}` · `DELETE /admin/evaluacion/preguntas/{id}` | admin | Edita o elimina una pregunta |
| `GET /admin/evaluacion/config` · `PUT /admin/evaluacion/config` | admin | Fecha y hora, duración y número de preguntas |
| `POST /evaluacion/iniciar` | postulante | Sortea las preguntas únicas del postulante y fija la hora de cierre |
| `GET /evaluacion/pregunta/{n}` · `PUT /evaluacion/respuesta/{n}` | postulante | Mostrar y guardar respuesta, una por una |
| `POST /evaluacion/finalizar` | postulante | Califica, guarda la nota, genera el PDF y notifica |
| `GET /admin/evaluacion/resultados` | admin | Notas y reportes PDF |

El servidor es la autoridad del tiempo: al llegar a la hora de cierre, `finalizar` se ejecuta con las respuestas guardadas.

### Banco de preguntas y configuración (solo `administrador`)
Todas las rutas `/admin/evaluacion/*` son solo para el administrador (`403` para el analista, también al leer: el listado trae la respuesta correcta). Las respuestas correctas **nunca** se envían al postulante.

Una pregunta es `{ enunciado, opciones: [texto, ...], respuesta_correcta }`, con `respuesta_correcta` el **índice desde 0** de la opción correcta. Reglas: enunciado hasta 1000 caracteres; de 2 a 8 opciones, no vacías, sin repetir y de hasta 500 caracteres; índice dentro de las opciones.

- `POST /admin/evaluacion/preguntas`, según el `Content-Type`:
  - `application/json` con una pregunta → `201` con la pregunta creada (`{ id, enunciado, opciones, respuesta_correcta }`).
  - `application/json` con `{ "preguntas": [ ... ] }` → carga en bloque, `201 { creadas }`. Máximo 500 por carga.
  - `text/csv`, o `multipart/form-data` con el campo `archivo` → carga en bloque desde un CSV de hasta 1 MB, `201 { creadas }`. Columnas `enunciado,opciones,respuesta_correcta` (en cualquier orden, con cabecera); las opciones van en una sola celda separadas por una barra vertical. Acepta `;` como separador (Excel en español) y BOM de UTF-8. El archivo debe estar en UTF-8 (en Excel, "CSV UTF-8"); otra codificación es `400` en `fields.archivo`. Ejemplo de una fila: `"¿Capital de Ecuador?",Guayaquil|Quito|Cuenca,1`.
  - Una carga en bloque guarda **todas las preguntas o ninguna**. Si hay errores, `400` con un error por posición en `fields`: `preguntas[3].opciones[1]` (JSON) o `filas[4].respuesta_correcta` (CSV, con el número de línea del archivo; la cabecera es la línea 1). Se listan hasta 50 errores. Otro `Content-Type` es `400`; un archivo de más de 1 MB, `422`.
- `GET /admin/evaluacion/preguntas?buscar=&pagina=1&por_pagina=20` → `{ items: [ { id, enunciado, opciones, respuesta_correcta } ], total, pagina, por_pagina }`, ordenado por id. `buscar` filtra por texto del enunciado sin distinguir mayúsculas; `por_pagina` va de 1 a 100.
- `PUT /admin/evaluacion/preguntas/{id}` (una pregunta completa) → `200` con la pregunta; `404` si no existe. `DELETE /admin/evaluacion/preguntas/{id}` → `204`; `404` si no existe. Ninguna de las dos se puede una vez que algún postulante inició la evaluación (`409`), y no se puede eliminar una pregunta si el banco quedaría con menos de las que pide la configuración (`409`).
- `GET /admin/evaluacion/config` → `{ inicia_en, duracion_minutos, num_preguntas, total_banco, configurada }`. Mientras no se configure: `inicia_en: null`, 30 minutos, 10 preguntas y `configurada: false`.
- `PUT /admin/evaluacion/config` con `{ inicia_en, duracion_minutos, num_preguntas }` → `200` con la misma forma. `inicia_en` en RFC 3339 con zona horaria (por ejemplo `2026-11-20T09:00:00-05:00`) y siempre se responde en UTC; la duración va de 1 a 480 minutos. `422` si `num_preguntas` supera las preguntas del banco (`fields.num_preguntas` dice cuántas hay).

## Utilidad
`GET /health` → `200 { "status": "ok" }` (sin prefijo `/api/v1`).
