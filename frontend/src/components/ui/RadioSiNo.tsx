import { cx } from '@/lib/cx'
import Icon from './Icon'

interface RadioSiNoProps {
  valor: boolean | null | undefined
  onChange: (valor: boolean) => void
  deshabilitado?: boolean
  id?: string
  'aria-labelledby'?: string
  'aria-label'?: string
  'aria-describedby'?: string
  'aria-invalid'?: boolean
}

const opciones: Array<{ etiqueta: string; valor: boolean }> = [
  { etiqueta: 'Sí', valor: true },
  { etiqueta: 'No', valor: false },
]

// Selector Sí / No en forma de píldora; el valor es un booleano (null/undefined = sin responder).
export default function RadioSiNo({ valor, onChange, deshabilitado, ...aria }: RadioSiNoProps) {
  return (
    <div
      role="radiogroup"
      className="inline-flex shrink-0 rounded-full bg-surface-container-high p-1"
      {...aria}
    >
      {opciones.map((o) => {
        const activo = valor === o.valor
        return (
          <button
            key={o.etiqueta}
            type="button"
            role="radio"
            aria-checked={activo}
            disabled={deshabilitado}
            onClick={() => onChange(o.valor)}
            className={cx(
              'flex items-center gap-2 rounded-full px-6 py-2 text-body-bold transition-colors',
              'focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
              'disabled:cursor-not-allowed disabled:opacity-60',
              activo
                ? 'bg-primary-container text-on-primary shadow-sm'
                : 'text-on-surface-variant hover:text-on-surface',
            )}
          >
            <Icon
              nombre={activo ? 'check_circle' : 'circle'}
              className={cx('text-[18px]', !activo && 'opacity-40')}
            />
            {o.etiqueta}
          </button>
        )
      })}
    </div>
  )
}
