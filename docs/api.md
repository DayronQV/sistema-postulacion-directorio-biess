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
| `POST /auth/login` | `{ cedula, password }` | `200 { token, expires_in }`; `401` si es incorrecto; `429 LOCKED` tras 3 fallos |
| `POST /auth/recuperar` | `{ cedula, fecha_expedicion? , codigo_dactilar? }` | `202`; el correo es simulado |

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

Forma de `PUT /postulacion` (núcleo):
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
  "cargo": "afiliados"
}
```
`cargo` ∈ `afiliados | jubilados | empleador`.

## Utilidad
`GET /health` → `200 { "status": "ok" }` (sin prefijo `/api/v1`).
