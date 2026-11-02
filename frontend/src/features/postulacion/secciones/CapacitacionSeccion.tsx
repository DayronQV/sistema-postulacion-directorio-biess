import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import { useCapacitacion } from '../useCapacitacion'
import Capacitacion from './Capacitacion'

// Capacitación conectada a la API: el contenedor solo la monta.
export default function CapacitacionSeccion() {
  const { cargando, error, recargar, props } = useCapacitacion()

  if (cargando || (!props && !error)) return <Loader texto="Cargando su postulación..." />
  if (error) {
    return (
      <div className="flex flex-col gap-space-sm">
        <Alert tipo="error" titulo="No se pudo cargar la postulación">
          {error.message}
        </Alert>
        <Button variante="secondary" onClick={recargar}>
          Reintentar
        </Button>
      </div>
    )
  }
  return props && <Capacitacion {...props} />
}
