import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Loader from '@/components/ui/Loader'
import { cx } from '@/lib/cx'
import type { PreguntaExamen } from '@/services/evaluacion'

interface PreguntaActualProps {
  pregunta: PreguntaExamen | null
  cargando: boolean
  guardando: boolean
  elegida: number | undefined
  onElegir: (opcion: number) => void
  onAnterior: () => void
  onSiguiente: () => void
}

// Una pregunta con sus opciones. Elegir una opción la guarda en el servidor; la elegida queda
// marcada al volver a la pregunta.
export default function PreguntaActual({
  pregunta,
  cargando,
  guardando,
  elegida,
  onElegir,
  onAnterior,
  onSiguiente,
}: PreguntaActualProps) {
  if (!pregunta) return cargando ? <Loader texto="Cargando la pregunta..." /> : null
  const esPrimera = pregunta.numero <= 1
  const esUltima = pregunta.numero >= pregunta.total
  return (
    <Card titulo={`Pregunta ${pregunta.numero} de ${pregunta.total}`}>
      <div className="flex flex-col gap-space-lg select-none">
        <p className="text-body-lg text-on-surface">{pregunta.enunciado}</p>

        <div role="radiogroup" aria-label="Opciones de respuesta" className="flex flex-col gap-3">
          {pregunta.opciones.map((opcion, i) => {
            const activa = elegida === i
            return (
              <label
                key={i}
                className={cx(
                  'flex cursor-pointer items-start gap-3 rounded-xl border p-4 text-body-md transition-colors',
                  activa
                    ? 'border-primary-container bg-surface-container-low'
                    : 'border-outline-variant hover:bg-surface-container-low',
                  (guardando || cargando) && 'cursor-wait opacity-80',
                )}
              >
                <input
                  type="radio"
                  name="opcion"
                  checked={activa}
                  disabled={guardando || cargando}
                  onChange={() => onElegir(i)}
                  className="mt-0.5 h-5 w-5 shrink-0 accent-primary-container"
                />
                <span>{opcion}</span>
              </label>
            )
          })}
        </div>

        <div className="flex items-center justify-between gap-3">
          <Button
            variante="secondary"
            icono="chevron_left"
            disabled={esPrimera || guardando || cargando}
            onClick={onAnterior}
          >
            Anterior
          </Button>
          <Button
            variante="secondary"
            disabled={esUltima || guardando || cargando}
            onClick={onSiguiente}
          >
            Siguiente
          </Button>
        </div>
      </div>
    </Card>
  )
}
