import { z } from 'zod'
import { cedulaValida } from './cedula'

// Esquemas de los formularios de registro e inicio de sesión. Reproducen las reglas
// del servidor para avisar antes de enviar; el servidor sigue siendo quien decide.

export const MIN_PASSWORD = 8
export const MAX_PASSWORD_BYTES = 72 // límite de bcrypt en el servidor

export const esquemaCedula = z
  .string()
  .trim()
  .min(1, 'Ingrese su número de cédula')
  .refine(cedulaValida, 'Cédula no válida')

export function fechaLocalISO(fecha = new Date()): string {
  const anio = fecha.getFullYear()
  const mes = String(fecha.getMonth() + 1).padStart(2, '0')
  const dia = String(fecha.getDate()).padStart(2, '0')
  return `${anio}-${mes}-${dia}`
}

export function mensajeFechaExpedicion(
  valor: string,
  hoy = fechaLocalISO(),
): string | undefined {
  if (!valor) return 'Ingrese la fecha de expedición'

  const partes = /^(\d{4})-(\d{2})-(\d{2})$/.exec(valor)
  if (!partes) return 'Fecha no válida'

  const [, anioTexto, mesTexto, diaTexto] = partes
  const anio = Number(anioTexto)
  const mes = Number(mesTexto)
  const dia = Number(diaTexto)
  const fecha = new Date(anio, mes - 1, dia)
  if (
    fecha.getFullYear() !== anio ||
    fecha.getMonth() !== mes - 1 ||
    fecha.getDate() !== dia
  ) {
    return 'Fecha no válida'
  }
  if (valor > hoy) return 'La fecha no puede ser futura'
  return undefined
}

export const esquemaIdentidad = z
  .object({
    cedula: esquemaCedula,
    dato: z.enum(['fecha', 'codigo']),
    fechaExpedicion: z.string(),
    codigoDactilar: z.string(),
  })
  .superRefine((v, ctx) => {
    if (v.dato === 'fecha') {
      const mensaje = mensajeFechaExpedicion(v.fechaExpedicion)
      if (mensaje) {
        ctx.addIssue({
          code: 'custom',
          path: ['fechaExpedicion'],
          message: mensaje,
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

const esquemaPassword = z.string().superRefine((p, ctx) => {
  const msg = mensajePassword(p)
  if (msg) ctx.addIssue({ code: 'custom', message: msg })
})

const esquemaConfirmacion = z.string().min(1, 'Confirme su contraseña')

const coincidenPasswords = {
  check: (v: { password: string; confirmacion: string }) => v.password === v.confirmacion,
  params: { path: ['confirmacion'], message: 'Las contraseñas no coinciden' },
}

export const esquemaCuenta = z
  .object({
    email: z
      .string()
      .trim()
      .min(1, 'Ingrese su correo electrónico')
      .pipe(z.email('Correo electrónico no válido')),
    password: esquemaPassword,
    confirmacion: esquemaConfirmacion,
  })
  .refine(coincidenPasswords.check, coincidenPasswords.params)

export type FormCuenta = z.infer<typeof esquemaCuenta>

// Nueva contraseña desde el enlace del correo: misma regla que en el registro.
export const esquemaRestablecer = z
  .object({ password: esquemaPassword, confirmacion: esquemaConfirmacion })
  .refine(coincidenPasswords.check, coincidenPasswords.params)

export type FormRestablecer = z.infer<typeof esquemaRestablecer>

// En el login la contraseña solo se exige: la regla de formato es del registro.
export const esquemaLogin = z.object({
  cedula: esquemaCedula,
  password: z.string().min(1, 'Ingrese su contraseña'),
})

export type FormLogin = z.infer<typeof esquemaLogin>
