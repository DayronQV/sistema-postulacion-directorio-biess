import { useEffect, useRef, useState } from 'react'
import Campo from '@/components/form/Campo'
import DocumentoUpload from '@/components/form/DocumentoUpload'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import DateInput from '@/components/ui/DateInput'
import Input from '@/components/ui/Input'
import Select from '@/components/ui/Select'
import TablaFilas from '@/components/ui/TablaFilas'
import { fechaLocalISO } from '@/features/auth/esquemas'
import { eliminarDocumento, type DocumentoSubido } from '@/services/documentos'
import { ApiError } from '@/services/errores'
import type { PublicacionAcademica, TipoPublicacion } from '@/services/postulacion'

export interface PublicacionesProps {
  tienePublicaciones: boolean
  valor: PublicacionAcademica[]
  onGuardar: (tiene: boolean, publicaciones: PublicacionAcademica[]) => Promise<void> | void
}

type FilaEditor = Omit<PublicacionAcademica, 'documento_id'> & { documento: DocumentoSubido | null }

const vacia = (): FilaEditor => ({ titulo: '', tipo: '' as TipoPublicacion, medio: '', fecha: '', relacionada: false, documento: null })

function alEditor(fila: PublicacionAcademica): FilaEditor {
  return { ...fila, documento: { documento_id: fila.documento_id, nombre: `respaldo-${fila.titulo || 'publicacion'}.pdf`, tamano_bytes: 0 } }
}

function alDominio(fila: FilaEditor): PublicacionAcademica {
  return { titulo: fila.titulo.trim(), tipo: fila.tipo, medio: fila.medio.trim(), fecha: fila.fecha, relacionada: fila.relacionada, documento_id: fila.documento?.documento_id ?? '' }
}

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
  if (!fila.documento) errores.documento = 'Adjunte el documento de respaldo'
  return errores
}

