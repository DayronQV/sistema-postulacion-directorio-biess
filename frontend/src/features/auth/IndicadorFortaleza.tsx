import { cx } from '@/lib/cx'
import { fortalezaPassword } from './fortaleza'

const colores = ['', 'bg-error', 'bg-secondary', 'bg-success'] as const

interface IndicadorFortalezaProps {
  password: string
}

// Barra de tres tramos con la fortaleza estimada de la contraseña. Con la contraseña vacía
// queda invisible pero ocupa su lugar: si apareciera recién con la primera letra, el
// formulario crecería de golpe y la página saltaría.
export default function IndicadorFortaleza({ password }: IndicadorFortalezaProps) {
  const { nivel, etiqueta } = fortalezaPassword(password)
  return (
    <div
      className={cx('flex items-center gap-space-sm', nivel === 0 && 'invisible')}
      aria-live="polite"
      data-testid="indicador-fortaleza"
    >
      <div className="grid flex-1 grid-cols-3 gap-1" aria-hidden="true">
        {[1, 2, 3].map((tramo) => (
          <span
            key={tramo}
            className={cx(
              'h-1.5 rounded-full',
              tramo <= nivel ? colores[nivel] : 'bg-surface-container-high',
            )}
          />
        ))}
      </div>
      <span className="text-label-sm text-on-surface-variant">Seguridad: {etiqueta}</span>
    </div>
  )
}
