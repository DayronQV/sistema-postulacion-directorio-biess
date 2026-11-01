import type { ComponentProps } from 'react'
import { cx } from '@/lib/cx'
import Icon from './Icon'
import { bordeControl, estiloControl } from './estilosControl'

export interface OpcionSelect {
  valor: string | number
  etiqueta: string
}

interface SelectProps extends ComponentProps<'select'> {
  opciones: OpcionSelect[]
  placeholder?: string
  invalido?: boolean
}

export default function Select({
  opciones,
  placeholder,
  invalido,
  className,
  ...resto
}: SelectProps) {
  return (
    <div className="relative">
      <select
        aria-invalid={invalido || undefined}
        className={cx(estiloControl, bordeControl(invalido), 'appearance-none pr-10', className)}
        {...resto}
      >
        {placeholder !== undefined && <option value="">{placeholder}</option>}
        {opciones.map((o) => (
          <option key={o.valor} value={o.valor}>
            {o.etiqueta}
          </option>
        ))}
      </select>
      <Icon
        nombre="expand_more"
        className="pointer-events-none absolute right-3 top-3.5 text-[20px] text-on-surface-variant"
      />
    </div>
  )
}