export default function Publicaciones({ tienePublicaciones, valor, onGuardar }: PublicacionesProps) {
  const [tiene, setTiene] = useState(tienePublicaciones)
  const [filas, setFilas] = useState(valor)
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)
  const subidos = useRef(new Set<string>())

  useEffect(() => () => { const pendientes = [...subidos.current]; subidos.current.clear(); descartar(pendientes) }, [])

  function cambiarRespuesta(valorNuevo: boolean) {
    setTiene(valorNuevo)
    if (!valorNuevo) { setFilas([]); setIndice(undefined) }
    setGuardado(false)
  }

  function abrir(fila?: PublicacionAcademica) {
    setIndice(fila ? filas.indexOf(fila) : null)
    setBorrador(fila ? alEditor(fila) : vacia())
    setErrores({})
    setGuardado(false)
  }

  function aceptarFila() {
    const encontrados = erroresDe(borrador)
    setErrores(encontrados)
    if (Object.keys(encontrados).length) return
    const fila = alDominio(borrador)
    setFilas((actuales) => indice === null ? [...actuales, fila] : actuales.map((actual, i) => i === indice ? fila : actual))
    setIndice(undefined)
  }

  async function guardar() {
    if (tiene && filas.length === 0) {
      setErrorGuardar(new ApiError(400, 'VALIDATION_ERROR', 'Agregue al menos una publicación o investigación.'))
      return
    }
    setGuardando(true); setGuardado(false); setErrorGuardar(null)
    try {
      await onGuardar(tiene, tiene ? filas : [])
      const usados = new Set((tiene ? filas : []).map((fila) => fila.documento_id))
      const sobrantes = [...subidos.current].filter((id) => !usados.has(id))
      subidos.current.clear(); descartar(sobrantes); setGuardado(true)
    } catch (error) {
      setErrorGuardar(error instanceof ApiError ? error : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la información.'))
    } finally { setGuardando(false) }
  }

  const actualizar = <K extends keyof FilaEditor>(campo: K, valor: FilaEditor[K]) => setBorrador((actual) => ({ ...actual, [campo]: valor }))
  const erroresServidor = errorGuardar?.fields ? Object.entries(errorGuardar.fields).filter(([campo]) => campo.startsWith('publicaciones')) : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby="publicaciones-titulo">
      <div>
        <h2 id="publicaciones-titulo" className="text-headline-sm text-primary">Publicaciones e investigaciones</h2>
        <p className="text-body-md text-on-surface-variant">¿Cuenta con publicaciones e investigaciones académicas?</p>
      </div>
      <div className="flex gap-6">
        <label><input type="radio" name="tiene-publicaciones" checked={tiene} onChange={() => cambiarRespuesta(true)} /> Sí</label>
        <label><input type="radio" name="tiene-publicaciones" checked={!tiene} onChange={() => cambiarRespuesta(false)} /> No</label>
      </div>

      {tiene && (
        <>
          <TablaFilas
            titulo="Publicaciones registradas"
            columnas={[
              { clave: 'titulo', titulo: 'Título' },
              { clave: 'tipo', titulo: 'Tipo', render: (fila) => fila.tipo === 'publicacion' ? 'Publicación' : 'Trabajo de investigación' },
              { clave: 'medio', titulo: 'Medio' },
              { clave: 'fecha', titulo: 'Fecha' },
              { clave: 'relacionada', titulo: 'Relacionada', render: (fila) => fila.relacionada ? 'Sí' : 'No' },
            ]}
            filas={filas}
            obtenerClave={(fila) => filas.indexOf(fila)}
            onAgregar={() => abrir()}
            onEditar={abrir}
            onEliminar={(fila) => { setFilas((actuales) => actuales.filter((actual) => actual !== fila)); setIndice(undefined); setGuardado(false) }}
            textoAgregar="Agregar publicación"
            deshabilitado={guardando}
          />
          {indice !== undefined && (
            <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
              <Campo id="publicacion-titulo" etiqueta="Título" requerido error={errores.titulo}>
                <Input id="publicacion-titulo" value={borrador.titulo} onChange={(e) => actualizar('titulo', e.target.value)} invalido={Boolean(errores.titulo)} />
              </Campo>
              <Campo id="publicacion-tipo" etiqueta="Tipo" requerido error={errores.tipo}>
                <Select id="publicacion-tipo" opciones={[{ valor: 'publicacion', etiqueta: 'Publicación' }, { valor: 'investigacion', etiqueta: 'Trabajo de investigación' }]} placeholder="Seleccione..." value={borrador.tipo} onChange={(e) => actualizar('tipo', e.target.value as TipoPublicacion)} invalido={Boolean(errores.tipo)} />
              </Campo>
              <Campo id="publicacion-medio" etiqueta="Medio de publicación" requerido error={errores.medio}>
                <Input id="publicacion-medio" value={borrador.medio} onChange={(e) => actualizar('medio', e.target.value)} invalido={Boolean(errores.medio)} />
              </Campo>
              <Campo id="publicacion-fecha" etiqueta="Fecha" requerido error={errores.fecha}>
                <DateInput id="publicacion-fecha" max={fechaLocalISO()} value={borrador.fecha} onChange={(e) => actualizar('fecha', e.target.value)} invalido={Boolean(errores.fecha)} />
              </Campo>
              <Campo id="publicacion-relacionada" etiqueta="¿Guarda relación con su ejercicio o formación profesional?" requerido>
                <Select id="publicacion-relacionada" opciones={[{ valor: 'true', etiqueta: 'Sí' }, { valor: 'false', etiqueta: 'No' }]} value={String(borrador.relacionada)} onChange={(e) => actualizar('relacionada', e.target.value === 'true')} />
              </Campo>
              <div className="md:col-span-2">
                <DocumentoUpload tipo="publicacion" maxMb={2} valor={borrador.documento} onChange={(documento) => { if (documento) subidos.current.add(documento.documento_id); actualizar('documento', documento) }} etiqueta="Documento de respaldo (PDF, máximo 2 MB)" eliminarEnServidor={false} />
                {errores.documento && <p className="mt-1 text-label-sm text-error">{errores.documento}</p>}
              </div>
              <div className="flex justify-end gap-3 md:col-span-2"><Button variante="secondary" onClick={() => setIndice(undefined)}>Cancelar</Button><Button onClick={aceptarFila}>{indice === null ? 'Agregar' : 'Actualizar'}</Button></div>
            </div>
          )}
        </>
      )}

      {guardado && <Alert tipo="success">La información se guardó correctamente.</Alert>}
      {errorGuardar && <Alert tipo="error" titulo={errorGuardar.message}>{erroresServidor.length > 0 && <ul className="list-disc pl-4">{erroresServidor.map(([campo, mensaje]) => <li key={campo}>{campo}: {mensaje}</li>)}</ul>}</Alert>}
      <div className="flex justify-end"><Button onClick={guardar} cargando={guardando} textoCargando="Guardando...">Guardar información</Button></div>
    </section>
  )
}
