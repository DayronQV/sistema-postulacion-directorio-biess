import { useEffect, useRef, type ReactNode } from 'react'
import { cx } from '@/lib/cx'

interface PanelEditorProps {
  // Identifica la fila que se edita: cuando cambia (otra fila, o "Agregar" después de
  // "Editar"), el panel vuelve a llevar la pantalla hasta él.
  clave: unknown
  className?: string
  children: ReactNode
}

// Panel del formulario de una fila (agregar o editar) que aparece debajo de una tabla. Al
// abrirse baja la pantalla hasta él y pone el cursor en su primer campo: en el celular
// queda fuera de la vista y, sin esto, parece que el botón "Agregar" no hizo nada.
export default function PanelEditor({ clave, className, children }: PanelEditorProps) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const panel = ref.current
    if (!panel) return
    const sinMovimiento = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches
    panel.scrollIntoView?.({ behavior: sinMovimiento ? 'auto' : 'smooth', block: 'start' })
    panel
      .querySelector<HTMLElement>('input, select, textarea, button')
      ?.focus({ preventScroll: true })
  }, [clave])

  // scroll-mt deja libre la barra superior fija, que si no taparía el primer campo.
  return (
    <div ref={ref} className={cx('scroll-mt-24', className)}>
      {children}
    </div>
  )
}
