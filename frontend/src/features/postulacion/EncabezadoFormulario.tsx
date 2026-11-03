import type { ReactNode } from 'react'
import EncabezadoPagina from '@/components/layout/EncabezadoPagina'

interface EncabezadoFormularioProps {
  codigo?: string
  // Etiqueta del estado (Borrador, Enviada, Evaluación Técnica).
  estado: ReactNode
}

// Encabezado del formulario de postulación, el mismo en borrador y enviada.
export default function EncabezadoFormulario({ codigo, estado }: EncabezadoFormularioProps) {
  return (
    <EncabezadoPagina
      icono="description"
      titulo="Formulario de postulación"
      descripcion="Completa la información requerida para tu postulación"
      detalle={
        <>
          {codigo && (
            <p className="text-label-md text-on-surface-variant">
              Código único:{' '}
              <span className="font-mono font-semibold text-on-surface">{codigo}</span>
            </p>
          )}
          {estado}
        </>
      }
    />
  )
}
