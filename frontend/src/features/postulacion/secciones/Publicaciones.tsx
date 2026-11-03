import { useEffect, useRef, useState } from 'react'
import Campo from '@/components/form/Campo'
import DocumentoUpload from '@/components/form/DocumentoUpload'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import DateInput from '@/components/ui/DateInput'
import Input from '@/components/ui/Input'
import RadioSiNo from '@/components/ui/RadioSiNo'
import Select from '@/components/ui/Select'
import TablaFilas from '@/components/ui/TablaFilas'
import { fechaLocalISO } from '@/features/auth/esquemas'
import { eliminarDocumento, type DocumentoSubido } from '@/services/documentos'
import { ApiError } from '@/services/errores'
import type { PublicacionAcademica, TipoPublicacion } from '@/services/postulacion'

// Igual que domain.MaxPublicaciones en el servidor.
export const MAX_PUBLICACIONES = 50

export interface PublicacionesProps {
  tienePublicaciones: boolean
  valor: PublicacionAcademica[]
  onGuardar: (tiene: boolean, publicaciones: PublicacionAcademica[]) => Promise<void> | void
}

// En el editor "relacionada" empieza sin responder (null): es una pregunta Sí/No
// obligatoria y no debe quedar en "No" por omisión.
type FilaEditor = Omit<PublicacionAcademica, 'documento_id' | 'relacionada'> & {
  relacionada: boolean | null
  documento: DocumentoSubido | null
}

const tipos = [
  { valor: 'publicacion', etiqueta: 'Publicación' },
  { valor: 'investigacion', etiqueta: 'Trabajo de investigación' },
] as const

const etiquetaTipo = Object.fromEntries(tipos.map((t) => [t.valor, t.etiqueta]))

const vacia = (): FilaEditor => ({
  titulo: '',
  tipo: '' as TipoPublicacion,
  medio: '',
  fecha: '',
  relacionada: null,
  documento: null,
})

// La API solo devuelve documento_id: el nombre es descriptivo y el tamaño 0 indica a
// FileUpload que no lo muestre.
function alEditor(fila: PublicacionAcademica): FilaEditor {
  return {
    ...fila,
    documento: {
      documento_id: fila.documento_id,
      nombre: `respaldo-${fila.titulo || 'publicacion'}.pdf`,
      tamano_bytes: 0,
    },
  }
}

function alDominio(fila: FilaEditor): PublicacionAcademica {
  return {
    titulo: fila.titulo.trim(),
    tipo: fila.tipo,
    medio: fila.medio.trim(),
    fecha: fila.fecha,
    relacionada: fila.relacionada === true,
    documento_id: fila.documento?.documento_id ?? '',
  }
}

// Borra PDF que ya no usa ninguna fila. Un fallo (por ejemplo 409 si sigue en uso) se
// ignora: lo peor es que el archivo quede guardado.
function descartar(ids: string[]) {
  for (const id of ids) eliminarDocumento(new AbortController().signal, id).catch(() => undefined)
}

function erroresDe(fila: FilaEditor) {
  const errores: Partial<Record<keyof FilaEditor, string>> = {}
  if (!fila.titulo.trim()) errores.titulo = 'Campo obligatorio'
  if (!fila.tipo) errores.tipo = 'Seleccione el tipo'
  if (!fila.medio.trim()) errores.medio = 'Campo obligatorio'
  if (!fila.fecha) errores.fecha = 'Ingrese la fecha'
  else if (fila.fecha > fechaLocalISO()) errores.fecha = 'La fecha no puede ser futura'
  if (fila.relacionada === null) errores.relacionada = 'Seleccione Sí o No'
  if (!fila.documento) errores.documento = 'Adjunte el documento de respaldo'
  return errores
}

