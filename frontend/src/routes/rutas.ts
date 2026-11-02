import type { Rol } from '@/store/sesion'

export const RUTA_LOGIN = '/login'
// Única pantalla privada del postulante por ahora; el primer ingreso va aquí siempre.
export const RUTA_FORMULARIO = '/postulacion'

// Página principal de cada rol, a la que se manda a quien ya tiene sesión y abre el login
// o el registro. Las pantallas de administración llegan con EF-13; hasta entonces analistas
// y administradores van al inicio.
export function paginaPrincipal(rol?: Rol): string {
  return rol === 'analista' || rol === 'administrador' ? '/' : RUTA_FORMULARIO
}
