import type { ComponentProps } from 'react'
import { cx } from '@/lib/cx'
import { bordeControl, estiloControl } from './estilosControl'

export interface InputProps extends ComponentProps<'input'> {
  invalido?: boolean
}

export default function Input({ invalido, className, ...resto }: InputProps) {
  return (
    <input
      aria-invalid={invalido || undefined}
      className={cx(
        estiloControl,
        bordeControl(invalido),
        'read-only:bg-surface-container-low',
        className,
      )}
      {...resto}
    />
  )
}
