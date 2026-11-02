import type { Rol } from '@/store/sesion'
import { http } from './http'

// Inicio de sesión del panel de administración (docs/api.md, Administración).

export interface CredencialesAdmin {
  usuario: string
  password: string
}

export interface RespuestaLoginAdmin {
  token: string
  expires_in: number // segundos
  usuario: string
  rol: Extract<Rol, 'analista' | 'administrador'>
}

// POST /admin/auth/login: mismo bloqueo por intentos que el login del postulante.
export async function iniciarSesionAdmin(signal: AbortSignal, credenciales: CredencialesAdmin) {
  const { data } = await http.post<RespuestaLoginAdmin>('/admin/auth/login', credenciales, {
    signal,
  })
  return data
}
