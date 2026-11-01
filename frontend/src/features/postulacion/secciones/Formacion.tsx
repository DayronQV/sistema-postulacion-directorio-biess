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
import type {
  AreaFormacion,
  FormacionAcademica,
  NivelFormacion,
  TipoInstitucion,
} from '@/services/postulacion'

// Igual que domain.MaxFormaciones en el servidor.
export const MAX_FORMACIONES = 20

export interface FormacionProps {
  valor: FormacionAcademica[]
  onGuardar: (formacion: FormacionAcademica[]) => Promise<void> | void
}

type FilaEditor = Omit<FormacionAcademica, 'documento_id'> & {
  documento: DocumentoSubido | null
}

const niveles = [
  { valor: 'tercer_nivel', etiqueta: 'Tercer Nivel de Grado' },
  { valor: 'especialidad', etiqueta: 'Cuarto Nivel - Especialidad Académica' },
  { valor: 'maestria', etiqueta: 'Cuarto Nivel - Maestría Académica' },
  { valor: 'phd', etiqueta: 'Cuarto Nivel PhD - Doctorado' },
] as const

const areas = [
  { valor: 'finanzas', etiqueta: 'Finanzas' },
  { valor: 'mercado_capitales', etiqueta: 'Mercado de Capitales' },
  { valor: 'administracion_empresas', etiqueta: 'Administración de Empresas' },
  { valor: 'economia', etiqueta: 'Economía' },
  { valor: 'derecho', etiqueta: 'Derecho' },
  { valor: 'afines', etiqueta: 'Materias afines' },
] as const

const etiquetaNivel = Object.fromEntries(niveles.map((n) => [n.valor, n.etiqueta]))

const vacia = (): FilaEditor => ({
  nivel: '' as NivelFormacion,
  institucion: '',
  tipo: '' as TipoInstitucion,
  titulo: '',
  area: '' as AreaFormacion,
  senescyt_numero: '',
  senescyt_fecha: '',
  documento: null,
})

// La API solo devuelve documento_id: el nombre es descriptivo y el tamaño 0 indica a
// FileUpload que no lo muestre.
function alEditor(fila: FormacionAcademica): FilaEditor {
  return {
    ...fila,
    documento: fila.documento_id
      ? {
          documento_id: fila.documento_id,
          nombre: `respaldo-${fila.titulo || 'formacion'}.pdf`,
          tamano_bytes: 0,
        }
      : null,
  }
}

function alDominio(fila: FilaEditor): FormacionAcademica {
  return {
    nivel: fila.nivel,
    institucion: fila.institucion.trim(),
    tipo: fila.tipo,
    titulo: fila.titulo.trim(),
    area: fila.area,
    senescyt_numero: fila.senescyt_numero.trim(),
    senescyt_fecha: fila.senescyt_fecha,
    documento_id: fila.documento?.documento_id ?? '',
  }
}

// Borra PDF que ya no usa ninguna fila. Un fallo (por ejemplo 409 si sigue en uso) se
// ignora: lo peor es que el archivo quede guardado.
function descartar(documentos: string[]) {
  for (const id of documentos) {
    eliminarDocumento(new AbortController().signal, id).catch(() => undefined)
  }
}

function erroresDe(fila: FilaEditor) {
  const errores: Partial<Record<keyof FilaEditor, string>> = {}
  if (!fila.nivel) errores.nivel = 'Seleccione el nivel de instrucción'
  if (!fila.institucion.trim()) errores.institucion = 'Campo obligatorio'
  if (!fila.tipo) errores.tipo = 'Seleccione el tipo de institución'
  if (!fila.titulo.trim()) errores.titulo = 'Campo obligatorio'
  if (!fila.area) errores.area = 'Seleccione el área de estudios'
  if (!fila.senescyt_numero.trim()) errores.senescyt_numero = 'Campo obligatorio'
  if (!fila.senescyt_fecha) errores.senescyt_fecha = 'Ingrese la fecha de registro'
  else if (fila.senescyt_fecha > fechaLocalISO())
    errores.senescyt_fecha = 'La fecha no puede ser futura'
  // El respaldo es opcional: formacion_academica.documento_id admite NULL y la
  // especificación solo lo exige en Experiencia.
  return errores
}

