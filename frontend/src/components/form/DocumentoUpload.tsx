import { useEffect, useRef, useState } from 'react'
import FileUpload, { type EstadoSubida } from '@/components/ui/FileUpload'
import { ApiError } from '@/services/errores'
import {
  descargarDocumento,
  eliminarDocumento,
  subirDocumento,
  type DocumentoSubido,
  type TipoDocumento,
} from '@/services/documentos'

interface DocumentoUploadProps {
  tipo: TipoDocumento
  maxMb: 2 | 4
  valor: DocumentoSubido | null
  onChange: (documento: DocumentoSubido | null) => void
  etiqueta?: string
}

// Une el selector visual con la API. Las secciones solo conservan documento_id y
// reciben estados de carga, éxito y error sin repetir la lógica de subida.
export default function DocumentoUpload({
  tipo,
  maxMb,
  valor,
  onChange,
  etiqueta,
}: DocumentoUploadProps) {
  const control = useRef<AbortController | null>(null)
  const [estado, setEstado] = useState<EstadoSubida>('inactivo')
  const [progreso, setProgreso] = useState(0)
  const [error, setError] = useState<string>()

  useEffect(() => () => control.current?.abort(), [])

  async function seleccionar(archivo: File) {
    control.current?.abort()
    const actual = new AbortController()
    control.current = actual
    setEstado('subiendo')
    setProgreso(0)
    setError(undefined)
    try {
      const documento = await subirDocumento(actual.signal, tipo, archivo, setProgreso)
      if (!actual.signal.aborted) {
        onChange(documento)
        setEstado('exito')
      }
    } catch (fallo) {
      if (!actual.signal.aborted) {
        setError(
          fallo instanceof ApiError
            ? fallo.message
            : 'No se pudo subir el documento. Inténtelo nuevamente.',
        )
        setEstado('error')
      }
    }
  }

  async function descargar() {
    if (!valor) return
    const actual = new AbortController()
    control.current = actual
    try {
      const contenido = await descargarDocumento(actual.signal, valor.documento_id)
      const url = URL.createObjectURL(contenido)
      const enlace = document.createElement('a')
      enlace.href = url
      enlace.download = valor.nombre
      enlace.click()
      URL.revokeObjectURL(url)
    } catch (fallo) {
      if (!actual.signal.aborted) {
        setError(fallo instanceof ApiError ? fallo.message : 'No se pudo descargar el documento.')
        setEstado('error')
      }
    }
  }

  async function quitar() {
    if (!valor) return
    const actual = new AbortController()
    control.current = actual
    setError(undefined)
    try {
      await eliminarDocumento(actual.signal, valor.documento_id)
      if (!actual.signal.aborted) {
        onChange(null)
        setEstado('inactivo')
      }
    } catch (fallo) {
      if (!actual.signal.aborted) {
        setError(fallo instanceof ApiError ? fallo.message : 'No se pudo eliminar el documento.')
        setEstado('error')
      }
    }
  }

  return (
    <FileUpload
      etiqueta={etiqueta}
      archivo={valor ? { nombre: valor.nombre, tamano: valor.tamano_bytes } : null}
      estado={estado}
      progreso={progreso}
      mensajeError={error}
      maxMb={maxMb}
      onSeleccionar={seleccionar}
      onDescargar={valor ? descargar : undefined}
      onQuitar={valor ? quitar : undefined}
    />
  )
}
