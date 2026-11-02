import { useEffect, useRef, useState } from 'react'
import Campo from '@/components/form/Campo'
import DocumentoUpload from '@/components/form/DocumentoUpload'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import DateInput from '@/components/ui/DateInput'
import Input from '@/components/ui/Input'
import Select from '@/components/ui/Select'
import TablaFilas from '@/components/ui/TablaFilas'
import Textarea from '@/components/ui/Textarea'
import { fechaLocalISO } from '@/features/auth/esquemas'
import { calcularDuracion, sumarDuraciones, textoDuracion } from '@/features/postulacion/experiencia'
import { eliminarDocumento, type DocumentoSubido } from '@/services/documentos'
import { ApiError } from '@/services/errores'
import type {
  ActividadExperiencia,
  ExperienciaProfesional,
  TipoExperiencia,
} from '@/services/postulacion'

export const MAX_DESCRIPCION = 1000

export interface ExperienciaProps {
  valor: ExperienciaProfesional[]
  onGuardar: (experiencia: ExperienciaProfesional[]) => Promise<void> | void
}

type FilaEditor = Omit<ExperienciaProfesional, 'documento_id' | 'duracion'> & {
  documento: DocumentoSubido | null
}

const actividades = [
  { valor: 'finanzas', etiqueta: 'Finanzas' },
  { valor: 'mercado_capitales', etiqueta: 'Mercado de Capitales' },
  { valor: 'administracion_empresas', etiqueta: 'Administración de Empresas' },
  { valor: 'economia', etiqueta: 'Economía' },
  { valor: 'derecho', etiqueta: 'Derecho' },
  { valor: 'afines', etiqueta: 'Materias afines' },
] as const

const vacia = (): FilaEditor => ({
  tipo: '' as TipoExperiencia,
  cargo: '',
  institucion: '',
  fecha_inicio: '',
  fecha_fin: '',
  actividades: '' as ActividadExperiencia,
  descripcion: '',
  documento: null,
})

function alEditor(fila: ExperienciaProfesional): FilaEditor {
  return {
    tipo: fila.tipo,
    cargo: fila.cargo,
    institucion: fila.institucion,
    fecha_inicio: fila.fecha_inicio,
    fecha_fin: fila.fecha_fin,
    actividades: fila.actividades,
    descripcion: fila.descripcion,
    documento: {
      documento_id: fila.documento_id,
      nombre: `respaldo-${fila.cargo || 'experiencia'}.pdf`,
      tamano_bytes: 0,
    },
  }
}

function alDominio(fila: FilaEditor): ExperienciaProfesional {
  return {
    tipo: fila.tipo,
    cargo: fila.cargo.trim(),
    institucion: fila.institucion.trim(),
    fecha_inicio: fila.fecha_inicio,
    fecha_fin: fila.fecha_fin,
    actividades: fila.actividades,
    descripcion: fila.descripcion.trim(),
    documento_id: fila.documento?.documento_id ?? '',
    duracion: calcularDuracion(fila.fecha_inicio, fila.fecha_fin),
  }
}

function descartar(documentos: string[]) {
  for (const id of documentos) {
    eliminarDocumento(new AbortController().signal, id).catch(() => undefined)
  }
}

function erroresDe(fila: FilaEditor) {
  const errores: Partial<Record<keyof FilaEditor, string>> = {}
  if (!fila.tipo) errores.tipo = 'Seleccione el tipo de experiencia'
  if (!fila.cargo.trim()) errores.cargo = 'Campo obligatorio'
  if (!fila.institucion.trim()) errores.institucion = 'Campo obligatorio'
  if (!fila.fecha_inicio) errores.fecha_inicio = 'Ingrese la fecha inicial'
  else if (fila.fecha_inicio > fechaLocalISO()) errores.fecha_inicio = 'La fecha no puede ser futura'
  if (!fila.fecha_fin) errores.fecha_fin = 'Ingrese la fecha final'
  else if (fila.fecha_fin > fechaLocalISO()) errores.fecha_fin = 'La fecha no puede ser futura'
  else if (fila.fecha_inicio && fila.fecha_fin < fila.fecha_inicio)
    errores.fecha_fin = 'La fecha final no puede ser anterior a la inicial'
  if (!fila.actividades) errores.actividades = 'Seleccione una actividad relacionada'
  if (!fila.descripcion.trim()) errores.descripcion = 'Campo obligatorio'
  else if ([...fila.descripcion].length > MAX_DESCRIPCION)
    errores.descripcion = `Máximo ${MAX_DESCRIPCION} caracteres`
  if (!fila.documento) errores.documento = 'Adjunte el documento de respaldo'
  return errores
}

