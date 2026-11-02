import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { sesionExpirada, useSesion, type Rol } from '@/store/sesion'
import { RUTA_FORMULARIO, RUTA_LOGIN } from './rutas'

interface RutaProtegidaProps {
  roles?: Rol[]
  redirigirA?: string
}

// Estado que recibe el login al llegar desde una ruta protegida.
export interface EstadoRedireccion {
  desde: string // ruta privada pedida, para volver a ella tras iniciar sesión
  expirada: boolean // la sesión terminó por vencimiento: el login avisa "Su sesión expiró"
}

// Sin sesión redirige al login (guardando la ruta de origen y si la sesión expiró); con
// sesión pero sin el rol pedido vuelve al inicio; en el primer ingreso solo deja abrir el
// formulario. Es solo comodidad de interfaz: el servidor vuelve a validar.
export default function RutaProtegida({ roles, redirigirA = RUTA_LOGIN }: RutaProtegidaProps) {
  const sesion = useSesion()
  const ubicacion = useLocation()

  if (!sesion) {
    const estado: EstadoRedireccion = {
      desde: ubicacion.pathname + ubicacion.search,
      expirada: sesionExpirada(),
    }
    return <Navigate to={redirigirA} replace state={estado} />
  }
  if (roles && (!sesion.rol || !roles.includes(sesion.rol))) return <Navigate to="/" replace />
  if (sesion.primerIngreso && ubicacion.pathname !== RUTA_FORMULARIO) {
    return <Navigate to={RUTA_FORMULARIO} replace />
  }
  return <Outlet />
}
