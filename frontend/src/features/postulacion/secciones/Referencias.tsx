import { useState } from 'react'
import Campo from '@/components/form/Campo'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import TablaFilas from '@/components/ui/TablaFilas'

export interface ReferenciaPersonal {
  nombres_apellidos: string
  relacion: string
  telefono: string
  correo: string
}

interface ReferenciasProps {
  valor: ReferenciaPersonal[]
  onGuardar: (referencias: ReferenciaPersonal[]) => Promise<void> | void
}

const vacia = (): ReferenciaPersonal => ({
  nombres_apellidos: '',
  relacion: '',
  telefono: '',
  correo: '',
})

function erroresDe(referencia: ReferenciaPersonal) {
  const errores: Partial<Record<keyof ReferenciaPersonal, string>> = {}
  if (!referencia.nombres_apellidos.trim()) errores.nombres_apellidos = 'Campo obligatorio'
  if (!referencia.relacion.trim()) errores.relacion = 'Campo obligatorio'
  if (!/^\+?[0-9]{7,15}$/.test(referencia.telefono.trim())) errores.telefono = 'Teléfono no válido'
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(referencia.correo.trim()))
    errores.correo = 'Correo no válido'
  return errores
}

export default function Referencias({ valor, onGuardar }: ReferenciasProps) {
  const [filas, setFilas] = useState(valor)
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState(false)

  function abrir(referencia?: ReferenciaPersonal) {
    const posicion = referencia ? filas.indexOf(referencia) : null
    setIndice(posicion)
    setBorrador(referencia ? { ...referencia } : vacia())
    setErrores({})
    setGuardado(false)
  }

  function aceptarFila() {
    const encontrados = erroresDe(borrador)
    setErrores(encontrados)
    if (Object.keys(encontrados).length > 0) return
    setFilas((actuales) =>
      indice === null
        ? [...actuales, borrador]
        : actuales.map((fila, posicion) => (posicion === indice ? borrador : fila)),
    )
    setIndice(undefined)
  }

  async function guardar() {
    setGuardando(true)
    setGuardado(false)
    setErrorGuardar(false)
    try {
      await onGuardar(filas)
      setGuardado(true)
    } catch {
      setErrorGuardar(true)
    } finally {
      setGuardando(false)
    }
  }

  const actualizar = (campo: keyof ReferenciaPersonal, valorCampo: string) =>
    setBorrador((actual) => ({ ...actual, [campo]: valorCampo }))

  return (
    <section className="flex flex-col gap-5" aria-labelledby="referencias-titulo">
      <div>
        <h2 id="referencias-titulo" className="text-headline-sm text-primary">
          Referencias personales
        </h2>
        <p className="text-body-md text-on-surface-variant">
          Registre entre 1 y 3 personas de contacto.
        </p>
      </div>

      <TablaFilas
        columnas={[
          { clave: 'nombres_apellidos', titulo: 'Nombres y apellidos' },
          { clave: 'relacion', titulo: 'Relación' },
          { clave: 'telefono', titulo: 'Teléfono' },
          { clave: 'correo', titulo: 'Correo' },
        ]}
        filas={filas}
        obtenerClave={(referencia) =>
          `${referencia.nombres_apellidos}-${referencia.telefono}-${referencia.correo}`
        }
        onAgregar={() => abrir()}
        onEditar={abrir}
        onEliminar={(referencia) =>
          setFilas((actuales) => actuales.filter((fila) => fila !== referencia))
        }
        textoAgregar="Agregar referencia"
        deshabilitado={filas.length >= 3 || guardando}
      />

      {indice !== undefined && (
        <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
          {(
            [
              ['nombres_apellidos', 'Nombres y apellidos'],
              ['relacion', 'Relación'],
              ['telefono', 'Teléfono'],
              ['correo', 'Correo'],
            ] as const
          ).map(([campo, etiqueta]) => (
            <Campo
              key={campo}
              id={`referencia-${campo}`}
              etiqueta={etiqueta}
              requerido
              error={errores[campo]}
            >
              <Input
                id={`referencia-${campo}`}
                value={borrador[campo]}
                onChange={(evento) => actualizar(campo, evento.target.value)}
                inputMode={campo === 'telefono' ? 'tel' : campo === 'correo' ? 'email' : undefined}
                invalido={Boolean(errores[campo])}
              />
            </Campo>
          ))}
          <div className="flex justify-end gap-3 md:col-span-2">
            <Button variante="secondary" onClick={() => setIndice(undefined)}>
              Cancelar
            </Button>
            <Button onClick={aceptarFila}>{indice === null ? 'Agregar' : 'Actualizar'}</Button>
          </div>
        </div>
      )}

      {guardado && <Alert tipo="success">La información se guardó correctamente.</Alert>}
      {errorGuardar && <Alert tipo="error">No se pudo guardar la información.</Alert>}
      <div className="flex justify-end">
        <Button onClick={guardar} cargando={guardando} textoCargando="Guardando...">
          Guardar información
        </Button>
      </div>
    </section>
  )
}
