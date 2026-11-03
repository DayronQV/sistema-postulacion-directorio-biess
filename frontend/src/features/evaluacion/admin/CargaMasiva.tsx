import { useRef, useState, type ChangeEvent } from 'react'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import { cargarPreguntasCsv } from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import { analizarCsv, erroresDeCarga, MAX_BYTES_CSV, type AnalisisCsv } from './preguntas'

const PLANTILLA =
  'enunciado,opciones,respuesta_correcta\n"¿Capital de Ecuador?",Guayaquil|Quito|Cuenca,1\n'
const ENLACE_PLANTILLA = `data:text/csv;charset=utf-8,${encodeURIComponent(PLANTILLA)}`

interface CargaMasivaProps {
  // Se llama cuando el servidor guardó las preguntas, para refrescar el listado.
  onCargado: () => void
}

function resumen(texto: string) {
  return texto.length > 90 ? `${texto.slice(0, 90)}…` : texto
}

// Carga masiva del banco desde un archivo CSV: lee el archivo en el navegador para mostrar
// una vista previa con los errores por línea y solo deja cargar si todas las filas son válidas.
// El servidor guarda todas las preguntas o ninguna.
export default function CargaMasiva({ onCargado }: CargaMasivaProps) {
  const entrada = useRef<HTMLInputElement>(null)
  const [archivo, setArchivo] = useState<File | null>(null)
  const [analisis, setAnalisis] = useState<AnalisisCsv | null>(null)
  const [errorArchivo, setErrorArchivo] = useState('')
  const [cargando, setCargando] = useState(false)
  const [erroresServidor, setErroresServidor] = useState<string[]>([])
  const [errorGeneral, setErrorGeneral] = useState('')
  const [creadas, setCreadas] = useState<number | null>(null)

  function limpiar() {
    setArchivo(null)
    setAnalisis(null)
    setErrorArchivo('')
    setErroresServidor([])
    setErrorGeneral('')
    if (entrada.current) entrada.current.value = ''
  }

  async function elegir(e: ChangeEvent<HTMLInputElement>) {
    const elegido = e.target.files?.[0]
    limpiar()
    setCreadas(null)
    if (!elegido) return
    if (elegido.size > MAX_BYTES_CSV) {
      setErrorArchivo('El archivo supera 1 MB')
      return
    }
    setArchivo(elegido)
    setAnalisis(analizarCsv(await elegido.text()))
  }

  async function cargar() {
    if (!archivo) return
    setCargando(true)
    setErroresServidor([])
    setErrorGeneral('')
    try {
      const { creadas: total } = await cargarPreguntasCsv(new AbortController().signal, archivo)
      limpiar()
      setCreadas(total)
      onCargado()
    } catch (e) {
      if (e instanceof ApiError && e.fields) setErroresServidor(erroresDeCarga(e.fields))
      setErrorGeneral(e instanceof ApiError ? e.message : 'No se pudo cargar el archivo.')
    } finally {
      setCargando(false)
    }
  }

  const filas = analisis?.filas ?? []
  const conErrores = filas.filter((f) => f.errores.length > 0).length
  const puedeCargar =
    archivo !== null && analisis !== null && !analisis.errorGeneral && conErrores === 0

  return (
    <Card
      titulo="Carga masiva desde archivo"
      descripcion="Archivo CSV con las columnas enunciado, opciones (separadas por |) y respuesta_correcta (índice desde 0)."
    >
      <div className="flex flex-col gap-space-md">
        <div className="flex flex-wrap items-center gap-3">
          <label className="flex flex-col gap-1.5 text-label-md text-on-surface">
            Archivo CSV
            <input
              ref={entrada}
              type="file"
              accept=".csv,text/csv"
              onChange={(e) => void elegir(e)}
              className="text-body-md"
            />
          </label>
          <a
            href={ENLACE_PLANTILLA}
            download="plantilla-preguntas.csv"
            className="text-body-bold text-primary-container hover:underline"
          >
            Descargar plantilla
          </a>
        </div>

        {errorArchivo && <Alert tipo="error">{errorArchivo}</Alert>}
        {creadas !== null && (
          <Alert tipo="success" titulo="Carga completa">
            Se cargaron {creadas} pregunta{creadas === 1 ? '' : 's'} al banco.
          </Alert>
        )}
        {analisis?.errorGeneral && <Alert tipo="error">{analisis.errorGeneral}</Alert>}
        {errorGeneral && (
          <Alert tipo="error" titulo={errorGeneral}>
            {erroresServidor.length > 0 && (
              <ul className="mt-1 list-disc pl-5">
                {erroresServidor.map((linea) => (
                  <li key={linea}>{linea}</li>
                ))}
              </ul>
            )}
          </Alert>
        )}

        {filas.length > 0 && (
          <>
            <p className="text-body-md text-on-surface-variant">
              {filas.length} pregunta{filas.length === 1 ? '' : 's'} en el archivo
              {conErrores > 0 ? `; ${conErrores} con errores` : '; todas son válidas'}.
            </p>
            <div className="max-h-96 overflow-auto">
              <table className="w-full min-w-[640px] border-collapse text-left text-body-md">
                <thead>
                  <tr className="border-b border-outline-variant text-label-md text-on-surface-variant">
                    <th className="px-3 py-2">Línea</th>
                    <th className="px-3 py-2">Enunciado</th>
                    <th className="px-3 py-2">Opciones</th>
                    <th className="px-3 py-2">Correcta</th>
                    <th className="px-3 py-2">Estado</th>
                  </tr>
                </thead>
                <tbody>
                  {filas.map((f) => (
                    <tr key={f.linea} className="border-b border-outline-variant/60 align-top">
                      <td className="px-3 py-2 font-mono text-label-md">{f.linea}</td>
                      <td className="px-3 py-2">{resumen(f.enunciado)}</td>
                      <td className="px-3 py-2">{f.opciones.length}</td>
                      <td className="px-3 py-2">{f.respuesta_correcta ?? '—'}</td>
                      <td className="px-3 py-2">
                        {f.errores.length === 0 ? (
                          <Badge tono="success">Válida</Badge>
                        ) : (
                          <ul className="list-disc pl-4 text-label-sm text-error">
                            {f.errores.map((error) => (
                              <li key={error}>{error}</li>
                            ))}
                          </ul>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        )}

        <div className="flex gap-3">
          <Button
            icono="upload"
            disabled={!puedeCargar}
            cargando={cargando}
            onClick={() => void cargar()}
          >
            Cargar {filas.length > 0 ? filas.length : ''} preguntas
          </Button>
          {archivo && (
            <Button variante="secondary" disabled={cargando} onClick={limpiar}>
              Quitar archivo
            </Button>
          )}
        </div>
      </div>
    </Card>
  )
}
