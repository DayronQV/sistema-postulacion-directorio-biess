import type { ReactNode } from 'react'
import { cx } from '@/lib/cx'

type Tono = 'neutral' | 'primary' | 'success' | 'warning' | 'error'

interface BadgeProps {
  tono?: Tono
  children: ReactNode
}

const estilos: Record<Tono, string> = {
  neutral: 'bg-surface-container-highest text-on-surface',
  primary: 'bg-primary-container text-on-primary',
  success: 'bg-success-container text-on-success-container',
  warning: 'bg-secondary-container text-on-secondary-container',
  error: 'bg-error-container text-on-error-container',
}

export default function Badge({ tono = 'neutral', children }: BadgeProps) {
  return (
    <span
      className={cx(
        'inline-flex items-center rounded-lg px-space-sm py-0.5 text-label-sm',
        estilos[tono],
      )}
    >
      {children}
    </span>
  )
}
