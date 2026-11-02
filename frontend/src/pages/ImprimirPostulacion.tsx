import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Icon from '@/components/ui/Icon'
import Loader from '@/components/ui/Loader'
import DocumentoPostulacion from '@/features/postulacion/imprimible/DocumentoPostulacion'
import { useUbicacion } from '@/features/postulacion/imprimible/useUbicacion'
import { cargarPostulacion, useFormulario } from '@/features/postulacion/store'

// Vista final del formulario para imprimir (A4) o guardar como PDF desde el navegador.
// Los controles llevan print:hidden para que no salgan en el papel.
export default function ImprimirPostulacion() {
  const { postulacion, cargando, errorCarga } = useFormulario()
  const ubicacion = useUbicacion(postulacion?.domicilio)

  // Siempre se vuelve a leer al entrar: lo que imprime debe ser lo guardado en el servidor,
  // no una copia anterior que haya quedado en memoria (por ejemplo, de otra sesión).
  useEffect(() => {
    void cargarPostulacion()
  }, [])
  const listo = postulacion !== null && !cargando

  return (
    <main className="min-h-screen bg-surface-container p-margin-mobile print:bg-white print:p-0 sm:p-margin">
      <div className="mx-auto mb-space-md flex max-w-[210mm] flex-wrap items-center justify-between gap-3 print:hidden">
        <Link
          to="/postulacion"
          className="inline-flex items-center gap-1.5 text-body-bold text-primary-container hover:underline"
        >
          <Icon nombre="arrow_back" className="text-[18px]" />
          Volver al formulario
        </Link>
        <Button icono="print" onClick={() => window.print()} disabled={!listo}>
          Imprimir / Guardar PDF
        </Button>
      </div>

      {cargando && <Loader texto="Preparando el formulario..." />}
      {errorCarga && !cargando && (
        <div className="mx-auto flex max-w-[210mm] flex-col gap-3">
          <Alert tipo="error" titulo="No se pudo cargar la postulación">
            {errorCarga.message}
          </Alert>
          <div>
            <Button variante="secondary" icono="refresh" onClick={() => void cargarPostulacion()}>
              Reintentar
            </Button>
          </div>
        </div>
      )}
      {listo && !errorCarga && (
        <DocumentoPostulacion postulacion={postulacion} ubicacion={ubicacion} />
      )}
    </main>
  )
}
