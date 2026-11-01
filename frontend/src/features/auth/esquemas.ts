import { z } from 'zod'
import { cedulaValida } from './cedula'

// Esquemas de los formularios de registro. Reproducen las reglas del servidor para
// avisar antes de enviar; el servidor sigue siendo quien decide.

export const MIN_PASSWORD = 8
export const MAX_PASSWORD_BYTES = 72 // límite de bcrypt en el servidor

export const esquemaCedula = z
  .string()
  .trim()
  .min(1, 'Ingrese su número de cédula')
  .refine(cedulaValida, 'Cédula no válida')

export const esquemaIdentidad = z
  .object({
    cedula: esquemaCedula,
    dato: z.enum(['fecha', 'codigo']),
    fechaExpedicion: z.string(),
    codigoDactilar: z.string(),
  })
  .superRefine((v, ctx) => {
    if (v.dato === 'fecha') {
      if (!v.fechaExpedicion) {
        ctx.addIssue({
          code: 'custom',
          path: ['fechaExpedicion'],
          message: 'Ingrese la fecha de expedición',
        })
      } else if (v.fechaExpedicion > new Date().toISOString().slice(0, 10)) {
        ctx.addIssue({
          code: 'custom',
          path: ['fechaExpedicion'],
          message: 'La fecha no puede ser futura',
        })
      }
    } else if (!/^[A-Za-z0-9]{10}$/.test(v.codigoDactilar.trim())) {
      ctx.addIssue({
        code: 'custom',
        path: ['codigoDactilar'],
        message: 'El código dactilar tiene 10 letras o números',
      })
    }
  })

export type FormIdentidad = z.infer<typeof esquemaIdentidad>

// El mínimo cuenta caracteres (no bytes), como el servidor: "á" es un carácter.
export function mensajePassword(p: string): string | undefined {
  if ([...p].length < MIN_PASSWORD) return 'La contraseña debe tener al menos 8 caracteres'
  if (new TextEncoder().encode(p).length > MAX_PASSWORD_BYTES) {
    return 'La contraseña es demasiado larga (máximo 72 bytes)'
  }
  if (!/\p{L}/u.test(p) || !/\p{N}/u.test(p)) return 'La contraseña debe combinar letras y números'
  return undefined
}

export const esquemaCuenta = z
  .object({
    email: z
      .string()
      .trim()
      .min(1, 'Ingrese su correo electrónico')
      .pipe(z.email('Correo electrónico no válido')),
    password: z.string().superRefine((p, ctx) => {
      const msg = mensajePassword(p)
      if (msg) ctx.addIssue({ code: 'custom', message: msg })
    }),
    confirmacion: z.string().min(1, 'Confirme su contraseña'),
  })
  .refine((v) => v.password === v.confirmacion, {
    path: ['confirmacion'],
    message: 'Las contraseñas no coinciden',
  })

export type FormCuenta = z.infer<typeof esquemaCuenta>
