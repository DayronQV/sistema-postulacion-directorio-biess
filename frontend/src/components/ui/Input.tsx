import type { ComponentProps, ReactNode } from 'react'
import { cx } from '@/lib/cx'
import Icon from './Icon'
import { bordeControl, estiloControl } from './estilosControl'

export interface InputProps extends ComponentProps<'input'> {
  invalido?: boolean
  // Icono de Material Symbols dentro del campo, a la izquierda.
  icono?: string
  // Control dentro del campo, a la derecha (por ejemplo, mostrar u ocultar la contraseña).
  accion?: ReactNode
}

export default function Input({ invalido, icono, accion, className, ...resto }: InputProps) {
  const control = (
    <input
      aria-invalid={invalido || undefined}
      className={cx(
        estiloControl,
        bordeControl(invalido),
        'read-only:bg-surface-container-low',
        icono && 'pl-12',
        Boolean(accion) && 'pr-12',
        className,
      )}
      {...resto}
    />
  )
  if (!icono && !accion) return control
  return (
    <div className="relative">
      {icono && (
        <Icon
          nombre={icono}
          relleno
          className="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-[22px] text-outline"
        />
      )}
      {control}
      {accion && <div className="absolute right-2 top-1/2 -translate-y-1/2">{accion}</div>}
    </div>
  )
}
