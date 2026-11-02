import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import { useCargo } from '../useCargo'
import Cargo from './Cargo'
import PostulacionBloqueada from './PostulacionBloqueada'

// Selección del cargo conectada a la API: carga la postulación y, si el servidor la cerró
// (el postulante no cumple la calidad del cargo, o su nacionalidad), muestra la pantalla de
// bloqueo en lugar de la sección.
export default function CargoSeccion() {
  const s = useCargo()

  if (s.cargando) return <Loader texto="Cargando su postulación..." />
  if (s.error) {
    return (
      <div className="flex flex-col gap-space-sm">
        <Alert tipo="error" titulo="No se pudo cargar la postulación">
          {s.error.message}
        </Alert>
        <Button variante="secondary" onClick={s.recargar}>
          Reintentar
        </Button>
      </div>
    )
  }
  if (s.bloqueo) return <PostulacionBloqueada motivo={s.bloqueo} codigo={s.codigo} />
  return s.props && <Cargo {...s.props} />
}
