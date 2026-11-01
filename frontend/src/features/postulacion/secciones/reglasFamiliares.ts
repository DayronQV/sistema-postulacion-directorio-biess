import { z } from 'zod'
import type { Familiar, Institucion } from '@/services/postulacion'

// Reglas de la sección "Información adicional" (familiares en el IESS o BIESS). Reproducen
// las del servidor para avisar antes de enviar; el servidor sigue siendo quien decide.

export const MAX_FAMILIARES = 20
export const MAX_NOMBRES = 150
export const MAX_AREA = 100
export const MAX_CARGO = 100

export const PARENTESCOS = [
  'conviviente',
  'cónyuge',
  'cuñado/a',
  'hermano/a',
  'hijo político',
  'hijo/a',
  'madre',
  'nieto/a',
  'nuera',
  'padre',
  'primo hermano/a',
  'sobrino/a',
  'suegro/a',
  'tío/a',
  'yerno',
] as const

export const INSTITUCIONES = ['IESS', 'BIESS'] as const

export const etiquetaParentesco = (valor: string) =>
  valor ? valor.charAt(0).toUpperCase() + valor.slice(1) : ''

export const opcionesParentesco = PARENTESCOS.map((p) => ({
  valor: p,
  etiqueta: etiquetaParentesco(p),
}))

export const opcionesInstitucion = INSTITUCIONES.map((i) => ({ valor: i, etiqueta: i }))

const parentescos: readonly string[] = PARENTESCOS
const instituciones: readonly string[] = INSTITUCIONES

export const esquemaFamiliar = z.object({
  nombres_apellidos: z
    .string()
    .trim()
    .min(1, 'Ingrese los nombres y apellidos')
    .max(MAX_NOMBRES, `Máximo ${MAX_NOMBRES} caracteres`),
  parentesco: z.string().refine((v) => parentescos.includes(v), 'Seleccione el parentesco'),
  institucion: z
    .string()
    .refine((v): v is Institucion => instituciones.includes(v), 'Seleccione la institución'),
  area: z.string().trim().min(1, 'Ingrese el área').max(MAX_AREA, `Máximo ${MAX_AREA} caracteres`),
  cargo: z
    .string()
    .trim()
    .min(1, 'Ingrese el cargo')
    .max(MAX_CARGO, `Máximo ${MAX_CARGO} caracteres`),
})

// Lo que escribe el usuario en el formulario de una fila (la institución puede estar vacía).
export type FormFamiliar = z.input<typeof esquemaFamiliar>

export const familiarVacio: FormFamiliar = {
  nombres_apellidos: '',
  parentesco: '',
  institucion: '',
  area: '',
  cargo: '',
}

export const etiquetasCampo: Record<keyof Familiar, string> = {
  nombres_apellidos: 'Nombres y apellidos',
  parentesco: 'Parentesco',
  institucion: 'Institución',
  area: 'Área',
  cargo: 'Cargo',
}

export interface ErroresServidor {
  // Error de la sección completa (campo `familiares`).
  general?: string
  // Mensajes por fila, con el índice de la fila (desde 0) como clave.
  filas: Record<number, string[]>
  // Errores de otros campos del formulario que el servidor rechazó en el mismo guardado.
  otros: string[]
}

const CAMPO_FILA = /^familiares\[(\d+)\]\.(\w+)$/

// Reparte los `fields` de un error de la API (docs/api.md): `familiares`,
// `familiares[2].area`, ... en errores de la sección, de cada fila y de otros campos.
export function repartirErrores(fields: Record<string, string> | undefined): ErroresServidor {
  const out: ErroresServidor = { filas: {}, otros: [] }
  for (const [campo, mensaje] of Object.entries(fields ?? {})) {
    const fila = CAMPO_FILA.exec(campo)
    if (campo === 'familiares') {
      out.general = mensaje
    } else if (fila) {
      const indice = Number(fila[1])
      const nombre = etiquetasCampo[fila[2] as keyof Familiar] ?? fila[2]
      ;(out.filas[indice] ??= []).push(`${nombre}: ${mensaje}`)
    } else {
      out.otros.push(`${campo}: ${mensaje}`)
    }
  }
  return out
}
