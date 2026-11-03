import { Navigate, Outlet } from 'react-router-dom'
import { useSesion } from '@/store/sesion'
import { paginaPrincipal } from './rutas'

// Rutas solo para quien no tiene sesión (login y registro): con sesión abierta se manda a
// su página principal en lugar de mostrar el formulario de ingreso otra vez.
export default function RutaInvitado() {
  const sesion = useSesion()
  if (sesion) return <Navigate to={paginaPrincipal(sesion.rol)} replace />
  return <Outlet />
}
