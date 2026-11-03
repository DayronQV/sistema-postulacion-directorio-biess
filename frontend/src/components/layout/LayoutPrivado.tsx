import type { ReactNode } from 'react'
import { Outlet } from 'react-router-dom'
import { RUTA_FORMULARIO } from '@/routes/rutas'
import { cerrarSesion } from '@/store/sesion'
import MarcoInstitucional, { type ItemMenu } from './MarcoInstitucional'

interface LayoutPrivadoProps {
  // El login no devuelve datos del usuario (docs/api.md), así que el nombre llega por props si se tiene.
  usuario?: { nombre: string; cedula: string }
  onCerrarSesion?: () => void
  children?: ReactNode
}

// Por ahora el postulante tiene una sola pantalla; cada módulo nuevo agrega aquí su enlace.
const MENU: ItemMenu[] = [{ ruta: RUTA_FORMULARIO, titulo: 'Inicio', icono: 'home' }]

// Pantallas del postulante con sesión, dentro del marco institucional.
export default function LayoutPrivado({
  usuario,
  onCerrarSesion = cerrarSesion,
  children,
}: LayoutPrivadoProps) {
  return (
    <MarcoInstitucional
      menu={MENU}
      inicio={RUTA_FORMULARIO}
      etiquetaMenu="Menú principal"
      nombreUsuario={usuario?.nombre ?? 'Postulante'}
      detalleUsuario={usuario && `C.I. ${usuario.cedula}`}
      onCerrarSesion={onCerrarSesion}
    >
      {children ?? <Outlet />}
    </MarcoInstitucional>
  )
}
