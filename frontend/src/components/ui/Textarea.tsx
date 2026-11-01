import type { ComponentProps } from 'react'
import { cx } from '@/lib/cx'
import { bordeControl } from './estilosControl'

interface TextareaProps extends ComponentProps<'textarea'> {
  invalido?: boolean
}

export default function Textarea({ invalido, className, rows = 4, ...resto }: TextareaProps) {
  return (
    <textarea
      rows={rows}
      aria-invalid={invalido || undefined}
      className={cx(
        'w-full rounded-xl border bg-surface-container-lowest px-4 py-3 text-body-md text-on-surface shadow-sm',
        'placeholder:text-outline focus:outline-none focus:ring-2 focus:ring-primary-container',
        'disabled:cursor-not-allowed disabled:bg-surface-container-low',
        bordeControl(invalido),
        className,
      )}
      {...resto}
    />
  )
}
