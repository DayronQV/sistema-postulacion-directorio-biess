import type { ComponentProps, ReactNode } from 'react'
import { cx } from '@/lib/cx'

interface CheckboxProps extends Omit<ComponentProps<'input'>, 'type'> {
  etiqueta: ReactNode
  invalido?: boolean
}

export default function Checkbox({ etiqueta, invalido, className, id, ...resto }: CheckboxProps) {
  return (
    <label
      htmlFor={id}
      className="flex cursor-pointer items-start gap-3 text-body-md text-on-surface"
    >
      <input
        id={id}
        type="checkbox"
        aria-invalid={invalido || undefined}
        className={cx(
          'mt-0.5 h-5 w-5 shrink-0 rounded-lg border-outline accent-primary-container',
          'focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
          className,
        )}
        {...resto}
      />
      <span>{etiqueta}</span>
    </label>
  )
}
