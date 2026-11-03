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

**Periodo de postulación.** La respuesta incluye `periodo_abierto` (solo lectura): `true` si la fase Postulación está publicada, con el formulario habilitado y dentro de sus fechas. Fuera del periodo, crear la postulación (`GET /postulacion` cuando aún no existe), `PUT /postulacion`, `POST /postulacion/cargo/validar`, `POST /postulacion/enviar`, `POST /postulacion/documentos` y `DELETE /postulacion/documentos/{id}` responden `403` "El periodo de postulación no está habilitado"; leer una postulación existente y descargar sus documentos sigue funcionando.

## Administración (requiere rol `analista` o `administrador`)
| Método y ruta | Descripción |
| --- | --- |
| `POST /admin/auth/login` | `{ usuario, password }` → `{ token, expires_in, usuario, rol }` (`expires_in` en segundos; mismo bloqueo por intentos: `401` "sus credenciales de ingreso son incorrectas", `429 LOCKED` al tercer fallo). Público |
| `GET /admin/sesion` | `{ usuario, rol }` del token. Todas las rutas `/admin/*` salvo el login exigen token de `analista` o `administrador` (`401` sin token, `403` con token de postulante); las de escritura, solo `administrador` (`403` para el analista) |
| `GET /admin/fases` | `[{ id, nombre, orden, publicada, habilita_formulario, fecha_inicio, fecha_fin }]` en orden (`Postulación`, `Méritos`, `Evaluación`). Las fechas son RFC 3339 o `null` (sin límite por ese lado) |
| `PUT /admin/fases/{id}` | `{ publicada, habilita_formulario, fecha_inicio, fecha_fin }` → la fase actualizada. Reemplaza el estado de la fase: `publicada` y `habilita_formulario` son **obligatorios** (un cuerpo vacío o incompleto es `400`, para no cerrar el periodo sin querer) y una fecha ausente o `null` significa sin fecha. Solo `administrador` (`403` para el analista); `404` si no existe. `400` con `fields`: `publicada` y `habilita_formulario` obligatorios; `habilita_formulario` solo en la fase Postulación y con `publicada: true`; `fecha_fin` no anterior a `fecha_inicio`. Habilitar la fase Postulación abre el formulario EF-12 (ver *Periodo de postulación*) |
| `GET /admin/dashboard` | `{ por_cargo: { afiliados, jubilados, empleador }, por_estado: { borrador, enviada, cerrada }, total }`. Cuenta todas las postulaciones: un borrador sin cargo elegido suma al `total` pero a ningún cargo |
| `GET /admin/postulaciones?cargo=&estado=&elegible=&pagina=` | Nómina paginada (`20` por página): `{ items, total, pagina, por_pagina }`. Cada fila incluye `id`, `codigo`, `cedula`, `nombres`, `apellidos`, `cargo`, `estado` y `elegible` |
| `GET /admin/postulaciones/{id}` | Formulario íntegro, con duración individual y total de experiencias |
| `GET /admin/postulaciones/{id}/documentos/{docId}` | Descarga un certificado solo si pertenece a esa postulación; en otro caso, `404` |
| `GET /admin/matriz?formato=xlsx\|pdf&cargo=&estado=&elegible=` | Descarga el cuadro comparativo en Excel o PDF. Los filtros son opcionales y usan las mismas reglas de la nómina |
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
| `GET /admin/evaluacion/resultados` · `GET /admin/evaluacion/resultados/{id}/reporte` | admin | Notas de todos los postulantes y descarga del reporte PDF de cada examen |

El servidor es la autoridad del tiempo: al llegar a la hora de cierre, el examen se finaliza solo con las respuestas guardadas (ver *Respuestas, calificación y reporte*).

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
- `GET /admin/evaluacion/config` → `{ duracion_minutos, num_preguntas, total_banco, configurada }`. Mientras no se configure: 30 minutos, 10 preguntas y `configurada: false`. **Cuándo se puede rendir el examen no está aquí**: lo fija la fase Evaluación (`/admin/fases`).
- `PUT /admin/evaluacion/config` con `{ duracion_minutos, num_preguntas }` (los dos obligatorios) → `200` con la misma forma. La duración va de 1 a 480 minutos. Un `inicia_en` que llegue se ignora. `422` si `num_preguntas` supera las preguntas del banco (`fields.num_preguntas` dice cuántas hay).

### Evaluación del postulante (solo `postulante`)
Las rutas `/evaluacion/*` exigen token de postulante (`401` sin token, `403` con token de analista o administrador). Todas las horas van en UTC y el tiempo restante lo calcula el servidor.

