import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import type { ApiError } from '@/services/errores'

const formatoFecha = new Intl.DateTimeFormat('es-EC', { dateStyle: 'long', timeStyle: 'short' })

// Cuándo empieza o cerró la evaluación, si el servidor lo informa en `fields`.
function cuando(error: ApiError): string | null {
  const iso = error.fields?.inicia_en ?? error.fields?.cierra_en
  const fecha = iso ? new Date(iso) : null
  return fecha && !Number.isNaN(fecha.getTime()) ? formatoFecha.format(fecha) : null
}

interface InstruccionesExamenProps {
  iniciando: boolean
  error: ApiError | null
  onIniciar: () => void
  // Si ya rindió, el servidor responde 409 y se ofrece ver su resultado.
  onVerResultado: () => void
  cargandoResultado: boolean
}

// Pantalla previa al examen: instrucciones y el botón "Iniciar Evaluación".
export default function InstruccionesExamen({
  iniciando,
  error,
  onIniciar,
  onVerResultado,
  cargandoResultado,
}: InstruccionesExamenProps) {
  const fecha = error ? cuando(error) : null
  return (
    <Card
      titulo="Evaluación técnica"
      descripcion="Lea las instrucciones antes de empezar. Tiene un solo intento."
    >
      <div className="flex flex-col gap-space-md">
        <ul className="flex list-disc flex-col gap-2 pl-5 text-body-md text-on-surface">
          <li>
            Verá las preguntas una por una, y puede volver a una pregunta para cambiar su respuesta.
          </li>
          <li>Cada respuesta se guarda en cuanto la elige.</li>
          <li>
            Un reloj en cuenta regresiva queda fijo arriba. Al llegar a 00:00 el examen se cierra y
            se califica solo, con lo que haya respondido.
          </li>
          <li>
            Durante el examen no se puede usar el botón «Atrás» del navegador ni copiar, cortar,
            pegar o abrir el menú contextual.
          </li>
          <li>Si recarga la página, el examen sigue en curso con el tiempo que le queda.</li>
          <li>Al terminar recibirá su nota y el reporte llegará a su correo.</li>
        </ul>

        {error && (
          <Alert tipo={error.status === 409 ? 'info' : 'error'} titulo={error.message}>
            {fecha && <p>Fecha y hora: {fecha}</p>}
            {error.status === 409 && (
              <div className="mt-2">
                <Button variante="secondary" cargando={cargandoResultado} onClick={onVerResultado}>
                  Ver mi resultado
                </Button>
              </div>
            )}
          </Alert>
        )}

        <div>
          <Button
            icono="play_arrow"
            cargando={iniciando}
            textoCargando="Iniciando..."
            onClick={onIniciar}
          >
            Iniciar Evaluación
          </Button>
        </div>
      </div>
    </Card>
  )
}