export default function Publicaciones({
  tienePublicaciones,
  valor,
  onGuardar,
}: PublicacionesProps) {
  const [tiene, setTiene] = useState(tienePublicaciones)
  // Con "No" las filas se ocultan pero no se borran hasta guardar: si la respuesta fue un
  // error y se vuelve a "Sí", siguen ahí.
  const [filas, setFilas] = useState(valor)
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  // Hay cambios que todavía no se guardaron: la respuesta Sí/No o filas eliminadas. Agregar o
  // editar una fila la guarda al momento. Solo entonces se ofrece "Guardar información", y
  // con "Sí" solo si queda alguna fila: sin filas no hay nada que guardar hasta agregar una.
  const [pendiente, setPendiente] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)
  // PDF subidos en esta pantalla y todavía no guardados (ver Formacion).
  const subidos = useRef(new Set<string>())

  useEffect(
    () => () => {
      const pendientes = [...subidos.current]
      subidos.current.clear()
      descartar(pendientes)
    },
    [],
  )

  function cambiarRespuesta(nueva: boolean) {
    setTiene(nueva)
    if (!nueva) setIndice(undefined)
    setGuardado(false)
    setPendiente(true)
  }

  function abrir(fila?: PublicacionAcademica) {
    setIndice(fila ? filas.indexOf(fila) : null)
    setBorrador(fila ? alEditor(fila) : vacia())
    setErrores({})
    setGuardado(false)
  }

  // Guarda la respuesta Sí/No y la lista en el servidor. Devuelve false si falla, y entonces
  // la lista en pantalla no cambia.
  async function guardar(lista: PublicacionAcademica[]): Promise<boolean> {
    if (tiene && lista.length === 0) {
      setErrorGuardar(
        new ApiError(400, 'VALIDATION_ERROR', 'Agregue al menos una publicación o investigación.'),
      )
      return false
    }
    // Con "No" el servidor borra las filas: se descartan también las que quedaron ocultas.
    const enviadas = tiene ? lista : []
    setGuardando(true)
    setGuardado(false)
    setErrorGuardar(null)
    try {
      await onGuardar(tiene, enviadas)
      const usados = new Set(enviadas.map((fila) => fila.documento_id))
      const sobrantes = [...subidos.current].filter((id) => !usados.has(id))
      subidos.current.clear()
      descartar(sobrantes)
      setFilas(enviadas)
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
    const encontrados = erroresDe(borrador)
    setErrores(encontrados)
    if (Object.keys(encontrados).length > 0) return
    const fila = alDominio(borrador)
    const lista =
      indice === null
        ? [...filas, fila]
        : filas.map((actual, posicion) => (posicion === indice ? fila : actual))
    if (await guardar(lista)) setIndice(undefined)
  }

  const actualizar = <K extends keyof FilaEditor>(campo: K, valorCampo: FilaEditor[K]) =>
    setBorrador((actual) => ({ ...actual, [campo]: valorCampo }))

  const erroresServidor = errorGuardar?.fields
    ? Object.entries(errorGuardar.fields).filter(([campo]) => campo.startsWith('publicaciones'))
    : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby="publicaciones-titulo">
      <div className="flex flex-col gap-3">
        <h2 id="publicaciones-titulo" className="text-headline-sm text-primary">
          Publicaciones e investigaciones
        </h2>
        <p id="publicaciones-pregunta" className="text-body-md text-on-surface-variant">
          ¿Cuenta con publicaciones e investigaciones académicas?
        </p>
        <RadioSiNo
          aria-labelledby="publicaciones-pregunta"
          valor={tiene}
          onChange={cambiarRespuesta}
          deshabilitado={guardando}
        />
      </div>

      {tiene && (
        <>
          <TablaFilas
            titulo="Publicaciones registradas"
            columnas={[
              { clave: 'titulo', titulo: 'Título' },
              { clave: 'tipo', titulo: 'Tipo', render: (fila) => etiquetaTipo[fila.tipo] },
              { clave: 'medio', titulo: 'Medio' },
              { clave: 'fecha', titulo: 'Fecha' },
              {
                clave: 'relacionada',
                titulo: 'Relacionada',
                render: (fila) => (fila.relacionada ? 'Sí' : 'No'),
              },
            ]}
            filas={filas}
            obtenerClave={(fila) => filas.indexOf(fila)}
            onAgregar={filas.length >= MAX_PUBLICACIONES ? undefined : () => abrir()}
            onEditar={abrir}
            onEliminar={(fila) => {
              setFilas((actuales) => actuales.filter((actual) => actual !== fila))
              setIndice(undefined)
              setGuardado(false)
              setPendiente(true)
            }}
            textoAgregar="Agregar publicación"
            deshabilitado={guardando}
          />

          {indice !== undefined && (
            <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
              <Campo id="publicacion-titulo" etiqueta="Título" requerido error={errores.titulo}>
                <Input
                  id="publicacion-titulo"
                  value={borrador.titulo}
                  onChange={(e) => actualizar('titulo', e.target.value)}
                  invalido={Boolean(errores.titulo)}
                />
              </Campo>
              <Campo id="publicacion-tipo" etiqueta="Tipo" requerido error={errores.tipo}>
                <Select
                  id="publicacion-tipo"
                  opciones={[...tipos]}
                  placeholder="Seleccione..."
                  value={borrador.tipo}
                  onChange={(e) => actualizar('tipo', e.target.value as TipoPublicacion)}
                  invalido={Boolean(errores.tipo)}
                />
              </Campo>
              <Campo
                id="publicacion-medio"
                etiqueta="Medio de publicación"
                requerido
                error={errores.medio}
              >
                <Input
                  id="publicacion-medio"
                  value={borrador.medio}
                  onChange={(e) => actualizar('medio', e.target.value)}
                  invalido={Boolean(errores.medio)}
                />
              </Campo>
              <Campo id="publicacion-fecha" etiqueta="Fecha" requerido error={errores.fecha}>
                <DateInput
                  id="publicacion-fecha"
                  max={fechaLocalISO()}
                  value={borrador.fecha}
                  onChange={(e) => actualizar('fecha', e.target.value)}
                  invalido={Boolean(errores.fecha)}
                />
              </Campo>
              <div className="md:col-span-2">
                <Campo
                  id="publicacion-relacionada"
                  etiqueta="¿Guarda relación con el campo de su ejercicio o formación profesional?"
                  requerido
                  error={errores.relacionada}
                >
                  <RadioSiNo
                    id="publicacion-relacionada"
                    aria-labelledby="publicacion-relacionada-etiqueta"
                    aria-invalid={Boolean(errores.relacionada) || undefined}
                    valor={borrador.relacionada}
                    onChange={(relacionada) => actualizar('relacionada', relacionada)}
                  />
                </Campo>
              </div>
              <div className="md:col-span-2">
                <DocumentoUpload
                  tipo="publicacion"
                  maxMb={2}
                  valor={borrador.documento}
                  onChange={(documento) => {
                    if (documento) subidos.current.add(documento.documento_id)
                    actualizar('documento', documento)
                  }}
                  etiqueta="Documento de respaldo (PDF, máximo 2 MB)"
                  eliminarEnServidor={false}
                />
                {errores.documento && (
                  <p role="alert" className="mt-1 text-label-sm text-error">
                    {errores.documento}
                  </p>
                )}
              </div>
              <div className="flex justify-end gap-3 md:col-span-2">
                <Button
                  variante="secondary"
                  disabled={guardando}
                  onClick={() => setIndice(undefined)}
                >
                  Cancelar
                </Button>
                <Button onClick={guardarFila} cargando={guardando} textoCargando="Guardando...">
                  Guardar
                </Button>
              </div>
            </div>
          )}
        </>
      )}

      {guardado && <Alert tipo="success">La información se guardó correctamente.</Alert>}
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
      {indice === undefined && pendiente && (!tiene || filas.length > 0) && (
        <div className="flex justify-end">
          <Button onClick={() => guardar(filas)} cargando={guardando} textoCargando="Guardando...">
            Guardar información
          </Button>
        </div>
      )}
    </section>
  )
}
