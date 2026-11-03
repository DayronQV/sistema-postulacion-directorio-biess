import { Outlet } from 'react-router-dom'
import MarcoInstitucional from '@/components/layout/MarcoInstitucional'
import { RUTA_ADMIN } from '@/routes/rutas'
import { cerrarSesion, useSesion } from '@/store/sesion'
import { menuParaRol } from './menu'

const nombreRol = { analista: 'Analista', administrador: 'Administrador', postulante: 'Postulante' }

// Pantallas del panel de administración, dentro del marco institucional: el menú del panel
// según el rol y "Cerrar sesión" (al cerrar, la ruta protegida lleva al login del panel).
export default function LayoutAdmin() {
  const rol = useSesion()?.rol
  return (
    <MarcoInstitucional
      menu={menuParaRol(rol)}
      inicio={RUTA_ADMIN}
      etiquetaMenu="Menú del panel"
      nombreUsuario={rol ? nombreRol[rol] : 'Usuario'}
      detalleUsuario="Panel de administración"
      onCerrarSesion={cerrarSesion}
    >
      <Outlet />
    </MarcoInstitucional>
  )
}
