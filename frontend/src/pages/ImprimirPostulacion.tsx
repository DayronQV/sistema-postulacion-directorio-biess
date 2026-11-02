import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Icon from '@/components/ui/Icon'
import Loader from '@/components/ui/Loader'
import DocumentoPostulacion from '@/features/postulacion/imprimible/DocumentoPostulacion'
import { cargarPostulacion, useFormulario } from '@/features/postulacion/store'

// Vista final del formulario para imprimir (A4) o guardar como PDF desde el navegador.
// Los controles llevan print:hidden para que no salgan en el papel.
export default function ImprimirPostulacion() {
  const { postulacion, cargando, errorCarga } = useFormulario()

  useEffect(() => {
    if (!postulacion) void cargarPostulacion()
    // Solo al entrar: si ya está en el store se reutiliza.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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
        <Button icono="print" onClick={() => window.print()} disabled={!postulacion}>
          Imprimir / Guardar PDF
        </Button>
      </div>

      {cargando && !postulacion && <Loader texto="Preparando el formulario..." />}
      {errorCarga && !postulacion && (
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
      {postulacion && <DocumentoPostulacion postulacion={postulacion} />}
    </main>
  )
}
