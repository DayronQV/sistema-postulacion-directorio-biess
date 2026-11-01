import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import { useInformacionPersonal } from '../useInformacionPersonal'
import InformacionPersonal from './InformacionPersonal'
import PostulacionBloqueada from './PostulacionBloqueada'

// Información personal conectada a la API: carga la postulación y, si está cerrada por
// nacionalidad, muestra la pantalla de bloqueo en lugar de la sección.
export default function InformacionPersonalSeccion() {
  const s = useInformacionPersonal()

  if (s.cargando) return <Loader texto="Cargando su postulación..." />
  if (s.errorCarga) {
    return (
      <div className="flex flex-col gap-space-sm">
        <Alert tipo="error" titulo="No se pudo cargar la postulación">
          {s.errorCarga.message}
        </Alert>
        <Button variante="secondary" onClick={() => void s.recargar()}>
          Reintentar
        </Button>
      </div>
    )
  }
  if (s.bloqueo) return <PostulacionBloqueada motivo={s.bloqueo} codigo={s.codigo} />
  return s.props && <InformacionPersonal {...s.props} />
}
