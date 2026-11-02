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
import type { Capacitacion as CapacitacionDatos, TipoCapacitacion } from '@/services/postulacion'

export interface CapacitacionProps {
  valor: CapacitacionDatos[]
  onGuardar: (capacitaciones: CapacitacionDatos[]) => Promise<void> | void
}

type FilaEditor = Omit<CapacitacionDatos, 'documento_id'> & { documento: DocumentoSubido | null }
const vacia = (): FilaEditor => ({ tipo: '' as TipoCapacitacion, nombre: '', institucion: '', fecha_inicio: '', fecha_fin: '', horas: 0, documento: null })
function alEditor(fila: CapacitacionDatos): FilaEditor { return { ...fila, documento: fila.documento_id ? { documento_id: fila.documento_id, nombre: `respaldo-${fila.nombre || 'capacitacion'}.pdf`, tamano_bytes: 0 } : null } }
function alDominio(fila: FilaEditor): CapacitacionDatos { return { tipo: fila.tipo, nombre: fila.nombre.trim(), institucion: fila.institucion.trim(), fecha_inicio: fila.fecha_inicio, fecha_fin: fila.fecha_fin, horas: fila.horas, documento_id: fila.documento?.documento_id ?? '' } }
function descartar(ids: string[]) { for (const id of ids) eliminarDocumento(new AbortController().signal, id).catch(() => undefined) }
function erroresDe(fila: FilaEditor) {
  const errores: Partial<Record<keyof FilaEditor, string>> = {}
  if (!fila.tipo) errores.tipo = 'Seleccione el tipo de evento'
  if (!fila.nombre.trim()) errores.nombre = 'Campo obligatorio'
  if (!fila.institucion.trim()) errores.institucion = 'Campo obligatorio'
  if (!fila.fecha_inicio) errores.fecha_inicio = 'Ingrese la fecha inicial'; else if (fila.fecha_inicio > fechaLocalISO()) errores.fecha_inicio = 'La fecha no puede ser futura'
  if (!fila.fecha_fin) errores.fecha_fin = 'Ingrese la fecha final'; else if (fila.fecha_fin > fechaLocalISO()) errores.fecha_fin = 'La fecha no puede ser futura'; else if (fila.fecha_inicio && fila.fecha_fin < fila.fecha_inicio) errores.fecha_fin = 'La fecha final no puede ser anterior a la inicial'
  if (!Number.isInteger(fila.horas) || fila.horas <= 0) errores.horas = 'Ingrese un número de horas mayor a cero'
  return errores
}