- `POST /evaluacion/iniciar` (sin cuerpo) → `{ total_preguntas, inicio, cierre, tiempo_restante_segundos }`. `201` si crea el examen y `200` si el postulante ya tenía uno en curso (devuelve el mismo, con el tiempo que le queda).
  - `403` si no tiene la postulación enviada o si no superó la Fase de Méritos (`FORBIDDEN`).
  - **El horario del examen lo fija la fase Evaluación** (`GET` y `PUT /admin/fases`, el mismo panel de las demás fases): solo se puede iniciar con la fase **publicada** y con la hora entre su `fecha_inicio` y su `fecha_fin` (sin fecha en un lado, no hay límite por ese lado). Con la fase sin publicar es `403` "La evaluación todavía no está habilitada"; antes de la fecha de inicio es `403` "La evaluación aún no ha comenzado" y `fields.inicia_en` dice cuándo empieza; desde la fecha de fin es `403` "El plazo para iniciar la evaluación terminó" y `fields.cierra_en` dice cuándo cerró.
  - Quien ya empezó conserva su examen hasta su hora de cierre, aunque después se cierre o se despublique la fase.
  - Hay un **solo intento por postulante**: uno ya terminado o vencido no se reabre (`409`). `422` si el banco no tiene las preguntas que pide la configuración.
  - El sorteo toma `num_preguntas` preguntas distintas del banco, en orden aleatorio, y cada pregunta con sus opciones en orden aleatorio (`crypto/rand`), distinto para cada postulante. `duracion_minutos` y `num_preguntas` salen de la configuración de la evaluación (30 y 10 mientras no se configure).
  - **La hora de cierre del examen de cada postulante es la menor entre su inicio más `duracion_minutos` y la `fecha_fin` de la fase.** Ejemplo: la fase va de 10:00 a 12:00 y el examen dura 1 hora; quien inicia a las 10:00 cierra a las 11:00, quien inicia a las 11:50 cierra a las 12:00 y solo tiene 10 minutos.
- `GET /evaluacion/pregunta/{n}` (n desde 1) → `{ numero, total, id, enunciado, opciones, cierre, tiempo_restante_segundos, opcion_elegida }`. Las opciones vienen en el orden que le tocó al postulante y **nunca** incluye la respuesta correcta. `opcion_elegida` es la posición (desde 0) de la opción que ya marcó, o `null` si todavía no respondió: sirve para mostrarla marcada si vuelve a la pregunta. `404` si no ha iniciado la evaluación o `n` no existe en su examen; `409` si el examen ya terminó o se agotó el tiempo.

### Respuestas, calificación y reporte (EF-14)
- `PUT /evaluacion/respuesta/{n}` con `{ opcion }` (n desde 1; `opcion` es la **posición** de la opción tal como se le mostró, desde 0) → `204`. Se puede cambiar la respuesta las veces que quiera mientras el examen siga abierto. `400` con `fields.opcion` si falta o está fuera de las opciones de la pregunta, `404` si no inició el examen o `n` no existe, y `409` si el examen ya terminó o se pasó la hora de cierre. Se guarda como el índice de esa opción en el banco, para calificar sin depender del orden de cada examen.
- `POST /evaluacion/finalizar` (sin cuerpo) → `{ nota, aciertos, total, finalizado_at }`. Califica con las respuestas guardadas: la **nota es el porcentaje de aciertos sobre 100**, con dos decimales; una pregunta sin responder cuenta como no acertada. Es idempotente: si el examen ya estaba finalizado devuelve el mismo resultado sin volver a calificar ni a enviar el correo. `404` si no inició el examen.
- **Cierre automático:** un examen cuya hora de cierre ya pasó se finaliza solo, con las respuestas guardadas hasta ese momento y con `finalizado_at` igual a su hora de cierre. Ocurre al consultarlo, al intentar guardar una respuesta tarde, al abrir los resultados del administrador, y en una tarea del servidor que revisa **cada minuto**, así que quien cierra el navegador también recibe su nota.
- **Reporte por correo:** al finalizar, el servidor envía al correo del postulante el reporte PDF de su resultado. Si el envío falla, la nota queda guardada y el error se registra en el log; el administrador puede descargar el reporte igual.
- `GET /admin/evaluacion/resultados` → `{ items: [ { intento_id, codigo, cedula, nombres, apellidos, cargo, estado, inicio, cierre, finalizado_at, nota, aciertos, total_preguntas } ] }`, ordenado por apellidos y nombres. `estado` es `en_curso` o `finalizado`; `nota`, `aciertos` y `finalizado_at` son `null` mientras sigue en curso. Antes de listar finaliza los exámenes vencidos.
- `GET /admin/evaluacion/resultados/{id}/reporte` → el PDF (`application/pdf`, descarga con `Content-Disposition: attachment`). `404` si el resultado no existe y `409` si el examen sigue en curso. El reporte lleva los datos del postulante, la nota, el detalle por pregunta (correcta, incorrecta o sin responder, **sin mostrar la respuesta correcta**) y un código de verificación con QR. Ambas rutas son solo para el administrador (`403` para el analista).

## Utilidad
`GET /health` → `200 { "status": "ok" }` (sin prefijo `/api/v1`).
