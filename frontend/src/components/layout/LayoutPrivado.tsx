import type { ReactNode } from 'react'
import { Outlet } from 'react-router-dom'
import Badge from '@/components/ui/Badge'
import Icon from '@/components/ui/Icon'
import Marca from './Marca'
import { cerrarSesion } from '@/store/sesion'
import PieDePagina from './PieDePagina'

interface LayoutPrivadoProps {
  // El login no devuelve datos del usuario (docs/api.md), así que el nombre llega por props si se tiene.
  usuario?: { nombre: string; cedula: string }
  onCerrarSesion?: () => void
  children?: ReactNode
}

// Pantallas con sesión: encabezado con el usuario y el botón Cerrar sesión.
export default function LayoutPrivado({
  usuario,
  onCerrarSesion = cerrarSesion,
  children,
}: LayoutPrivadoProps) {
  return (
    <div className="flex min-h-screen flex-col bg-surface">
      <header className="sticky top-0 z-40 bg-surface-container-lowest shadow-[0_1px_8px_rgba(0,0,0,0.04)]">
        <div className="flex h-16 items-center justify-between gap-4 px-margin-mobile sm:px-margin">
          <Marca />
          <div className="flex shrink-0 items-center gap-3">
            {usuario && (
              <div className="hidden flex-col text-right md:flex">
                <span className="text-body-bold text-on-surface">{usuario.nombre}</span>
                <span className="text-legal-code text-on-surface-variant">
                  C.I. {usuario.cedula}
                </span>
              </div>
            )}
            <Badge>Postulante</Badge>
            <button
              type="button"
              onClick={onCerrarSesion}
              className="inline-flex items-center gap-1.5 rounded-lg px-space-md py-space-xs text-label-md text-error hover:bg-error-container hover:text-on-error-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
            >
              <Icon nombre="logout" className="text-[18px]" />
              <span className="hidden sm:inline">Cerrar sesión</span>
            </button>
          </div>
        </div>
      </header>
      <div className="flex flex-1 flex-col">{children ?? <Outlet />}</div>
      <PieDePagina />
    </div>
  )
}
