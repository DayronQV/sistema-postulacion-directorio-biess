import type { ReactNode } from 'react'
import { cx } from '@/lib/cx'
import Icon from './Icon'

export type TipoAlert = 'success' | 'error' | 'warning' | 'info'

interface AlertProps {
  tipo?: TipoAlert
  titulo?: string
  children?: ReactNode
  onCerrar?: () => void
}

const estilos: Record<TipoAlert, { caja: string; icono: string; nombre: string }> = {
  success: {
    caja: 'bg-success-container text-on-success-container',
    icono: 'text-success',
    nombre: 'check_circle',
  },
  error: {
    caja: 'bg-error-container text-on-error-container',
    icono: 'text-error',
    nombre: 'error',
  },
  warning: {
    caja: 'bg-secondary-container text-on-secondary-container',
    icono: 'text-secondary',
    nombre: 'warning',
  },
  info: {
    caja: 'bg-primary-fixed text-on-primary-fixed-variant',
    icono: 'text-primary-container',
    nombre: 'info',
  },
}

export default function Alert({ tipo = 'info', titulo, children, onCerrar }: AlertProps) {
  const e = estilos[tipo]
  return (
    <div
      role={tipo === 'error' || tipo === 'warning' ? 'alert' : 'status'}
      className={cx('flex items-start gap-3 rounded-xl p-4 shadow-sm', e.caja)}
    >
      <Icon nombre={e.nombre} relleno className={cx('mt-0.5 shrink-0 text-[22px]', e.icono)} />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        {titulo && <p className="text-body-bold leading-tight">{titulo}</p>}
        {children && <div className="text-label-sm leading-normal">{children}</div>}
      </div>
      {onCerrar && (
        <button
          type="button"
          aria-label="Cerrar aviso"
          onClick={onCerrar}
          className="shrink-0 rounded-lg p-1 hover:bg-black/5 focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
        >
          <Icon nombre="close" className="text-[18px]" />
        </button>
      )}
    </div>
  )
}