export default function Experiencia({ valor, onGuardar }: ExperienciaProps) {
  const [filas, setFilas] = useState(valor)
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)
  const subidos = useRef(new Set<string>())

  useEffect(
    () => () => {
      const pendientes = [...subidos.current]
      subidos.current.clear()
      descartar(pendientes)
    },
    [],
  )

  function abrir(fila?: ExperienciaProfesional) {
    setIndice(fila ? filas.indexOf(fila) : null)
    setBorrador(fila ? alEditor(fila) : vacia())
    setErrores({})
    setGuardado(false)
  }

  function aceptarFila() {
    const encontrados = erroresDe(borrador)
    setErrores(encontrados)
    if (Object.keys(encontrados).length > 0) return
    const fila = alDominio(borrador)
    setFilas((actuales) =>
      indice === null
        ? [...actuales, fila]
        : actuales.map((actual, posicion) => (posicion === indice ? fila : actual)),
    )
    setIndice(undefined)
    setGuardado(false)
  }

  async function guardar() {
    setGuardando(true)
    setGuardado(false)
    setErrorGuardar(null)
    try {
      await onGuardar(filas)
      const usados = new Set(filas.map((fila) => fila.documento_id))
      const sobrantes = [...subidos.current].filter((id) => !usados.has(id))
      subidos.current.clear()
      descartar(sobrantes)
      setGuardado(true)
    } catch (error) {
      setErrorGuardar(
        error instanceof ApiError
          ? error
          : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la experiencia profesional.'),
      )
    } finally {
      setGuardando(false)
    }
  }

  const actualizar = <K extends keyof FilaEditor>(campo: K, valorCampo: FilaEditor[K]) =>
    setBorrador((actual) => ({ ...actual, [campo]: valorCampo }))
  const total = sumarDuraciones(filas.map((fila) => fila.duracion))
  const erroresServidor = errorGuardar?.fields
    ? Object.entries(errorGuardar.fields).filter(([campo]) => campo.startsWith('experiencia'))
    : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby="experiencia-titulo">
      <div>
        <h2 id="experiencia-titulo" className="text-headline-sm text-primary">
          Experiencia profesional
        </h2>
        <p className="text-body-md text-on-surface-variant">
          Registre sus periodos laborales y el respaldo de cada experiencia.
        </p>
      </div>

      <TablaFilas
        titulo="Experiencias registradas"
        columnas={[
          { clave: 'cargo', titulo: 'Cargo o puesto' },
          { clave: 'institucion', titulo: 'Institución o empresa' },
          { clave: 'fecha_inicio', titulo: 'Desde' },
          { clave: 'fecha_fin', titulo: 'Hasta' },
          { clave: 'duracion', titulo: 'Tiempo', render: (fila) => textoDuracion(fila.duracion) },
        ]}
        filas={filas}
        obtenerClave={(fila) => filas.indexOf(fila)}
        onAgregar={() => abrir()}
        onEditar={abrir}
        onEliminar={(fila) => {
          setFilas((actuales) => actuales.filter((actual) => actual !== fila))
          setIndice(undefined)
          setGuardado(false)
        }}
        textoAgregar="Agregar experiencia"
        deshabilitado={guardando}
      />

      <p className="rounded-xl bg-primary-fixed p-4 text-body-bold text-on-primary-fixed">
        Tiempo total: {textoDuracion(total)}
      </p>

      {indice !== undefined && (
        <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
          <Campo id="experiencia-tipo" etiqueta="Tipo" requerido error={errores.tipo}>
            <Select
              id="experiencia-tipo"
              opciones={[
                { valor: 'general', etiqueta: 'General' },
                { valor: 'direccion', etiqueta: 'Dirección' },
              ]}
              placeholder="Seleccione..."
              value={borrador.tipo}
              onChange={(e) => actualizar('tipo', e.target.value as TipoExperiencia)}
              invalido={Boolean(errores.tipo)}
            />
          </Campo>
          <Campo id="experiencia-cargo" etiqueta="Cargo o puesto" requerido error={errores.cargo}>
            <Input id="experiencia-cargo" value={borrador.cargo} onChange={(e) => actualizar('cargo', e.target.value)} invalido={Boolean(errores.cargo)} />
          </Campo>
          <Campo id="experiencia-institucion" etiqueta="Institución o empresa" requerido error={errores.institucion}>
            <Input id="experiencia-institucion" value={borrador.institucion} onChange={(e) => actualizar('institucion', e.target.value)} invalido={Boolean(errores.institucion)} />
          </Campo>
          <Campo id="experiencia-actividades" etiqueta="Actividades relacionadas a" requerido error={errores.actividades}>
            <Select id="experiencia-actividades" opciones={[...actividades]} placeholder="Seleccione..." value={borrador.actividades} onChange={(e) => actualizar('actividades', e.target.value as ActividadExperiencia)} invalido={Boolean(errores.actividades)} />
          </Campo>
          <Campo id="experiencia-inicio" etiqueta="Fecha de inicio" requerido error={errores.fecha_inicio}>
            <DateInput id="experiencia-inicio" max={fechaLocalISO()} value={borrador.fecha_inicio} onChange={(e) => actualizar('fecha_inicio', e.target.value)} invalido={Boolean(errores.fecha_inicio)} />
          </Campo>
          <Campo id="experiencia-fin" etiqueta="Fecha de finalización" requerido error={errores.fecha_fin}>
            <DateInput id="experiencia-fin" min={borrador.fecha_inicio || undefined} max={fechaLocalISO()} value={borrador.fecha_fin} onChange={(e) => actualizar('fecha_fin', e.target.value)} invalido={Boolean(errores.fecha_fin)} />
          </Campo>
          <div className="md:col-span-2">
            <Campo id="experiencia-descripcion" etiqueta="Descripción de funciones" requerido error={errores.descripcion}>
              <Textarea id="experiencia-descripcion" maxLength={MAX_DESCRIPCION} value={borrador.descripcion} onChange={(e) => actualizar('descripcion', e.target.value)} invalido={Boolean(errores.descripcion)} />
            </Campo>
            <p className="text-right text-label-sm text-on-surface-variant">{[...borrador.descripcion].length}/{MAX_DESCRIPCION}</p>
          </div>
          <div className="md:col-span-2">
            <DocumentoUpload
              tipo="experiencia"
              maxMb={2}
              valor={borrador.documento}
              onChange={(documento) => {
                if (documento) subidos.current.add(documento.documento_id)
                actualizar('documento', documento)
              }}
              etiqueta="Documento de respaldo (PDF, máximo 2 MB)"
              eliminarEnServidor={false}
            />
            {errores.documento && <p className="mt-1 text-label-sm text-error">{errores.documento}</p>}
          </div>
          <div className="flex justify-end gap-3 md:col-span-2">
            <Button variante="secondary" onClick={() => setIndice(undefined)}>Cancelar</Button>
            <Button onClick={aceptarFila}>{indice === null ? 'Agregar' : 'Actualizar'}</Button>
          </div>
        </div>
      )}

      {guardado && <Alert tipo="success">La información se guardó correctamente.</Alert>}
      {errorGuardar && (
        <Alert tipo="error" titulo={errorGuardar.message}>
          {erroresServidor.length > 0 && <ul className="list-disc pl-4">{erroresServidor.map(([campo, mensaje]) => <li key={campo}>{campo}: {mensaje}</li>)}</ul>}
        </Alert>
      )}
      <div className="flex justify-end">
        <Button onClick={guardar} cargando={guardando} textoCargando="Guardando...">Guardar información</Button>
      </div>
    </section>
  )
}
