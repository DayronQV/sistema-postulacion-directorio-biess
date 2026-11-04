import { useState } from 'react'
import Campo from '@/components/form/Campo'
import Alert from '@/components/ui/Alert'
import { useAvisoDeGuardado } from '@/hooks/useAvisoDeGuardado'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import PanelEditor from '@/components/ui/PanelEditor'
import TablaFilas from '@/components/ui/TablaFilas'
import { ApiError } from '@/services/errores'
import type { ReferenciaPersonal } from '@/services/postulacion'

export type { ReferenciaPersonal }

// Límite de la sección (issue #31). El mínimo de una referencia se exige al enviar la
// postulación; en borrador se puede guardar la lista vacía.
export const MAX_REFERENCIAS = 3

export interface ReferenciasProps {
  valor: ReferenciaPersonal[]
  onGuardar: (referencias: ReferenciaPersonal[]) => Promise<void> | void
}

const vacia = (): ReferenciaPersonal => ({
  nombres_apellidos: '',
  relacion: '',
  telefono: '',
  correo: '',
})

const limpiar = (r: ReferenciaPersonal): ReferenciaPersonal => ({
  nombres_apellidos: r.nombres_apellidos.trim(),
  relacion: r.relacion.trim(),
  telefono: r.telefono.trim(),
  correo: r.correo.trim(),
})

// Las mismas reglas que valida el servidor (referencias_validacion.go); el servidor
// vuelve a validar siempre.
function erroresDe(referencia: ReferenciaPersonal) {
  const errores: Partial<Record<keyof ReferenciaPersonal, string>> = {}
  if (!referencia.nombres_apellidos) errores.nombres_apellidos = 'Campo obligatorio'
  if (!referencia.relacion) errores.relacion = 'Campo obligatorio'
  if (!referencia.telefono) errores.telefono = 'Campo obligatorio'
  else if (!/^\+?[0-9]{7,15}$/.test(referencia.telefono)) errores.telefono = 'Teléfono no válido'
  if (!referencia.correo) errores.correo = 'Campo obligatorio'
  else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(referencia.correo))
    errores.correo = 'Correo no válido'
  return errores
}

export default function Referencias({ valor, onGuardar }: ReferenciasProps) {
  const [filas, setFilas] = useState(valor)
  // undefined: editor cerrado; null: agregando; número: editando esa fila.
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  useAvisoDeGuardado(guardado, 'La información se guardó correctamente.')
  // Hay cambios en la lista que todavía no se guardaron (hoy, solo filas eliminadas).
  // Agregar o editar una fila la guarda al momento. Solo entonces se ofrece
  // "Guardar información".
  const [pendiente, setPendiente] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)
  const completa = filas.length >= MAX_REFERENCIAS

  function abrir(referencia?: ReferenciaPersonal) {
    const posicion = referencia ? filas.indexOf(referencia) : null
    setIndice(posicion)
    setBorrador(referencia ? { ...referencia } : vacia())
    setErrores({})
    setGuardado(false)
  }

  function eliminar(referencia: ReferenciaPersonal) {
    setFilas((actuales) => actuales.filter((fila) => fila !== referencia))
    setIndice(undefined)
    setGuardado(false)
    setPendiente(true)
  }

  // Guarda la lista en el servidor. Devuelve false si falla, y entonces la lista en pantalla
  // no cambia.
  async function guardar(lista: ReferenciaPersonal[]): Promise<boolean> {
    setGuardando(true)
    setGuardado(false)
    setErrorGuardar(null)
    try {
      await onGuardar(lista)
      setFilas(lista)
      setPendiente(false)
      setGuardado(true)
      return true
    } catch (error) {
      setErrorGuardar(
        error instanceof ApiError
          ? error
          : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la información.'),
      )
      return false
    } finally {
      setGuardando(false)
    }
  }

  // "Guardar" del formulario: valida, agrega o actualiza la fila y guarda la lista. Si el
  // servidor la rechaza, el formulario sigue abierto con lo que escribió la persona.
  async function guardarFila() {
    const limpia = limpiar(borrador)
    const encontrados = erroresDe(limpia)
    setErrores(encontrados)
    if (Object.keys(encontrados).length > 0) return
    const lista =
      indice === null
        ? [...filas, limpia]
        : filas.map((actual, posicion) => (posicion === indice ? limpia : actual))
    if (await guardar(lista)) setIndice(undefined)
  }

  const actualizar = (campo: keyof ReferenciaPersonal, valorCampo: string) =>
    setBorrador((actual) => ({ ...actual, [campo]: valorCampo }))

  const erroresServidor = errorGuardar?.fields
    ? Object.entries(errorGuardar.fields).filter(([campo]) => campo.startsWith('referencias'))
    : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby="referencias-titulo">
      <div>
        <h2 id="referencias-titulo" className="text-headline-sm text-primary">
          Referencias personales
        </h2>
        <p className="text-body-md text-on-surface-variant">
          Registre entre 1 y {MAX_REFERENCIAS} personas de contacto.
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
        // Cada fila es un objeto distinto, así que su posición es una clave estable y no
        // se repite aunque dos referencias tengan los mismos datos.
        obtenerClave={(referencia) => filas.indexOf(referencia)}
        // Al llegar al máximo solo se oculta Agregar: Editar y Eliminar siguen disponibles.
        onAgregar={completa ? undefined : () => abrir()}
        onEditar={abrir}
        onEliminar={eliminar}
        textoAgregar="Agregar referencia"
        titulo="Referencias"
        deshabilitado={guardando}
      />
      {completa && (
        <p className="text-label-sm text-on-surface-variant">
          Ya registró el máximo de {MAX_REFERENCIAS} referencias.
        </p>
      )}

      {indice !== undefined && (
        <PanelEditor
          clave={indice}
          className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2"
        >
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
                aria-describedby={errores[campo] ? `referencia-${campo}-error` : undefined}
              />
            </Campo>
          ))}
          <div className="flex justify-end gap-3 md:col-span-2">
            <Button variante="secondary" disabled={guardando} onClick={() => setIndice(undefined)}>
              Cancelar
            </Button>
            <Button onClick={guardarFila} cargando={guardando} textoCargando="Guardando...">
              Guardar
            </Button>
          </div>
        </PanelEditor>
      )}

      {errorGuardar && (
        <Alert tipo="error" titulo={errorGuardar.message}>
          {erroresServidor.length > 0 && (
            <ul className="list-disc pl-4">
              {erroresServidor.map(([campo, mensaje]) => (
                <li key={campo}>
                  {campo}: {mensaje}
                </li>
              ))}
            </ul>
          )}
        </Alert>
      )}
      {indice === undefined && pendiente && (
        <div className="flex justify-end">
          <Button onClick={() => guardar(filas)} cargando={guardando} textoCargando="Guardando...">
            Guardar información
          </Button>
        </div>
      )}
    </section>
  )
}
