import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import { usePublicaciones } from '../usePublicaciones'
import Publicaciones from './Publicaciones'

// Publicaciones e investigaciones conectada a la API: el contenedor solo la monta.
export default function PublicacionesSeccion() {
  const { cargando, error, recargar, props } = usePublicaciones()

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
  return props && <Publicaciones {...props} />
}
