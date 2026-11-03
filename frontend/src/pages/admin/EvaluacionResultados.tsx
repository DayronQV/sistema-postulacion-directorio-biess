import { useEffect, useState } from 'react'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Loader from '@/components/ui/Loader'
import { useApi } from '@/hooks/useApi'
import {
  descargarReporte,
  listarResultados,
  type ResultadoEvaluacion,
} from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import EncabezadoPagina from '@/components/layout/EncabezadoPagina'

const etiquetasCargo: Record<string, string> = {
  afiliados: 'Afiliados',
  jubilados: 'Jubilados',
  empleador: 'Empleador',
}

const formatoFecha = new Intl.DateTimeFormat('es-EC', { dateStyle: 'medium', timeStyle: 'short' })

function fecha(valor: string | null) {
  return valor ? formatoFecha.format(new Date(valor)) : '—'
}

// Resultados de la evaluación teórica (EF-14): nota de cada postulante y descarga del
// reporte PDF de los exámenes ya finalizados.
export default function EvaluacionResultados() {
  const { data: resultados, cargando, error, ejecutar } = useApi(listarResultados)
  const [descargando, setDescargando] = useState('')
  const [errorDescarga, setErrorDescarga] = useState('')

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  async function descargar(r: ResultadoEvaluacion) {
    setDescargando(r.intento_id)
    setErrorDescarga('')
    try {
      const archivo = await descargarReporte(new AbortController().signal, r.intento_id)
      const url = URL.createObjectURL(archivo.contenido)
      const enlace = document.createElement('a')
      enlace.href = url
      enlace.download = archivo.nombre
      enlace.click()
      // Revocar en el mismo instante puede cortar la descarga en algunos navegadores.
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    } catch (e) {
      setErrorDescarga(e instanceof ApiError ? e.message : 'No se pudo descargar el reporte.')
    } finally {
      setDescargando('')
    }
  }

  return (
    <main className="mx-auto flex w-full max-w-[1400px] flex-1 flex-col gap-space-md p-margin-mobile sm:p-margin">
      <EncabezadoPagina
        icono="grading"
        titulo="Resultados de la evaluación"
        descripcion="Notas de los postulantes que iniciaron el examen. La nota es el porcentaje de aciertos."
        acciones={
          <Button
            variante="secondary"
            icono="refresh"
            disabled={cargando}
            onClick={() => void ejecutar()}
          >
            Actualizar
          </Button>
        }
      />

      {cargando && !resultados && <Loader texto="Cargando resultados..." />}
      {error && (
        <Alert tipo="error" titulo="No se pudieron cargar los resultados">
          {error.message}
        </Alert>
      )}
      {errorDescarga && (
        <Alert tipo="error" onCerrar={() => setErrorDescarga('')}>
          {errorDescarga}
        </Alert>
      )}

      {resultados && (
        <Card titulo={`${resultados.length} examen${resultados.length === 1 ? '' : 'es'}`}>
          <div className="overflow-x-auto">
            <table className="w-full min-w-[900px] border-collapse text-left text-body-md">
              <thead>
                <tr className="border-b border-outline-variant text-label-md text-on-surface-variant">
                  <th className="px-3 py-2">Código</th>
                  <th className="px-3 py-2">Postulante</th>
                  <th className="px-3 py-2">Cargo</th>
                  <th className="px-3 py-2">Estado</th>
                  <th className="px-3 py-2">Nota</th>
                  <th className="px-3 py-2">Aciertos</th>
                  <th className="px-3 py-2">Finalizó</th>
                  <th className="px-3 py-2 text-right">Reporte</th>
                </tr>
              </thead>
              <tbody>
                {resultados.map((r) => (
                  <tr key={r.intento_id} className="border-b border-outline-variant/60">
                    <td className="px-3 py-3 font-mono text-label-md">{r.codigo}</td>
                    <td className="px-3 py-3">
                      {r.nombres} {r.apellidos}
                      <span className="block text-label-sm text-on-surface-variant">
                        {r.cedula}
                      </span>
                    </td>
                    <td className="px-3 py-3">{etiquetasCargo[r.cargo] ?? r.cargo}</td>
                    <td className="px-3 py-3">
                      <Badge tono={r.estado === 'finalizado' ? 'success' : 'warning'}>
                        {r.estado === 'finalizado' ? 'Finalizado' : 'En curso'}
                      </Badge>
                    </td>
                    <td className="px-3 py-3 font-semibold">
                      {r.nota !== null ? `${r.nota.toFixed(2)} / 100` : '—'}
                    </td>
                    <td className="px-3 py-3">
                      {r.aciertos !== null ? `${r.aciertos} de ${r.total_preguntas}` : '—'}
                    </td>
                    <td className="px-3 py-3">{fecha(r.finalizado_at)}</td>
                    <td className="px-3 py-3 text-right">
                      {r.estado === 'finalizado' ? (
                        <Button
                          variante="secondary"
                          icono="picture_as_pdf"
                          aria-label={`Descargar el reporte de ${r.nombres} ${r.apellidos}`}
                          cargando={descargando === r.intento_id}
                          onClick={() => void descargar(r)}
                        >
                          PDF
                        </Button>
                      ) : (
                        <span className="text-label-sm text-on-surface-variant">
                          Disponible al terminar
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {resultados.length === 0 && (
              <p className="py-space-xl text-center text-body-md text-on-surface-variant">
                Todavía ningún postulante inició la evaluación.
              </p>
            )}
          </div>
        </Card>
      )}
    </main>
  )
}
