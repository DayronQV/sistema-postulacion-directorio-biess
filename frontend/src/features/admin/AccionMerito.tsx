import { useState } from 'react'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import { ModalConfirmacion } from '@/components/ui/Modal'
import { useApi } from '@/hooks/useApi'
import { marcarMerito, type ResultadoMerito } from '@/services/adminPostulaciones'
import { useSesion } from '@/store/sesion'

interface AccionMeritoProps {
  postulacionId: string
  // Código de la postulación, para la confirmación (opcional).
  codigo?: string
  superoMerito: boolean
  // Solo una postulación enviada se puede marcar; en otro estado se muestra sin acción.
  enviada: boolean
  onMarcado?: (resultado: ResultadoMerito) => void
}

const llamada = (signal: AbortSignal, id: string, supero: boolean) =>
  marcarMerito(signal, id, supero)

// Marca o quita la marca de "superó la Fase de Méritos" de una postulación (EF-13). Pensado
// para la nómina y el detalle de postulaciones: el administrador la cambia con
// confirmación; el analista solo ve el estado (el servidor también le responde 403).
export default function AccionMerito({
  postulacionId,
  codigo,
  superoMerito,
  enviada,
  onMarcado,
}: AccionMeritoProps) {
  const esAdministrador = useSesion()?.rol === 'administrador'
  const [supero, setSupero] = useState(superoMerito)
  const [confirmando, setConfirmando] = useState(false)
  const { ejecutar, cargando, error } = useApi(llamada)

  const nuevo = !supero
  const confirmar = async () => {
    const r = await ejecutar(postulacionId, nuevo)
    setConfirmando(false)
    if (r) {
      setSupero(r.supero_merito)
      onMarcado?.(r)
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-2">
        <Badge tono={supero ? 'success' : 'neutral'}>
          {supero ? 'Superó la Fase de Méritos' : 'Fase de Méritos pendiente'}
        </Badge>
        {esAdministrador && enviada && (
          <Button
            variante={supero ? 'ghost' : 'secondary'}
            icono={supero ? 'undo' : 'verified'}
            cargando={cargando}
            textoCargando="Guardando..."
            onClick={() => setConfirmando(true)}
          >
            {supero ? 'Quitar la marca' : 'Marcar que superó'}
          </Button>
        )}
      </div>
      {error && <Alert tipo="error">{error.message}</Alert>}

      <ModalConfirmacion
        abierto={confirmando}
        titulo={nuevo ? 'Marcar que superó la Fase de Méritos' : 'Quitar la marca de méritos'}
        mensaje={
          nuevo
            ? `La postulación ${codigo ?? ''} pasará a "Evaluación Técnica" y el postulante quedará habilitado para la evaluación teórica.`
            : `La postulación ${codigo ?? ''} volverá a "Enviada" y el postulante dejará de estar habilitado para la evaluación teórica.`
        }
        textoConfirmar={nuevo ? 'Marcar' : 'Quitar la marca'}
        peligroso={!nuevo}
        cargando={cargando}
        onConfirmar={() => void confirmar()}
        onCancelar={() => setConfirmando(false)}
      />
    </div>
  )
}
