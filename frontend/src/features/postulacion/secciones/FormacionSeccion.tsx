import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import { useFormacion } from '../useFormacion'
import Formacion from './Formacion'

// Formación académica conectada a la API: el contenedor solo la monta.
export default function FormacionSeccion() {
  const { cargando, error, recargar, props } = useFormacion()

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
  return props && <Formacion {...props} />
}
