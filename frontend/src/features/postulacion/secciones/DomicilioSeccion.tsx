import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import { useDomicilio } from '../useDomicilio'
import Domicilio from './Domicilio'

// Domicilio conectado a la API: el contenedor del formulario solo lo monta.
export default function DomicilioSeccion() {
  const d = useDomicilio()

  if (d.cargando) return <Loader texto="Cargando su postulación..." />
  if (d.errorCarga) {
    return (
      <div className="flex flex-col gap-space-sm">
        <Alert tipo="error" titulo="No se pudo cargar la postulación">
          {d.errorCarga.message}
        </Alert>
        <Button variante="secondary" onClick={() => void d.recargar()}>
          Reintentar
        </Button>
      </div>
    )
  }
  return (
    <Domicilio
      valoresIniciales={d.valoresIniciales}
      correoPrincipal={d.correoPrincipal}
      onGuardar={d.onGuardar}
      guardando={d.guardando}
      guardado={d.guardado}
      error={d.errorGuardado}
    />
  )
}
