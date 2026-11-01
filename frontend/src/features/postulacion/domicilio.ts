import { z } from 'zod'
import type { Domicilio as DomicilioApi } from '@/services/postulacion'

// Valores del formulario de domicilio. Los ids de los selects son texto (así los entrega el
// DOM); '' significa "sin elegir".
export interface ValoresDomicilio {
  paisId: string
  provinciaId: string
  cantonId: string
  provinciaTexto: string
  ciudadTexto: string
  telefonoCelular: string
  emailSecundario: string
  callePrincipal: string
  numero: string
  calleSecundaria: string
  sector: string
  referencia: string
  telefonoDomicilio: string
}

export const DOMICILIO_VACIO: ValoresDomicilio = {
  paisId: '',
  provinciaId: '',
  cantonId: '',
  provinciaTexto: '',
  ciudadTexto: '',
  telefonoCelular: '',
  emailSecundario: '',
  callePrincipal: '',
  numero: '',
  calleSecundaria: '',
  sector: '',
  referencia: '',
  telefonoDomicilio: '',
}

// Nombre del país cuya provincia y cantón se eligen de las listas (como en el servidor).
export const PAIS_ECUADOR = 'Ecuador'

// Longitudes máximas: las mismas que valida el servidor.
const MAX = { calle: 100, numero: 20, sector: 100, referencia: 200, geo: 100, correo: 254 }

const REQUERIDO = 'Campo obligatorio'
const maximo = (n: number) => `Máximo ${n} caracteres`
const texto = (max: number) => z.string().trim().max(max, maximo(max))

// Las reglas de provincia y cantón dependen del país: con Ecuador son ids de las listas y
// con cualquier otro país, texto libre. Por eso el esquema se arma según `esEcuador`.
export function crearEsquemaDomicilio(esEcuador: boolean) {
  return z
    .object({
      paisId: z.string().min(1, 'Seleccione el país'),
      provinciaId: z.string(),
      cantonId: z.string(),
      provinciaTexto: texto(MAX.geo),
      ciudadTexto: texto(MAX.geo),
      telefonoCelular: z
        .string()
        .trim()
        .min(1, 'Ingrese su teléfono celular')
        .regex(/^09\d{8}$/, 'El celular debe tener 10 dígitos y empezar con 09'),
      emailSecundario: z
        .string()
        .trim()
        .max(MAX.correo, maximo(MAX.correo))
        .refine((v) => v === '' || z.email().safeParse(v).success, 'Correo electrónico no válido'),
      callePrincipal: texto(MAX.calle),
      numero: texto(MAX.numero),
      calleSecundaria: texto(MAX.calle),
      sector: texto(MAX.sector),
      referencia: texto(MAX.referencia),
      telefonoDomicilio: z
        .string()
        .trim()
        .refine((v) => v === '' || /^\d{7,15}$/.test(v), 'Ingrese de 7 a 15 dígitos, sin espacios'),
    })
    .superRefine((v, ctx) => {
      if (!v.paisId) return
      const faltante = (path: keyof ValoresDomicilio) =>
        ctx.addIssue({ code: 'custom', path: [path], message: REQUERIDO })
      if (esEcuador) {
        if (!v.provinciaId) faltante('provinciaId')
        if (!v.cantonId) faltante('cantonId')
      } else {
        if (!v.provinciaTexto) faltante('provinciaTexto')
        if (!v.ciudadTexto) faltante('ciudadTexto')
      }
    })
}

// Campos del servidor (docs/api.md, claves de `fields`) → campos del formulario.
export const CAMPOS_API_DOMICILIO = {
  'domicilio.pais_id': 'paisId',
  'domicilio.provincia_id': 'provinciaId',
  'domicilio.canton_id': 'cantonId',
  'domicilio.provincia_texto': 'provinciaTexto',
  'domicilio.ciudad_texto': 'ciudadTexto',
  'domicilio.telefono_celular': 'telefonoCelular',
  'domicilio.email_secundario': 'emailSecundario',
  'domicilio.calle_principal': 'callePrincipal',
  'domicilio.numero': 'numero',
  'domicilio.calle_secundaria': 'calleSecundaria',
  'domicilio.sector': 'sector',
  'domicilio.referencia': 'referencia',
  'domicilio.telefono_domicilio': 'telefonoDomicilio',
} as const satisfies Record<string, keyof ValoresDomicilio>

const idATexto = (id: number | null) => (id === null ? '' : String(id))
const textoAId = (v: string) => (v === '' ? null : Number(v))

export function valoresDesdeApi(d: DomicilioApi): ValoresDomicilio {
  return {
    paisId: idATexto(d.pais_id),
    provinciaId: idATexto(d.provincia_id),
    cantonId: idATexto(d.canton_id),
    provinciaTexto: d.provincia_texto,
    ciudadTexto: d.ciudad_texto,
    telefonoCelular: d.telefono_celular,
    emailSecundario: d.email_secundario,
    callePrincipal: d.calle_principal,
    numero: d.numero,
    calleSecundaria: d.calle_secundaria,
    sector: d.sector,
    referencia: d.referencia,
    telefonoDomicilio: d.telefono_domicilio,
  }
}

export function domicilioParaApi(v: ValoresDomicilio): DomicilioApi {
  return {
    pais_id: textoAId(v.paisId),
    provincia_id: textoAId(v.provinciaId),
    canton_id: textoAId(v.cantonId),
    provincia_texto: v.provinciaTexto,
    ciudad_texto: v.ciudadTexto,
    telefono_celular: v.telefonoCelular,
    email_secundario: v.emailSecundario,
    calle_principal: v.callePrincipal,
    numero: v.numero,
    calle_secundaria: v.calleSecundaria,
    sector: v.sector,
    referencia: v.referencia,
    telefono_domicilio: v.telefonoDomicilio,
  }
}
