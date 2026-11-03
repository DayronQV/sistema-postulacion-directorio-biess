import { cx } from '@/lib/cx'

export type DatoIdentidad = 'fecha' | 'codigo'

interface SelectorDatoProps {
  valor: DatoIdentidad
  onChange: (valor: DatoIdentidad) => void
  deshabilitado?: boolean
}

const opciones: Array<{ valor: DatoIdentidad; etiqueta: string }> = [
  { valor: 'fecha', etiqueta: 'Fecha de expedición' },
  { valor: 'codigo', etiqueta: 'Código dactilar' },
]

// Elige con qué dato del Registro Civil se comprueba la identidad.
export default function SelectorDato({ valor, onChange, deshabilitado }: SelectorDatoProps) {
  return (
    <div
      role="radiogroup"
      aria-label="Dato para comprobar su identidad"
      className="grid grid-cols-2 gap-1 rounded-xl bg-surface-container-high p-1"
    >
      {opciones.map((o) => {
        const activo = valor === o.valor
        return (
          <button
            key={o.valor}
            type="button"
            role="radio"
            aria-checked={activo}
            disabled={deshabilitado}
            onClick={() => onChange(o.valor)}
            className={cx(
              'rounded-lg px-3 py-2 text-label-md transition-colors',
              'focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
              'disabled:cursor-not-allowed disabled:opacity-60',
              activo
                ? 'bg-primary-container text-on-primary shadow-sm'
                : 'text-on-surface-variant hover:text-on-surface',
            )}
          >
            {o.etiqueta}
          </button>
        )
      })}
    </div>
  )
}
