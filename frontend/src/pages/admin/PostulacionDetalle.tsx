import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Icon from '@/components/ui/Icon'
import Loader from '@/components/ui/Loader'
import AccionMerito from '@/features/admin/AccionMerito'
import DocumentoPostulacion from '@/features/postulacion/imprimible/DocumentoPostulacion'
import { useUbicacion } from '@/features/postulacion/imprimible/useUbicacion'
import { descargarDocumentoAdmin, obtenerPostulacionAdmin } from '@/services/adminPostulaciones'
import { ApiError } from '@/services/errores'
import type { Postulacion } from '@/services/postulacion'

interface Certificado {
  id: string
  etiqueta: string
}

function certificadosDe(p: Postulacion): Certificado[] {
  const certificados: Certificado[] = []
  if (p.respaldo_organizacion.documento_id) {
    certificados.push({
      id: p.respaldo_organizacion.documento_id,
      etiqueta: 'Respaldo de la organización',
    })
  }
  p.formacion.forEach(
    (fila, i) =>
      fila.documento_id &&
      certificados.push({ id: fila.documento_id, etiqueta: `Formación ${i + 1}: ${fila.titulo}` }),
  )
  p.experiencia.forEach(
    (fila, i) =>
      fila.documento_id &&
      certificados.push({ id: fila.documento_id, etiqueta: `Experiencia ${i + 1}: ${fila.cargo}` }),
  )
  p.publicaciones.forEach(
    (fila, i) =>
      fila.documento_id &&
      certificados.push({
        id: fila.documento_id,
        etiqueta: `Publicación ${i + 1}: ${fila.titulo}`,
      }),
  )
  p.capacitaciones.forEach(
    (fila, i) =>
      fila.documento_id &&
      certificados.push({
        id: fila.documento_id,
        etiqueta: `Capacitación ${i + 1}: ${fila.nombre}`,
      }),
  )
  return certificados
}

export default function PostulacionDetalle() {
  const { id = '' } = useParams()
  const [postulacion, setPostulacion] = useState<Postulacion>()
  const [cargando, setCargando] = useState(true)
  const [descargando, setDescargando] = useState('')
  const [error, setError] = useState('')
  const ubicacion = useUbicacion(postulacion?.domicilio)
  const certificados = useMemo(
    () => (postulacion ? certificadosDe(postulacion) : []),
    [postulacion],
  )

  useEffect(() => {
    const control = new AbortController()
    obtenerPostulacionAdmin(control.signal, id)
      .then((data) => {
        if (!control.signal.aborted) {
          setPostulacion(data)
          setError('')
        }
      })
      .catch((fallo: unknown) => {
        if (!control.signal.aborted)
          setError(fallo instanceof ApiError ? fallo.message : 'No se pudo cargar la postulación.')
      })
      .finally(() => {
        if (!control.signal.aborted) setCargando(false)
      })
    return () => control.abort()
  }, [id])

  const descargar = async (certificado: Certificado) => {
    const control = new AbortController()
    setDescargando(certificado.id)
    setError('')
    try {
      const archivo = await descargarDocumentoAdmin(control.signal, id, certificado.id)
      const url = URL.createObjectURL(archivo.contenido)
      const enlace = document.createElement('a')
      enlace.href = url
      enlace.download = archivo.nombre
      enlace.click()
      // Revocar en el mismo instante puede cortar la descarga en algunos navegadores.
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    } catch (fallo) {
      setError(fallo instanceof ApiError ? fallo.message : 'No se pudo descargar el certificado.')
    } finally {
      setDescargando('')
    }
  }

  return (
    <main className="min-h-screen flex-1 bg-surface-container p-margin-mobile print:bg-white print:p-0 sm:p-margin">
      <div className="mx-auto mb-space-md flex max-w-[210mm] flex-wrap items-center justify-between gap-3 print:hidden">
        <Link
          to="/admin/postulaciones"
          className="inline-flex items-center gap-1.5 text-body-bold text-primary-container hover:underline"
        >
          <Icon nombre="arrow_back" className="text-[18px]" /> Volver a la nómina
        </Link>
        <Button icono="print" onClick={() => window.print()} disabled={!postulacion}>
          Imprimir / Guardar PDF
        </Button>
      </div>

      {error && (
        <div className="mx-auto mb-space-md max-w-[210mm] print:hidden">
          <Alert tipo="error" titulo="No se pudo completar la operación">
            {error}
          </Alert>
        </div>
      )}
      {cargando && <Loader texto="Cargando expediente..." />}
      {postulacion && !cargando && (
        <>
          <div className="mx-auto mb-space-md max-w-[210mm] print:hidden">
            <Card
              titulo="Fase de Méritos"
              descripcion="Después de revisar la postulación, marque si el postulante superó la Fase de Méritos: así queda habilitado para la evaluación."
            >
              <AccionMerito
                postulacionId={postulacion.id}
                codigo={postulacion.codigo}
                superoMerito={Boolean(postulacion.supero_merito)}
                enviada={postulacion.estado === 'enviada'}
              />
            </Card>
          </div>
          <div className="mx-auto mb-space-md max-w-[210mm] print:hidden">
            <Card titulo="Certificados" descripcion="Documentos PDF adjuntos a la postulación">
              {certificados.length === 0 ? (
                <p className="text-body-md text-on-surface-variant">
                  Esta postulación no tiene certificados adjuntos.
                </p>
              ) : (
                <ul className="flex flex-col divide-y divide-outline-variant">
                  {certificados.map((certificado) => (
                    <li
                      key={certificado.id}
                      className="flex flex-wrap items-center justify-between gap-3 py-3"
                    >
                      <span className="text-body-md text-on-surface">{certificado.etiqueta}</span>
                      <Button
                        variante="secondary"
                        icono="download"
                        cargando={descargando === certificado.id}
                        textoCargando="Descargando..."
                        onClick={() => void descargar(certificado)}
                      >
                        Descargar PDF
                      </Button>
                    </li>
                  ))}
                </ul>
              )}
            </Card>
          </div>
          <DocumentoPostulacion postulacion={postulacion} ubicacion={ubicacion} />
        </>
      )}
    </main>
  )
}