export default function Formacion({ valor, onGuardar }: FormacionProps) {
  const [filas, setFilas] = useState(valor)
  const [indice, setIndice] = useState<number | null>()
  const [borrador, setBorrador] = useState<FilaEditor>(vacia())
  const [errores, setErrores] = useState<ReturnType<typeof erroresDe>>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)
  // PDF subidos en esta pantalla y todavía no guardados. Si al guardar ninguna fila los
  // usa (se canceló la edición, se eliminó la fila o se reemplazó el archivo) se borran,
  // para que no sigan sumando al límite de 50 MB. Los que ya estaban guardados los limpia
  // guardarFormacion comparando con la versión anterior.
  const subidos = useRef(new Set<string>())

  useEffect(
    () => () => {
      // Al salir sin guardar, lo subido aquí no lo referencia ninguna fila guardada.
      const pendientes = [...subidos.current]
      subidos.current.clear()
      descartar(pendientes)
    },
    [],
  )

  function abrir(fila?: FormacionAcademica) {
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
    } catch (e) {
      setErrorGuardar(
        e instanceof ApiError
          ? e
          : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la formación académica.'),
      )
    } finally {
      setGuardando(false)
    }
  }

  const actualizar = <K extends keyof FilaEditor>(campo: K, valorCampo: FilaEditor[K]) =>
    setBorrador((actual) => ({ ...actual, [campo]: valorCampo }))

  const erroresServidor = errorGuardar?.fields
    ? Object.entries(errorGuardar.fields).filter(([campo]) => campo.startsWith('formacion'))
    : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby="formacion-titulo">
      <div>
        <h2 id="formacion-titulo" className="text-headline-sm text-primary">
          Formación académica
        </h2>
        <p className="text-body-md text-on-surface-variant">
          Registre sus títulos y el respaldo emitido o registrado por la SENESCYT.
        </p>
      </div>

      <TablaFilas
        titulo="Títulos registrados"
        columnas={[
          { clave: 'nivel', titulo: 'Nivel', render: (fila) => etiquetaNivel[fila.nivel] },
          { clave: 'titulo', titulo: 'Título obtenido' },
          { clave: 'institucion', titulo: 'Institución' },
          { clave: 'senescyt_fecha', titulo: 'Fecha SENESCYT' },
        ]}
        filas={filas}
        obtenerClave={(fila) => filas.indexOf(fila)}
        onAgregar={filas.length >= MAX_FORMACIONES ? undefined : () => abrir()}
        onEditar={abrir}
        onEliminar={(fila) => {
          setFilas((actuales) => actuales.filter((actual) => actual !== fila))
          setIndice(undefined)
          setGuardado(false)
        }}
        textoAgregar="Agregar formación"
        deshabilitado={guardando}
      />

      {indice !== undefined && (
        <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
          <Campo
            id="formacion-nivel"
            etiqueta="Nivel de instrucción"
            requerido
            error={errores.nivel}
          >
            <Select
              id="formacion-nivel"
              opciones={[...niveles]}
              placeholder="Seleccione..."
              value={borrador.nivel}
              onChange={(e) => actualizar('nivel', e.target.value as NivelFormacion)}
              invalido={Boolean(errores.nivel)}
            />
          </Campo>
          {borrador.nivel && (
            <>
              <Campo
                id="formacion-institucion"
                etiqueta="Institución educativa"
                requerido
                error={errores.institucion}
              >
                <Input
                  id="formacion-institucion"
                  value={borrador.institucion}
                  onChange={(e) => actualizar('institucion', e.target.value)}
                  invalido={Boolean(errores.institucion)}
                />
              </Campo>
              <Campo
                id="formacion-tipo"
                etiqueta="Tipo de institución"
                requerido
                error={errores.tipo}
              >
                <Select
                  id="formacion-tipo"
                  opciones={[
                    { valor: 'nacional', etiqueta: 'Nacional' },
                    { valor: 'extranjera', etiqueta: 'Extranjera' },
                  ]}
                  placeholder="Seleccione..."
                  value={borrador.tipo}
                  onChange={(e) => actualizar('tipo', e.target.value as TipoInstitucion)}
                  invalido={Boolean(errores.tipo)}
                />
              </Campo>
              <Campo
                id="formacion-titulo-obtenido"
                etiqueta="Título obtenido"
                requerido
                error={errores.titulo}
              >
                <Input
                  id="formacion-titulo-obtenido"
                  value={borrador.titulo}
                  onChange={(e) => actualizar('titulo', e.target.value)}
                  invalido={Boolean(errores.titulo)}
                />
              </Campo>
              <Campo id="formacion-area" etiqueta="Área de estudios" requerido error={errores.area}>
                <Select
                  id="formacion-area"
                  opciones={[...areas]}
                  placeholder="Seleccione..."
                  value={borrador.area}
                  onChange={(e) => actualizar('area', e.target.value as AreaFormacion)}
                  invalido={Boolean(errores.area)}
                />
              </Campo>
              <Campo
                id="formacion-senescyt-numero"
                etiqueta="Número de registro SENESCYT"
                requerido
                error={errores.senescyt_numero}
              >
                <Input
                  id="formacion-senescyt-numero"
                  value={borrador.senescyt_numero}
                  onChange={(e) => actualizar('senescyt_numero', e.target.value)}
                  invalido={Boolean(errores.senescyt_numero)}
                />
              </Campo>
              <Campo
                id="formacion-senescyt-fecha"
                etiqueta="Fecha de registro SENESCYT"
                requerido
                error={errores.senescyt_fecha}
              >
                <DateInput
                  id="formacion-senescyt-fecha"
                  max={fechaLocalISO()}
                  value={borrador.senescyt_fecha}
                  onChange={(e) => actualizar('senescyt_fecha', e.target.value)}
                  invalido={Boolean(errores.senescyt_fecha)}
                />
              </Campo>
              <div className="md:col-span-2">
                <DocumentoUpload
                  tipo="formacion"
                  maxMb={2}
                  valor={borrador.documento}
                  onChange={(documento) => {
                    if (documento) subidos.current.add(documento.documento_id)
                    actualizar('documento', documento)
                  }}
                  etiqueta="Documento de respaldo (PDF, máximo 2 MB, opcional)"
                  eliminarEnServidor={false}
                />
              </div>
            </>
          )}
          <div className="flex justify-end gap-3 md:col-span-2">
            <Button variante="secondary" onClick={() => setIndice(undefined)}>
              Cancelar
            </Button>
            <Button onClick={aceptarFila}>{indice === null ? 'Agregar' : 'Actualizar'}</Button>
          </div>
        </div>
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
      <div className="flex justify-end">
        <Button onClick={guardar} cargando={guardando} textoCargando="Guardando...">
          Guardar información
        </Button>
      </div>
    </section>
  )
}
