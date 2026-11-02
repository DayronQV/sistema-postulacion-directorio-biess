import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import { limpiarErrorGuardado, useFormulario } from '../store'

interface PieSeccionProps {
  textoGuardar?: string
  deshabilitado?: boolean
}

// Botón "Guardar Progreso" con su aviso de error. Va dentro del <form> de cada sección.
export default function PieSeccion({
  textoGuardar = 'Guardar Progreso',
  deshabilitado,
}: PieSeccionProps) {
  const { guardando, errorGuardado } = useFormulario()
  return (
    <div className="flex flex-col gap-3">
      {errorGuardado && (
        <Alert tipo="error" titulo="No se pudo guardar" onCerrar={limpiarErrorGuardado}>
          {errorGuardado.message}
        </Alert>
      )}
      <div>
        <Button
          type="submit"
          icono="save"
          cargando={guardando}
          textoCargando="Guardando..."
          disabled={deshabilitado}
        >
          {textoGuardar}
        </Button>
      </div>
    </div>
  )
}
