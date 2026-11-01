import { useEffect, useId, type ReactNode } from 'react'
import Button from './Button'
import Icon from './Icon'

interface ModalProps {
  abierto: boolean
  titulo: string
  onCerrar: () => void
  children: ReactNode
  pie?: ReactNode
}

// Diálogo base: cierra con Escape o al pulsar el fondo.
export default function Modal({ abierto, titulo, onCerrar, children, pie }: ModalProps) {
  const tituloId = useId()

  useEffect(() => {
    if (!abierto) return
    const alPulsar = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCerrar()
    }
    document.addEventListener('keydown', alPulsar)
    return () => document.removeEventListener('keydown', alPulsar)
  }, [abierto, onCerrar])

  if (!abierto) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-tertiary/60 p-4"
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) onCerrar()
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={tituloId}
        className="flex max-h-[90vh] w-full max-w-[600px] flex-col overflow-hidden rounded-2xl bg-surface-container-lowest shadow-xl"
      >
        <div className="flex items-center justify-between gap-4 bg-surface-container-low px-6 py-5">
          <h2 id={tituloId} className="text-headline-sm text-primary">
            {titulo}
          </h2>
          <button
            type="button"
            aria-label="Cerrar"
            onClick={onCerrar}
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
