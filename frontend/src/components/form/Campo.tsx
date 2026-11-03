import type { ReactNode } from 'react'
import Icon from '@/components/ui/Icon'

interface CampoProps {
  id: string
  etiqueta: string
  requerido?: boolean
  error?: string
  ayuda?: string
  // Icono de Material Symbols junto a la etiqueta.
  icono?: string
  children: ReactNode
}

// Envoltorio de cualquier control: etiqueta, marca de obligatorio "*", ayuda y mensaje de error.
export default function Campo({
  id,
  etiqueta,
  requerido,
  error,
  ayuda,
  icono,
  children,
}: CampoProps) {
  return (
    <div className="flex flex-col gap-1.5">
      {/* El icono va fuera del <label> para que no forme parte del nombre accesible del campo. */}
      <div className="flex items-center gap-2">
        {icono && <Icon nombre={icono} relleno className="text-[22px] text-primary-container" />}
        <label id={`${id}-etiqueta`} htmlFor={id} className="text-label-md text-on-surface">
          {etiqueta}
          {requerido && (
            <span aria-hidden="true" className="ml-0.5 font-bold text-error">
              *
            </span>
          )}
        </label>
      </div>
      {children}
      {error ? (
        <p id={`${id}-error`} role="alert" className="text-label-sm text-error">
          {error}
        </p>
      ) : (
        ayuda && (
          <p id={`${id}-ayuda`} className="text-legal-code text-on-surface-variant">
            {ayuda}
          </p>
        )
      )}
    </div>
  )
}
