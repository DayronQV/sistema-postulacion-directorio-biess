import type { ReactNode } from 'react'
import { cx } from '@/lib/cx'

interface CardProps {
  titulo?: string
  descripcion?: string
  acciones?: ReactNode
  children: ReactNode
  className?: string
}

export default function Card({ titulo, descripcion, acciones, children, className }: CardProps) {
  return (
    <section className={cx('rounded-xl bg-surface-container-lowest p-6 shadow-sm', className)}>
      {(titulo || acciones) && (
        <header className="mb-space-md flex items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            {titulo && <h2 className="text-headline-sm text-primary">{titulo}</h2>}
            {descripcion && <p className="text-body-md text-on-surface-variant">{descripcion}</p>}
          </div>
          {acciones}
        </header>
      )}
      {children}
    </section>
  )
}
