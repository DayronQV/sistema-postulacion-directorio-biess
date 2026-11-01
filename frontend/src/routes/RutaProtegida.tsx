import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useSesion, type Rol } from '@/store/sesion'

interface RutaProtegidaProps {
  roles?: Rol[]
  redirigirA?: string
}

// Sin sesión redirige al login (guardando la ruta de origen); con sesión pero sin el rol
// pedido vuelve al inicio. Es solo comodidad de interfaz: el servidor vuelve a validar.
export default function RutaProtegida({ roles, redirigirA = '/login' }: RutaProtegidaProps) {
  const sesion = useSesion()
  const ubicacion = useLocation()

  if (!sesion) return <Navigate to={redirigirA} replace state={{ desde: ubicacion.pathname }} />
  if (roles && (!sesion.rol || !roles.includes(sesion.rol))) return <Navigate to="/" replace />
  return <Outlet />
}
