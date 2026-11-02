import { NavLink, Outlet } from 'react-router-dom'
import Badge from '@/components/ui/Badge'
import Icon from '@/components/ui/Icon'
import Marca from '@/components/layout/Marca'
import PieDePagina from '@/components/layout/PieDePagina'
import { cx } from '@/lib/cx'
import { cerrarSesion, useSesion } from '@/store/sesion'
import { menuParaRol } from './menu'

const nombreRol = { analista: 'Analista', administrador: 'Administrador', postulante: 'Postulante' }

// Pantallas del panel de administración: encabezado con el rol, menú del panel y
// "Cerrar sesión" (al cerrar, la ruta protegida lleva al login del panel).
export default function LayoutAdmin() {
  const rol = useSesion()?.rol
  const menu = menuParaRol(rol)

  return (
    <div className="flex min-h-screen flex-col bg-surface">
      <header className="sticky top-0 z-40 bg-surface-container-lowest shadow-[0_1px_8px_rgba(0,0,0,0.04)]">
        <div className="flex h-16 items-center justify-between gap-4 px-margin-mobile sm:px-margin">
          <Marca />
          <div className="flex shrink-0 items-center gap-3">
            <span className="hidden text-label-md text-on-surface-variant md:inline">
              Panel de administración
            </span>
            {rol && <Badge>{nombreRol[rol]}</Badge>}
            <button
              type="button"
              onClick={() => cerrarSesion()}
              className="inline-flex items-center gap-1.5 rounded-lg px-space-md py-space-xs text-label-md text-error hover:bg-error-container hover:text-on-error-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
            >
              <Icon nombre="logout" className="text-[18px]" />
              <span className="hidden sm:inline">Cerrar sesión</span>
            </button>
          </div>
        </div>
        <nav
          aria-label="Menú del panel"
          className="flex gap-1 overflow-x-auto border-t border-outline-variant px-margin-mobile sm:px-margin"
        >
          {menu.map((item) => (
            <NavLink
              key={item.ruta}
              to={item.ruta}
              end
              className={({ isActive }) =>
                cx(
                  'flex shrink-0 items-center gap-1.5 border-b-2 px-3 py-2.5 text-label-md focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
                  isActive
                    ? 'border-primary-container text-primary-container'
                    : 'border-transparent text-on-surface-variant hover:text-on-surface',
                )
              }
            >
              <Icon nombre={item.icono} className="text-[18px]" />
              {item.titulo}
            </NavLink>
          ))}
        </nav>
      </header>
      <div className="flex flex-1 flex-col">
        <Outlet />
      </div>
      <PieDePagina />
    </div>
  )
}
