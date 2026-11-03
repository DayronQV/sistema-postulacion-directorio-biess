import { useCallback, useEffect, useState } from 'react'
import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import Loader from '@/components/ui/Loader'
import { ApiError } from '@/services/errores'
import { obtenerDashboard, type DashboardPostulaciones } from '@/services/adminDashboard'

const INTERVALO_ACTUALIZACION = 30_000

const etiquetasCargo = {
  afiliados: 'Afiliados',
  jubilados: 'Jubilados',
  empleador: 'Empleador',
}

const etiquetasEstado = {
  borrador: 'Borrador',
  enviada: 'Enviada',
  cerrada: 'Cerrada',
}

interface BarraProps {
  etiqueta: string
  valor: number
  maximo: number
}

function Barra({ etiqueta, valor, maximo }: BarraProps) {
  const porcentaje = maximo === 0 ? 0 : Math.round((valor / maximo) * 100)
  return (
    <div className="grid grid-cols-[7rem_1fr_2rem] items-center gap-space-sm">
      <span className="text-label-md text-on-surface">{etiqueta}</span>
      <div
        className="h-3 overflow-hidden rounded-full bg-surface-container-high"
        aria-hidden="true"
      >
        <div
          className="h-full rounded-full bg-primary-container transition-[width] duration-300"
          style={{ width: `${porcentaje}%` }}
        />
      </div>
      <span className="text-right text-body-bold text-primary" aria-label={`${etiqueta}: ${valor}`}>
        {valor}
      </span>
    </div>
  )
}

export default function Dashboard() {
  const [datos, setDatos] = useState<DashboardPostulaciones>()
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')

  const cargar = useCallback((signal: AbortSignal) => {
    obtenerDashboard(signal)
      .then((respuesta) => {
        if (signal.aborted) return
        setDatos(respuesta)
        setError('')
      })
      .catch((fallo: unknown) => {
        if (signal.aborted) return
        setError(
          fallo instanceof ApiError
            ? fallo.message
            : 'No se pudieron cargar los indicadores del dashboard.',
        )
      })
      .finally(() => {
        if (!signal.aborted) setCargando(false)
      })
  }, [])

  useEffect(() => {
    let control = new AbortController()
    cargar(control.signal)
    const intervalo = window.setInterval(() => {
      control.abort()
      control = new AbortController()
      cargar(control.signal)
    }, INTERVALO_ACTUALIZACION)
    return () => {
      window.clearInterval(intervalo)
      control.abort()
    }
  }, [cargar])

  const maximoCargo = datos ? Math.max(...Object.values(datos.por_cargo)) : 0

  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <div>
        <h1 className="text-headline-lg text-primary">Dashboard de postulantes</h1>
        <p className="mt-1 text-body-md text-on-surface-variant">
          Los indicadores se actualizan automáticamente cada 30 segundos.
        </p>
      </div>

      {error && (
        <Alert tipo="error" titulo="No se pudo actualizar el dashboard">
          {error}
        </Alert>
      )}
      {cargando && !datos ? (
        <Loader texto="Cargando indicadores..." />
      ) : (
        datos && (
          <>
            <section
              aria-label="Resumen"
              className="grid gap-space-md sm:grid-cols-2 lg:grid-cols-4"
            >
              <Card titulo="Total">
                <p className="text-display text-primary">{datos.total}</p>
              </Card>
              {Object.entries(datos.por_estado).map(([estado, valor]) => (
                <Card key={estado} titulo={etiquetasEstado[estado as keyof typeof etiquetasEstado]}>
                  <p className="text-display text-primary">{valor}</p>
                </Card>
              ))}
            </section>

            <Card
              titulo="Postulantes por cargo"
              descripcion="Distribución según el cargo seleccionado"
            >
              <div className="flex flex-col gap-space-md">
                {Object.entries(datos.por_cargo).map(([cargo, valor]) => (
                  <Barra
                    key={cargo}
                    etiqueta={etiquetasCargo[cargo as keyof typeof etiquetasCargo]}
                    valor={valor}
                    maximo={maximoCargo}
                  />
                ))}
              </div>
            </Card>
          </>
        )
      )}
    </main>
  )
}
