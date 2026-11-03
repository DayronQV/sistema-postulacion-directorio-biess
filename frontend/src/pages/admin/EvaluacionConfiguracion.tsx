import { useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import Campo from '@/components/form/Campo'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Input from '@/components/ui/Input'
import Loader from '@/components/ui/Loader'
import { useApi } from '@/hooks/useApi'
import { guardarConfig, obtenerConfig, type ConfigEvaluacion } from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'

const MIN_DURACION = 1
const MAX_DURACION = 480

function entero(texto: string): number | null {
  return /^\d+$/.test(texto.trim()) ? Number(texto) : null
}

// Configuración de la evaluación teórica (EF-14): duración y número de preguntas que se
// sortean a cada postulante. La fecha y la hora en que se puede rendir las fija la fase
// Evaluación, en la pantalla de fases.
export default function EvaluacionConfiguracion() {
  const { data: config, cargando, error, ejecutar } = useApi(obtenerConfig)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <div>
        <h1 className="text-headline-lg text-primary">Configuración de la evaluación</h1>
        <p className="mt-1 text-body-md text-on-surface-variant">
          Defina la duración del examen y cuántas preguntas recibe cada postulante.
        </p>
      </div>

      {cargando && <Loader texto="Cargando la configuración..." />}
      {error && (
        <Alert tipo="error" titulo="No se pudo cargar la configuración">
          <div className="flex flex-col items-start gap-2">
            <span>{error.message}</span>
            <Button variante="secondary" icono="refresh" onClick={() => void ejecutar()}>
              Reintentar
            </Button>
          </div>
        </Alert>
      )}

      {config && <FormularioConfig config={config} />}
    </main>
  )
}

// Formulario con los valores que devolvió el servidor como punto de partida.
function FormularioConfig({ config }: { config: ConfigEvaluacion }) {
  const [duracion, setDuracion] = useState(String(config.duracion_minutos))
  const [preguntas, setPreguntas] = useState(String(config.num_preguntas))
  const [errores, setErrores] = useState<Record<string, string>>({})
  const [errorGeneral, setErrorGeneral] = useState('')
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [bancoTotal, setBancoTotal] = useState(config.total_banco)
  const [configurada, setConfigurada] = useState(config.configurada)

  async function guardar(e: FormEvent) {
    e.preventDefault()
    setGuardado(false)
    setErrorGeneral('')
    const nuevos: Record<string, string> = {}
    const minutos = entero(duracion)
    const cantidad = entero(preguntas)
    if (minutos === null || minutos < MIN_DURACION || minutos > MAX_DURACION) {
      nuevos.duracion_minutos = `Debe ser un número entero de ${MIN_DURACION} a ${MAX_DURACION}`
    }
    if (cantidad === null || cantidad < 1) {
      nuevos.num_preguntas = 'Debe ser un número entero mayor que cero'
    }
    setErrores(nuevos)
    if (minutos === null || cantidad === null || Object.keys(nuevos).length > 0) return

    setGuardando(true)
    try {
      const nueva = await guardarConfig(new AbortController().signal, {
        duracion_minutos: minutos,
        num_preguntas: cantidad,
      })
      setBancoTotal(nueva.total_banco)
      setConfigurada(nueva.configurada)
      setGuardado(true)
    } catch (e) {
      if (e instanceof ApiError) {
        setErrores(e.fields ?? {})
        setErrorGeneral(e.message)
      } else setErrorGeneral('No se pudo guardar la configuración.')
    } finally {
      setGuardando(false)
    }
  }

  return (
    <Card titulo="Parámetros del examen">
      <form noValidate onSubmit={guardar} className="flex flex-col gap-space-md">
        {!configurada && (
          <Alert tipo="info" titulo="Aún no se ha configurado">
            Se muestran los valores por defecto. Guarde la configuración para fijarlos.
          </Alert>
        )}
        {errorGeneral && (
          <Alert tipo="error" titulo="No se pudo guardar">
            {errorGeneral}
          </Alert>
        )}
        {guardado && <Alert tipo="success">Configuración guardada.</Alert>}

        <div className="grid gap-space-md sm:grid-cols-2">
          <Campo
            id="config-duracion"
            etiqueta="Duración (minutos)"
            requerido
            error={errores.duracion_minutos}
            ayuda={`De ${MIN_DURACION} a ${MAX_DURACION} minutos`}
          >
            <Input
              id="config-duracion"
              inputMode="numeric"
              value={duracion}
              invalido={Boolean(errores.duracion_minutos)}
              onChange={(e) => setDuracion(e.target.value)}
            />
          </Campo>
          <Campo
            id="config-preguntas"
            etiqueta="Número de preguntas"
            requerido
            error={errores.num_preguntas}
            ayuda={`El banco tiene ${bancoTotal} preguntas`}
          >
            <Input
              id="config-preguntas"
              inputMode="numeric"
              value={preguntas}
              invalido={Boolean(errores.num_preguntas)}
              onChange={(e) => setPreguntas(e.target.value)}
            />
          </Campo>
        </div>

        <p className="text-body-md text-on-surface-variant">
          La fecha y la hora en que se puede rendir el examen se fijan en la fase Evaluación, desde{' '}
          <Link to="/admin/fases" className="font-semibold text-primary-container underline">
            Fases
          </Link>
          .
        </p>

        <div>
          <Button type="submit" icono="save" cargando={guardando} textoCargando="Guardando...">
            Guardar configuración
          </Button>
        </div>
      </form>
    </Card>
  )
}
