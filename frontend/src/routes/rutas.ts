import type { Rol } from '@/store/sesion'

export const RUTA_LOGIN = '/login'
// Única pantalla privada del postulante por ahora; el primer ingreso va aquí siempre.
export const RUTA_FORMULARIO = '/postulacion'
// Evaluación teórica del postulante habilitado (EF-14).
export const RUTA_EVALUACION = '/evaluacion'
// Panel de administración (EF-13) y su login.
export const RUTA_ADMIN = '/admin'
export const RUTA_ADMIN_DASHBOARD = '/admin/dashboard'
export const RUTA_LOGIN_ADMIN = '/admin/login'

// Página principal de cada rol, a la que se manda a quien ya tiene sesión y abre un login
// o el registro: el panel para analistas y administradores, el formulario para postulantes.
export function paginaPrincipal(rol?: Rol): string {
  return rol === 'analista' || rol === 'administrador' ? RUTA_ADMIN : RUTA_FORMULARIO
}
