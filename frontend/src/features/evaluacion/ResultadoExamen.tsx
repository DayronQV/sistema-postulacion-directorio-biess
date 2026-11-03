import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import type { ResultadoExamen as Resultado } from '@/services/evaluacion'

const formatoFecha = new Intl.DateTimeFormat('es-EC', { dateStyle: 'long', timeStyle: 'short' })

// Pantalla final: la nota (porcentaje de aciertos sobre 100) y el aviso del reporte por correo.
export default function ResultadoExamen({ resultado }: { resultado: Resultado }) {
  return (
    <Card titulo="Evaluación finalizada">
      <div className="flex flex-col items-start gap-space-md">
        <div className="flex flex-col gap-1">
          <span className="text-label-md text-on-surface-variant">Su nota</span>
          <span className="text-display-sm font-bold text-primary" aria-label="Nota">
            {resultado.nota.toFixed(2)} / 100
          </span>
          <span className="text-body-md text-on-surface">
            {resultado.aciertos} aciertos de {resultado.total} preguntas
          </span>
        </div>
        <Alert tipo="success" titulo="Reporte enviado">
          El reporte de su resultado llegó a su correo. Finalizó el{' '}
          {formatoFecha.format(new Date(resultado.finalizado_at))}.
        </Alert>
        <Link to="/postulacion" className="font-semibold text-secondary underline">
          Volver a mi postulación
        </Link>
      </div>
    </Card>
  )
}
