import { useEffect } from 'react'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Loader from '@/components/ui/Loader'
import TarjetaFase from '@/features/admin/TarjetaFase'
import { useApi } from '@/hooks/useApi'
import { listarFases } from '@/services/adminFases'
import { useSesion } from '@/store/sesion'

// Fases del concurso (EF-13): el administrador las publica, fija sus fechas y habilita el
// formulario de postulación; el analista solo las consulta.
export default function Fases() {
  const editable = useSesion()?.rol === 'administrador'
  const { data: fases, ejecutar, cargando, error } = useApi(listarFases)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <h1 className="text-headline-lg text-primary">Fases del concurso</h1>
      {!editable && (
        <Alert tipo="info" titulo="Acceso de solo lectura">
          Como analista puede consultar las fases; solo el administrador las modifica.
        </Alert>
      )}

      {cargando && <Loader texto="Cargando las fases..." />}
      {error && (
        <Alert tipo="error" titulo="No se pudieron cargar las fases">
          <div className="flex flex-col items-start gap-2">
            <span>{error.message}</span>
            <Button variante="secondary" icono="refresh" onClick={() => void ejecutar()}>
              Reintentar
            </Button>
          </div>
        </Alert>
      )}
      {fases?.map((f) => (
        <TarjetaFase key={f.id} fase={f} editable={editable} />
      ))}
    </main>
  )
}
