import { useId, useRef, useState } from 'react'
import { cx } from '@/lib/cx'
import Icon from './Icon'
import Spinner from './Spinner'

export type EstadoSubida = 'inactivo' | 'subiendo' | 'exito' | 'error'

interface ArchivoMostrado {
  nombre: string
  tamano: number
}

interface FileUploadProps {
  etiqueta?: string
  archivo: ArchivoMostrado | null
  estado?: EstadoSubida
  mensajeError?: string
  maxMb: number
  onSeleccionar: (archivo: File) => void
  onQuitar?: () => void
}

function formatearTamano(bytes: number): string {
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

// Selector visual de un PDF. No sube nada: la subida la hace quien lo usa (issue de documentos).
// Las comprobaciones de tipo y tamaño son una ayuda; el servidor valida de nuevo.
export default function FileUpload({
  etiqueta = 'Documento de respaldo (PDF)',
  archivo,
  estado = 'inactivo',
  mensajeError,
  maxMb,
  onSeleccionar,
  onQuitar,
}: FileUploadProps) {
  const id = useId()
  const entrada = useRef<HTMLInputElement>(null)
  const [errorLocal, setErrorLocal] = useState<string>()
  const error = errorLocal ?? (estado === 'error' ? mensajeError : undefined)

  function alElegir(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0]
    e.target.value = ''
    if (!f) return
    if (f.type !== 'application/pdf') {
      setErrorLocal('Solo se permiten archivos PDF.')
      return
    }
    if (f.size > maxMb * 1024 * 1024) {
      setErrorLocal(`El archivo supera el máximo de ${maxMb} MB.`)
      return
    }
    setErrorLocal(undefined)
    onSeleccionar(f)
  }

  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-label-md text-on-surface">{etiqueta}</span>
      <div
        className={cx(
          'flex flex-wrap items-center gap-3 rounded-xl border border-dashed bg-surface-container-lowest p-4',
          error ? 'border-error' : 'border-outline-variant',
        )}
      >
        <Icon nombre="picture_as_pdf" className="text-[28px] text-primary-container" />
        <div className="flex min-w-[12rem] flex-1 flex-col">
          {archivo ? (
            <>
              <span className="truncate text-body-bold text-on-surface">{archivo.nombre}</span>
              <span className="text-label-sm text-on-surface-variant">
                {formatearTamano(archivo.tamano)}
              </span>
            </>
          ) : (
            <span className="text-body-md text-on-surface-variant">
              Ningún archivo seleccionado (máx. {maxMb} MB)
            </span>
          )}
        </div>
        {estado === 'subiendo' && (
          <span
            role="status"
            className="flex items-center gap-2 text-label-sm text-primary-container"
          >
            <Spinner className="h-4 w-4" /> Subiendo...
          </span>
        )}
        {estado === 'exito' && (
          <span className="flex items-center gap-1 text-label-sm text-success">
            <Icon nombre="check_circle" relleno className="text-[18px]" /> Subido
          </span>
        )}
        <input
          id={id}
          ref={entrada}
          type="file"
          accept="application/pdf"
          className="sr-only"
          onChange={alElegir}
          disabled={estado === 'subiendo'}
        />
        <label
          htmlFor={id}
          className="inline-flex h-10 cursor-pointer items-center rounded-xl border border-primary-container px-4 text-body-bold text-primary-container hover:bg-surface-container-low focus-within:ring-2"
        >
          {archivo ? 'Reemplazar' : 'Seleccionar archivo'}
        </label>
        {archivo && onQuitar && estado !== 'subiendo' && (
          <button
            type="button"
            aria-label="Quitar archivo"
            onClick={onQuitar}
            className="rounded-lg p-2 text-error hover:bg-error-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
          >
            <Icon nombre="delete" className="text-[18px]" />
          </button>
        )}
      </div>
      {error && (
        <p role="alert" className="text-label-sm text-error">
          {error}
        </p>
      )}
    </div>
  )
}
