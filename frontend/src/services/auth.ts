import { http } from './http'

// Llamadas de autenticación (docs/api.md, sección Autenticación). Reciben la señal de
// cancelación como primer argumento para usarse con useApi.

export interface DatosIdentidad {
  cedula: string
  fecha_expedicion?: string // AAAA-MM-DD
  codigo_dactilar?: string
}

export interface DatosCuenta {
  ticket: string
  email: string
  password: string
}

// POST /auth/identidad: primer paso del registro. Devuelve el ticket para el segundo paso.
export async function validarIdentidad(signal: AbortSignal, datos: DatosIdentidad) {
  const { data } = await http.post<{ ticket: string }>('/auth/identidad', datos, { signal })
  return data.ticket
}

// POST /auth/registro: segundo paso. Devuelve el id de la cuenta creada.
export async function registrarCuenta(signal: AbortSignal, datos: DatosCuenta) {
  const { data } = await http.post<{ id: string }>('/auth/registro', datos, { signal })
  return data.id
}

export interface Credenciales {
  cedula: string
  password: string
}

export interface RespuestaLogin {
  token: string
  expires_in: number // segundos
  primer_ingreso: boolean // aún no tiene postulación: va obligatoriamente al formulario
}

// POST /auth/login: inicio de sesión del postulante.
export async function iniciarSesionPostulante(signal: AbortSignal, credenciales: Credenciales) {
  const { data } = await http.post<RespuestaLogin>('/auth/login', credenciales, { signal })
  return data
}
