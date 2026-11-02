import { useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Checkbox from '@/components/ui/Checkbox'
import Icon from '@/components/ui/Icon'
import { ModalConfirmacion } from '@/components/ui/Modal'
import type { ApiError } from '@/services/errores'
import { faltantesDelServidor, type SeccionFaltante } from '../cierre'
import { CONSENTIMIENTO } from '../imprimible/textos'
import { enviarFormulario } from '../store'
import PieSeccion from './PieSeccion'
import type { PropsSeccion } from './tipos'

const RUTA_IMPRIMIR = '/postulacion/imprimir'

function ListaDeSecciones({
  secciones,
  onIrA,
}: {
  secciones: SeccionFaltante[]
  onIrA?: (id: string) => void
}) {
  return (
    <ul className="mt-2 flex list-disc flex-col gap-1 pl-5">
      {secciones.map((s) => (
        <li key={s.id}>
          <button
            type="button"
            onClick={() => onIrA?.(s.id)}
            className="font-semibold text-secondary underline focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
          >
            {s.titulo}
          </button>
          {s.motivo && <span className="text-on-surface-variant">: {s.motivo}</span>}
        </li>
      ))}
    </ul>
  )
}

// Cierre del formulario: consentimiento y declaración (Guardar Postulación) y envío al correo
// del postulante (Enviar Postulación). Tras el envío la postulación ya no se edita.
export default function Cierre({
  titulo,
  datos,
  guardando,
  onGuardar,
  onCambios,
  incompletas = [],
  onIrA,
}: PropsSeccion) {
  const [acepta, setAcepta] = useState(datos.consentimiento === true)
  const [falta, setFalta] = useState(false)
  const [confirmando, setConfirmando] = useState(false)
  const [errorEnvio, setErrorEnvio] = useState<ApiError | null>(null)

  const guardado = datos.consentimiento === true
  const conCambios = acepta !== guardado
  useEffect(() => {
    onCambios(conCambios)
  }, [conCambios, onCambios])

  if (datos.estado === 'enviada') {
    return (
      <Card titulo={titulo}>
        <div className="flex flex-col items-start gap-space-md">
          <Alert tipo="success" titulo="Postulación enviada">
            Su postulación <span className="font-mono font-semibold">{datos.codigo}</span> fue
            enviada y el formulario llegó a su correo. Ya no se puede modificar.
          </Alert>
          <p className="text-body-md text-on-surface-variant">
            Imprima el formulario, fírmelo y entréguelo junto con los requisitos en el lugar, fecha
            y horario establecidos por el IESS.
          </p>
          <Link
            to={RUTA_IMPRIMIR}
            className="inline-flex items-center gap-2 font-semibold text-secondary underline"
          >
            <Icon nombre="print" className="text-[20px]" />
            Ver e imprimir el formulario
          </Link>
        </div>
      </Card>
    )
  }

  async function guardar(e: FormEvent) {
    e.preventDefault()
    if (!acepta) {
      setFalta(true)
      return
    }
    setFalta(false)
    await onGuardar({ consentimiento: true })
  }

  async function enviar() {
    const error = await enviarFormulario()
    setConfirmando(false)
    setErrorEnvio(error)
  }

  const faltantesServidor = faltantesDelServidor(errorEnvio?.fields)
  const puedeEnviar = guardado && !conCambios && incompletas.length === 0

  return (
    <Card
      titulo={titulo}
      descripcion="Lea el consentimiento y la declaración, acepte y guarde; luego envíe su postulación."
    >
      <form noValidate onSubmit={guardar} className="flex flex-col gap-space-lg">
        <ol className="flex list-decimal flex-col gap-space-sm rounded-xl bg-surface-container-low p-4 pl-9 text-body-md text-on-surface">
          {CONSENTIMIENTO.map((texto) => (
            <li key={texto}>{texto}</li>
          ))}
        </ol>

        <div className="flex flex-col gap-1">
          <Checkbox
            id="acepto-consentimiento"
            checked={acepta}
            onChange={(e) => {
              setAcepta(e.target.checked)
              if (e.target.checked) setFalta(false)
            }}
            invalido={falta}
            etiqueta="He leído y acepto el consentimiento y la declaración"
          />
          {falta && (
            <p role="alert" className="text-body-sm text-error">
              Debe aceptar para guardar la postulación
            </p>
          )}
        </div>

        <PieSeccion textoGuardar="Guardar Postulación" />
      </form>

      <div className="mt-space-lg flex flex-col gap-space-md border-t border-outline-variant pt-space-lg">
        {incompletas.length > 0 && (
          <Alert tipo="warning" titulo="Aún faltan secciones por completar">
            Complete estas secciones para poder enviar:
            <ListaDeSecciones secciones={incompletas} onIrA={onIrA} />
          </Alert>
        )}
        {errorEnvio && (
          <Alert tipo="error" titulo={errorEnvio.message}>
            {faltantesServidor.length > 0 && (
              <ListaDeSecciones secciones={faltantesServidor} onIrA={onIrA} />
            )}
          </Alert>
        )}
        {incompletas.length === 0 && !guardado && (
          <Alert tipo="info">Guarde la postulación con su aceptación para poder enviarla.</Alert>
        )}
        <div>
          <Button
            icono="send"
            disabled={!puedeEnviar || guardando}
            onClick={() => setConfirmando(true)}
          >
            Enviar Postulación
          </Button>
        </div>
      </div>

      <ModalConfirmacion
        abierto={confirmando}
        titulo="Enviar la postulación"
        mensaje="Después de enviar no podrá editar su postulación. Recibirá el formulario completo en su correo."
        textoConfirmar="Enviar postulación"
        cargando={guardando}
        onConfirmar={() => void enviar()}
        onCancelar={() => setConfirmando(false)}
      />
    </Card>
  )
}