export default function Capacitacion({ valor, onGuardar }: CapacitacionProps) {
  const [filas, setFilas] = useState(valor)
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)
  const subidos = useRef(new Set<string>())
  useEffect(() => () => { const pendientes = [...subidos.current]; subidos.current.clear(); descartar(pendientes) }, [])

  function abrir(fila?: CapacitacionDatos) { setIndice(fila ? filas.indexOf(fila) : null); setBorrador(fila ? alEditor(fila) : vacia()); setErrores({}); setGuardado(false) }
  function aceptarFila() { const encontrados = erroresDe(borrador); setErrores(encontrados); if (Object.keys(encontrados).length) return; const fila = alDominio(borrador); setFilas((actuales) => indice === null ? [...actuales, fila] : actuales.map((actual, i) => i === indice ? fila : actual)); setIndice(undefined) }
  async function guardar() {
    setGuardando(true); setGuardado(false); setErrorGuardar(null)
    try { await onGuardar(filas); const usados = new Set(filas.map((fila) => fila.documento_id)); const sobrantes = [...subidos.current].filter((id) => !usados.has(id)); subidos.current.clear(); descartar(sobrantes); setGuardado(true) }
    catch (error) { setErrorGuardar(error instanceof ApiError ? error : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la capacitación.')) }
    finally { setGuardando(false) }
  }
  const actualizar = <K extends keyof FilaEditor>(campo: K, valorCampo: FilaEditor[K]) => setBorrador((actual) => ({ ...actual, [campo]: valorCampo }))
  const erroresServidor = errorGuardar?.fields ? Object.entries(errorGuardar.fields).filter(([campo]) => campo.startsWith('capacitaciones')) : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby="capacitacion-titulo">
      <div><h2 id="capacitacion-titulo" className="text-headline-sm text-primary">Capacitación</h2><p className="text-body-md text-on-surface-variant">Registre sus cursos, seminarios y talleres.</p></div>
      <TablaFilas titulo="Eventos registrados" columnas={[{ clave: 'tipo', titulo: 'Tipo' }, { clave: 'nombre', titulo: 'Nombre' }, { clave: 'institucion', titulo: 'Institución' }, { clave: 'fecha_inicio', titulo: 'Desde' }, { clave: 'fecha_fin', titulo: 'Hasta' }, { clave: 'horas', titulo: 'Horas' }]} filas={filas} obtenerClave={(fila) => filas.indexOf(fila)} onAgregar={() => abrir()} onEditar={abrir} onEliminar={(fila) => { setFilas((actuales) => actuales.filter((actual) => actual !== fila)); setIndice(undefined); setGuardado(false) }} textoAgregar="Agregar capacitación" deshabilitado={guardando} />
      {indice !== undefined && (
        <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
          <Campo id="capacitacion-tipo" etiqueta="Tipo de evento" requerido error={errores.tipo}><Select id="capacitacion-tipo" opciones={[{ valor: 'curso', etiqueta: 'Curso' }, { valor: 'seminario', etiqueta: 'Seminario' }, { valor: 'taller', etiqueta: 'Taller' }]} placeholder="Seleccione..." value={borrador.tipo} onChange={(e) => actualizar('tipo', e.target.value as TipoCapacitacion)} invalido={Boolean(errores.tipo)} /></Campo>
          <Campo id="capacitacion-nombre" etiqueta="Nombre del evento" requerido error={errores.nombre}><Input id="capacitacion-nombre" value={borrador.nombre} onChange={(e) => actualizar('nombre', e.target.value)} invalido={Boolean(errores.nombre)} /></Campo>
          <Campo id="capacitacion-institucion" etiqueta="Institución" requerido error={errores.institucion}><Input id="capacitacion-institucion" value={borrador.institucion} onChange={(e) => actualizar('institucion', e.target.value)} invalido={Boolean(errores.institucion)} /></Campo>
          <Campo id="capacitacion-horas" etiqueta="Número de horas" requerido error={errores.horas}><Input id="capacitacion-horas" type="number" min={1} step={1} value={borrador.horas || ''} onChange={(e) => actualizar('horas', Number(e.target.value))} invalido={Boolean(errores.horas)} /></Campo>
          <Campo id="capacitacion-inicio" etiqueta="Fecha de inicio" requerido error={errores.fecha_inicio}><DateInput id="capacitacion-inicio" max={fechaLocalISO()} value={borrador.fecha_inicio} onChange={(e) => actualizar('fecha_inicio', e.target.value)} invalido={Boolean(errores.fecha_inicio)} /></Campo>
          <Campo id="capacitacion-fin" etiqueta="Fecha de finalización" requerido error={errores.fecha_fin}><DateInput id="capacitacion-fin" min={borrador.fecha_inicio || undefined} max={fechaLocalISO()} value={borrador.fecha_fin} onChange={(e) => actualizar('fecha_fin', e.target.value)} invalido={Boolean(errores.fecha_fin)} /></Campo>
          <div className="md:col-span-2"><DocumentoUpload tipo="capacitacion" maxMb={2} valor={borrador.documento} onChange={(documento) => { if (documento) subidos.current.add(documento.documento_id); actualizar('documento', documento) }} etiqueta="Documento de respaldo (PDF, máximo 2 MB, opcional)" eliminarEnServidor={false} /></div>
          <div className="flex justify-end gap-3 md:col-span-2"><Button variante="secondary" onClick={() => setIndice(undefined)}>Cancelar</Button><Button onClick={aceptarFila}>{indice === null ? 'Agregar' : 'Actualizar'}</Button></div>
        </div>
      )}
      {guardado && <Alert tipo="success">La información se guardó correctamente.</Alert>}
      {errorGuardar && <Alert tipo="error" titulo={errorGuardar.message}>{erroresServidor.length > 0 && <ul className="list-disc pl-4">{erroresServidor.map(([campo, mensaje]) => <li key={campo}>{campo}: {mensaje}</li>)}</ul>}</Alert>}
      <div className="flex justify-end"><Button onClick={guardar} cargando={guardando} textoCargando="Guardando...">Guardar información</Button></div>
    </section>
  )
}
