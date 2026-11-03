import type { ReactNode } from 'react'
import { Outlet } from 'react-router-dom'
import Marca from './Marca'
import PieDePagina from './PieDePagina'

interface LayoutPublicoProps {
  children?: ReactNode
}

// Pantallas sin sesión: inicio, registro, login y recuperación.
export default function LayoutPublico({ children }: LayoutPublicoProps) {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="flex justify-center px-margin-mobile py-space-lg sm:px-margin">
        <Marca />
      </header>
      <div className="flex flex-1 flex-col">{children ?? <Outlet />}</div>
      <PieDePagina />
    </div>
  )
}
