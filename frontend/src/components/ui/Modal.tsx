import { useEffect, useId, useRef, type ReactNode } from 'react'
import Button from './Button'
import Icon from './Icon'

interface ModalProps {
  abierto: boolean
  titulo: string
  onCerrar: () => void
  children: ReactNode
  pie?: ReactNode
  // Mientras es true, Escape, el fondo y la X no cierran el diálogo (p. ej. durante un guardado).
  bloqueado?: boolean
}

const ENFOCABLES =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'

// Diálogo base: cierra con Escape o al pulsar el fondo. Al abrir mueve el foco dentro,
// lo mantiene ahí con Tab y al cerrar lo devuelve al elemento que lo abrió.
export default function Modal({ abierto, titulo, onCerrar, children, pie, bloqueado }: ModalProps) {
  const tituloId = useId()
  const dialogo = useRef<HTMLDivElement>(null)
  // Referencia a la última versión de onCerrar/bloqueado para no reiniciar el efecto del foco.
  const cerrar = useRef({ onCerrar, bloqueado })
  useEffect(() => {
    cerrar.current = { onCerrar, bloqueado }
  })

  useEffect(() => {
    if (!abierto) return
    const anterior = document.activeElement as HTMLElement | null
    const nodo = dialogo.current
    const primero = nodo?.querySelector<HTMLElement>(ENFOCABLES)
    ;(primero ?? nodo)?.focus()

    const alPulsar = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (!cerrar.current.bloqueado) cerrar.current.onCerrar()
        return
      }
      if (e.key !== 'Tab' || !nodo) return
      const enfocables = Array.from(nodo.querySelectorAll<HTMLElement>(ENFOCABLES))
      if (enfocables.length === 0) {
        e.preventDefault()
        return
      }
      const inicio = enfocables[0]
      const fin = enfocables[enfocables.length - 1]
      if (e.shiftKey && document.activeElement === inicio) {
        e.preventDefault()
        fin.focus()
      } else if (!e.shiftKey && document.activeElement === fin) {
        e.preventDefault()
        inicio.focus()
      }
    }
    document.addEventListener('keydown', alPulsar)
    return () => {
      document.removeEventListener('keydown', alPulsar)
      anterior?.focus()
    }
  }, [abierto])

  if (!abierto) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-tertiary/60 p-4"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget && !bloqueado) onCerrar()
      }}
    >
      <div
        ref={dialogo}
        role="dialog"
        aria-modal="true"
        aria-labelledby={tituloId}
        tabIndex={-1}
        className="flex max-h-[90vh] w-full max-w-[600px] flex-col overflow-hidden rounded-2xl bg-surface-container-lowest shadow-xl focus:outline-none"
      >
        <div className="flex items-center justify-between gap-4 bg-surface-container-low px-6 py-5">
          <h2 id={tituloId} className="text-headline-sm text-primary">
            {titulo}
          </h2>
          <button
            type="button"
            aria-label="Cerrar"
            onClick={onCerrar}
            disabled={bloqueado}
            className="flex h-9 w-9 items-center justify-center rounded-lg text-on-surface-variant hover:bg-surface-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
          >
            <Icon nombre="close" className="text-[20px]" />
          </button>
        </div>
        <div className="overflow-y-auto p-6">{children}</div>
        {pie && (
          <div className="flex items-center justify-end gap-3 bg-surface-container-low px-6 py-4">
            {pie}
          </div>
        )}
      </div>
    </div>
  )
}

interface ModalConfirmacionProps {
  abierto: boolean
  titulo: string
  mensaje: ReactNode
  textoConfirmar?: string
  textoCancelar?: string
  peligroso?: boolean
  cargando?: boolean
  onConfirmar: () => void
  onCancelar: () => void
}

export function ModalConfirmacion({
  abierto,
  titulo,
  mensaje,
  textoConfirmar = 'Confirmar',
  textoCancelar = 'Cancelar',
  peligroso,
  cargando,
  onConfirmar,
  onCancelar,
}: ModalConfirmacionProps) {
  return (
    <Modal
      abierto={abierto}
      titulo={titulo}
      onCerrar={onCancelar}
      bloqueado={cargando}
      pie={
        <>
          <Button variante="secondary" onClick={onCancelar} disabled={cargando}>
            {textoCancelar}
          </Button>
          <Button
            variante={peligroso ? 'danger' : 'primary'}
            onClick={onConfirmar}
            cargando={cargando}
          >
            {textoConfirmar}
          </Button>
        </>
      }
    >
      <div className="text-body-md text-on-surface-variant">{mensaje}</div>
    </Modal>
  )
}
