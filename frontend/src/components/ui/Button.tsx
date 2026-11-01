import type { ComponentProps } from 'react'
import { cx } from '@/lib/cx'
import Icon from './Icon'
import Spinner from './Spinner'

type Variante = 'primary' | 'secondary' | 'ghost' | 'danger'

interface ButtonProps extends ComponentProps<'button'> {
  variante?: Variante
  cargando?: boolean
  textoCargando?: string
  icono?: string
  anchoCompleto?: boolean
}

const estilos: Record<Variante, string> = {
  primary:
    'bg-primary-container text-on-primary shadow-sm hover:bg-primary active:bg-on-primary-fixed',
  secondary:
    'border border-primary-container bg-surface-container-lowest text-primary-container hover:bg-surface-container-low',
  ghost: 'bg-transparent text-primary-container hover:bg-surface-container',
  danger: 'bg-error text-on-error shadow-sm hover:opacity-90',
}

export default function Button({
  variante = 'primary',
  cargando = false,
  textoCargando,
  icono,
  anchoCompleto,
  disabled,
  className,
  children,
  type = 'button',
  ...resto
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled || cargando}
      aria-busy={cargando || undefined}
      className={cx(
        'inline-flex h-12 select-none items-center justify-center gap-2 rounded-xl px-6 text-body-bold transition-colors',
        'focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary focus-visible:ring-offset-2',
        'disabled:cursor-not-allowed disabled:bg-surface-container-high disabled:text-outline disabled:shadow-none',
        cargando && 'cursor-wait',
        estilos[variante],
        anchoCompleto && 'w-full',
        className,
      )}
      {...resto}
    >
      {cargando ? <Spinner /> : icono && <Icon nombre={icono} className="text-[20px]" />}
      <span>{cargando && textoCargando ? textoCargando : children}</span>
    </button>
  )
}
