import { validarIdentidad } from '@/services/auth'
import FormularioIdentidad from './FormularioIdentidad'

interface PasoIdentidadProps {
  onValidado: (ticket: string, cedula: string) => void
}

// Paso 1 del registro: valida la identidad y recibe el ticket para el paso 2.
export default function PasoIdentidad({ onValidado }: PasoIdentidadProps) {
  return (
    <FormularioIdentidad
      enviar={validarIdentidad}
      onExito={onValidado}
      textoBoton="Validar identidad"
      mensajeBloqueo="registro fallido, inténtelo nuevamente en 10 minutos"
    />
  )
}
