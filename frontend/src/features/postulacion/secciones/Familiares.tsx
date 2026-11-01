import { useId, useRef, useState, type KeyboardEvent } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import type { z } from 'zod'
import { CampoSelect, CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Modal, { ModalConfirmacion } from '@/components/ui/Modal'
import RadioSiNo from '@/components/ui/RadioSiNo'
import TablaFilas, { type ColumnaTabla } from '@/components/ui/TablaFilas'
import { ApiError } from '@/services/errores'
import type { Familiar } from '@/services/postulacion'
import {
  MAX_FAMILIARES,
  esquemaFamiliar,
  etiquetaParentesco,
  familiarVacio,
  opcionesInstitucion,
  opcionesParentesco,
  repartirErrores,
  type FormFamiliar,
} from './reglasFamiliares'

export interface ValoresFamiliares {
  tiene_familiares: boolean
  familiares: Familiar[]
}

export interface FamiliaresProps {
  valoresIniciales: ValoresFamiliares
  // Guarda la sección. Debe rechazar la promesa si falla; un ApiError con `fields`
  // (familiares[i].campo) muestra el error en la fila correspondiente.
  onGuardar: (valores: ValoresFamiliares) => Promise<void>
}

// Fila de la tabla: el familiar y una clave local estable para React.
type Fila = Familiar & { clave: number }

const conClave = (f: Familiar, clave: number): Fila => ({ ...f, clave })
const sinClave = (f: Fila): Familiar => ({
  nombres_apellidos: f.nombres_apellidos,
  parentesco: f.parentesco,
  institucion: f.institucion,
  area: f.area,
  cargo: f.cargo,
})

// Sección "Información adicional" del formulario de postulación (EF-12): familiares
// hasta el 2.º grado de consanguinidad o 4.º de afinidad en el IESS o el BIESS.
export default function Familiares({ valoresIniciales, onGuardar }: FamiliaresProps) {
  const preguntaId = useId()
  const siguienteClave = useRef(valoresIniciales.familiares.length)
  const [tiene, setTiene] = useState(valoresIniciales.tiene_familiares)
  const [filas, setFilas] = useState<Fila[]>(() =>
    valoresIniciales.familiares.map((f, i) => conClave(f, i)),
  )
  // indice null = fila nueva
  const [editando, setEditando] = useState<{ indice: number | null } | null>(null)
  const [porEliminar, setPorEliminar] = useState<number | null>(null)
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [error, setError] = useState<ApiError | null>(null)
  const [errorLocal, setErrorLocal] = useState<string | null>(null)

  const errores = repartirErrores(error?.fields)
  const errorGeneral = errorLocal ?? errores.general

  // Cualquier cambio deja obsoletos el aviso de éxito y los errores del último guardado
  // (los índices de las filas ya no coinciden).
  const modificar = () => {
    setGuardado(false)
    setError(null)
    setErrorLocal(null)
  }

  const cambiarRespuesta = (valor: boolean) => {
    modificar()
    setTiene(valor)
  }

  const aceptarFila = (f: Familiar) => {
    modificar()
    const indice = editando?.indice ?? null
    if (indice === null) {
      setFilas((actual) => [...actual, conClave(f, siguienteClave.current++)])
    } else {
      setFilas((actual) => actual.map((fila, i) => (i === indice ? conClave(f, fila.clave) : fila)))
    }
    setEditando(null)
  }

  const eliminarFila = () => {
    modificar()
    setFilas((actual) => actual.filter((_, i) => i !== porEliminar))
    setPorEliminar(null)
  }

  const guardar = async () => {
    modificar()
    if (filas.length === 0 && tiene) {
      setErrorLocal('Agregue al menos un familiar')
      return
    }
    setGuardando(true)
    try {
      await onGuardar({ tiene_familiares: tiene, familiares: tiene ? filas.map(sinClave) : [] })
      // Con "No" el servidor borra las filas: se descartan también las que quedaron ocultas.
      if (!tiene) setFilas([])
      setGuardado(true)
    } catch (e) {
      setError(
        e instanceof ApiError
          ? e
          : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la información.'),
      )
    } finally {
      setGuardando(false)
    }
  }

  const columnas: ColumnaTabla<Fila>[] = [
    {
      clave: 'nombres_apellidos',
      titulo: 'Nombres y apellidos',
      render: (fila) => {
        const mensajes = errores.filas[filas.indexOf(fila)]
        return (
          <div className="flex flex-col gap-1">
            <span>{fila.nombres_apellidos}</span>
            {mensajes?.map((m) => (
              <span key={m} role="alert" className="text-label-sm text-error">
                {m}
              </span>
            ))}
          </div>
        )
      },
    },
    { clave: 'parentesco', titulo: 'Parentesco', render: (f) => etiquetaParentesco(f.parentesco) },
    { clave: 'institucion', titulo: 'Institución' },
    { clave: 'area', titulo: 'Área' },
    { clave: 'cargo', titulo: 'Cargo' },
  ]

  const filaEditada = editando?.indice == null ? null : filas[editando.indice]

  return (
    <Card
      titulo="Información adicional"
      descripcion="Indique si tiene familiares trabajando en el IESS o el BIESS."
    >
      <div className="flex flex-col gap-space-md">
        {error && (
          <Alert tipo="error" titulo={error.message}>
            {errores.otros.length > 0 && (
              <ul className="list-disc pl-4">
                {errores.otros.map((m) => (
                  <li key={m}>{m}</li>
                ))}
              </ul>
            )}
          </Alert>
        )}
        {guardado && <Alert tipo="success">La información de familiares se guardó.</Alert>}

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p id={preguntaId} className="text-body-bold text-on-surface">
            ¿Tiene familiares hasta el 2.º grado de consanguinidad o 4.º de afinidad trabajando en
            el IESS o BIESS?
          </p>
          <RadioSiNo
            valor={tiene}
            onChange={cambiarRespuesta}
            deshabilitado={guardando}
            aria-labelledby={preguntaId}
          />
        </div>
        {errorGeneral && (
          <p role="alert" className="text-label-sm text-error">
            {errorGeneral}
          </p>
        )}

        {tiene && (
          <TablaFilas
            titulo="Familiares"
            columnas={columnas}
            filas={filas}
            obtenerClave={(f) => f.clave}
            onAgregar={
              filas.length < MAX_FAMILIARES ? () => setEditando({ indice: null }) : undefined
            }
            onEditar={(f) => setEditando({ indice: filas.indexOf(f) })}
            onEliminar={(f) => setPorEliminar(filas.indexOf(f))}
            textoAgregar="Agregar familiar"
            textoVacio="Aún no agregó familiares."
            deshabilitado={guardando}
          />
        )}

        <div>
          <Button onClick={guardar} cargando={guardando} textoCargando="Guardando...">
            Guardar información
          </Button>
        </div>
      </div>

      {editando && (
        <ModalFamiliar
          titulo={filaEditada ? 'Editar familiar' : 'Agregar familiar'}
          inicial={filaEditada ? sinClave(filaEditada) : familiarVacio}
          onAceptar={aceptarFila}
          onCerrar={() => setEditando(null)}
        />
      )}
      <ModalConfirmacion
        abierto={porEliminar !== null}
        titulo="Eliminar familiar"
        mensaje={`¿Eliminar a ${porEliminar === null ? '' : (filas[porEliminar]?.nombres_apellidos ?? '')} de la lista?`}
        textoConfirmar="Eliminar"
        peligroso
        onConfirmar={eliminarFila}
        onCancelar={() => setPorEliminar(null)}
      />
    </Card>
  )
}

interface ModalFamiliarProps {
  titulo: string
  inicial: FormFamiliar
  onAceptar: (familiar: Familiar) => void
  onCerrar: () => void
}

// Formulario de una fila. Se monta al abrirlo, así cada apertura parte de sus valores.
// No usa <form>: la sección puede ir dentro del formulario de la página y los <form>
// no se anidan; Enter en un campo acepta la fila.
function ModalFamiliar({ titulo, inicial, onAceptar, onCerrar }: ModalFamiliarProps) {
  const form = useForm<FormFamiliar, unknown, z.output<typeof esquemaFamiliar>>({
    resolver: zodResolver(esquemaFamiliar),
    defaultValues: inicial,
  })
  const aceptar = form.handleSubmit(onAceptar)

  const alPulsar = (e: KeyboardEvent<HTMLDivElement>) => {
    if (e.key === 'Enter' && (e.target as HTMLElement).tagName === 'INPUT') {
      e.preventDefault()
      void aceptar()
    }
  }

  return (
    <Modal
      abierto
      titulo={titulo}
      onCerrar={onCerrar}
      pie={
        <>
          <Button variante="secondary" onClick={onCerrar}>
            Cancelar
          </Button>
          <Button onClick={() => void aceptar()}>Aceptar</Button>
        </>
      }
    >
      <FormProvider {...form}>
        <div onKeyDown={alPulsar} className="flex flex-col gap-space-md">
          <CampoTexto
            name="nombres_apellidos"
            etiqueta="Nombres y apellidos"
            requerido
            autoComplete="off"
          />
          <CampoSelect
            name="parentesco"
            etiqueta="Parentesco"
            requerido
            opciones={opcionesParentesco}
          />
          <CampoSelect
            name="institucion"
            etiqueta="Institución"
            requerido
            opciones={opcionesInstitucion}
          />
          <CampoTexto name="area" etiqueta="Área" requerido autoComplete="off" />
          <CampoTexto name="cargo" etiqueta="Cargo" requerido autoComplete="off" />
        </div>
      </FormProvider>
    </Modal>
  )
}
