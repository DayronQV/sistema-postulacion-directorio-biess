import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Icon from '@/components/ui/Icon'
import Loader from '@/components/ui/Loader'
import { ModalConfirmacion } from '@/components/ui/Modal'
import Select from '@/components/ui/Select'
import { cx } from '@/lib/cx'
import { RUTA_EVALUACION } from '@/routes/rutas'
import EncabezadoFormulario from '@/features/postulacion/EncabezadoFormulario'
import EncabezadoSeccion from '@/features/postulacion/secciones/EncabezadoSeccion'
import { motivoDeBloqueo } from '@/features/postulacion/informacionPersonal'
import { SECCIONES } from '@/features/postulacion/secciones'
import PostulacionBloqueada from '@/features/postulacion/secciones/PostulacionBloqueada'
import {
  cargarPostulacion,
  guardarSeccion,
  marcarCambios,
  refrescarPostulacion,
  useFormulario,
} from '@/features/postulacion/store'

// Cuerpo de la tarjeta de cada sección. Las secciones se dibujan como tarjeta propia (Card) o
// como <section> con su título: aquí se aplanan dentro de la tarjeta del contenedor y su
// título queda solo para lectores de pantalla, porque la franja azul ya lo muestra.
const CUERPO_SECCION = cx(
  'p-space-lg',
  '[&>section]:rounded-none [&>section]:bg-transparent [&>section]:p-0 [&>section]:shadow-none',
  '[&>section>header_h2]:sr-only [&>section>div:first-child>h2]:sr-only [&>form>h2:first-child]:sr-only',
)

// Contenedor del formulario: menú de secciones, avance, código único y estado. El contenido
// de cada sección vive en features/postulacion/secciones.
export default function FormularioPostulacion() {
  const { postulacion, cargando, errorCarga, guardando, seccionesConCambios } = useFormulario()
  const [actualId, setActualId] = useState(SECCIONES[0].id)
  const [destinoPendiente, setDestinoPendiente] = useState<string | null>(null)

  useEffect(() => {
    void cargarPostulacion()
  }, [])

  // Avisa al cerrar o recargar la pestaña con cambios sin guardar.
  const hayCambios = seccionesConCambios.length > 0
  useEffect(() => {
    if (!hayCambios) return
    const aviso = (e: BeforeUnloadEvent) => e.preventDefault()
    window.addEventListener('beforeunload', aviso)
    return () => window.removeEventListener('beforeunload', aviso)
  }, [hayCambios])

  function irA(id: string) {
    if (id === actualId) return
    if (seccionesConCambios.includes(actualId)) {
      setDestinoPendiente(id)
      return
    }
    void refrescarPostulacion()
    setActualId(id)
  }

  function descartarYContinuar() {
    marcarCambios(actualId, false)
    void refrescarPostulacion()
    if (destinoPendiente) setActualId(destinoPendiente)
    setDestinoPendiente(null)
  }

  if (!postulacion) {
    return (
      <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col justify-center gap-space-md p-margin-mobile sm:p-margin">
        {cargando && <Loader texto="Cargando su postulación..." />}
        {errorCarga?.code === 'FORBIDDEN' && (
          // El periodo de postulación está cerrado y aún no tiene postulación: reintentar no
          // sirve hasta que el BIESS lo habilite.
          <Alert tipo="warning" titulo="El periodo de postulación no está habilitado">
            Todavía no puede iniciar su postulación. Vuelva a ingresar cuando el BIESS habilite el
            periodo.
          </Alert>
        )}
        {errorCarga && errorCarga.code !== 'FORBIDDEN' && (
          <>
            <Alert tipo="error" titulo="No se pudo cargar la postulación">
              {errorCarga.message}
            </Alert>
            <div>
              <Button variante="secondary" icono="refresh" onClick={() => void cargarPostulacion()}>
                Reintentar
              </Button>
            </div>
          </>
        )}
      </main>
    )
  }

  // Una postulación cerrada (por ejemplo, por nacionalidad) no muestra el resto del formulario.
  const bloqueo = motivoDeBloqueo(postulacion)
  if (bloqueo) {
    return (
      <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col justify-center gap-space-md p-margin-mobile sm:p-margin">
        <PostulacionBloqueada motivo={bloqueo} codigo={postulacion.codigo} />
      </main>
    )
  }

  // Una postulación enviada ya no se edita (PUT /postulacion responde 409): en lugar de las
  // secciones editables se muestra solo la confirmación del Cierre, con el enlace para imprimir.
  if (postulacion.estado === 'enviada') {
    const Confirmacion = (SECCIONES.find((s) => s.id === 'cierre') ?? SECCIONES[0]).Componente
    return (
      <main className="mx-auto flex w-full max-w-[1400px] flex-1 flex-col gap-space-md p-margin-mobile sm:p-margin">
        <EncabezadoFormulario
          codigo={postulacion.codigo}
          estado={
            <Badge tono={postulacion.supero_merito ? 'primary' : 'success'}>
              {postulacion.estado_visible ?? 'Enviada'}
            </Badge>
          }
        />
        {postulacion.supero_merito && (
          <Alert tipo="success" titulo="Superó la Fase de Méritos">
            Su postulación pasó a Evaluación Técnica.{' '}
            <Link to={RUTA_EVALUACION} className="font-semibold underline">
              Ir a la evaluación
            </Link>
          </Alert>
        )}
        <Confirmacion
          titulo="Cierre"
          datos={postulacion}
          guardando={false}
          onGuardar={async () => false}
          onCambios={() => undefined}
        />
      </main>
    )
  }

  const completas = SECCIONES.filter((s) => s.estaCompleta(postulacion)).length
  const porcentaje = Math.round((completas / SECCIONES.length) * 100)
  const actual = SECCIONES.find((s) => s.id === actualId) ?? SECCIONES[0]
  const Seccion = actual.Componente

  return (
    <main className="mx-auto flex w-full max-w-[1400px] flex-1 flex-col gap-space-md p-margin-mobile sm:p-margin">
      <EncabezadoFormulario
        codigo={postulacion.codigo}
        // Aquí siempre es borrador: la enviada y la cerrada tienen su propia pantalla.
        estado={<Badge tono="warning">En proceso</Badge>}
      />

      {/* Avance */}
      <div className="flex flex-wrap items-center gap-space-md rounded-2xl bg-surface-container-lowest px-space-lg py-3 shadow-[0_2px_14px_rgba(0,38,77,0.06)]">
        <span className="flex items-center gap-3 text-body-md text-on-surface">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-on-primary">
            <Icon nombre="check" className="text-[18px]" />
          </span>
          <span>
            Secciones completadas:{' '}
            <strong>
              {completas} de {SECCIONES.length}
            </strong>
          </span>
        </span>
        <div
          role="progressbar"
          aria-label="Avance del formulario"
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={porcentaje}
          className="h-2.5 min-w-[8rem] flex-1 overflow-hidden rounded-full bg-[#dfe6f1]"
        >
          <div
            className="h-full rounded-full bg-primary transition-all"
            style={{ width: `${porcentaje}%` }}
          />
        </div>
        <span className="rounded-full bg-[#eef3fa] px-4 py-1.5 text-label-md font-bold text-primary">
          {porcentaje}%
        </span>
      </div>

      {postulacion.periodo_abierto === false && (
        <Alert tipo="warning" titulo="El periodo de postulación no está habilitado">
          Puede revisar su postulación, pero no guardar cambios ni enviarla hasta que el BIESS
          habilite el periodo.
        </Alert>
      )}

      <div className="grid grid-cols-1 gap-space-md lg:grid-cols-[18rem_1fr]">
        {/* Móvil: menú desplegable */}
        <div className="lg:hidden">
          <label htmlFor="menu-secciones" className="mb-1.5 block text-label-md text-on-surface">
            Sección
          </label>
          <Select
            id="menu-secciones"
            value={actualId}
            onChange={(e) => irA(e.target.value)}
            opciones={SECCIONES.map((s, i) => ({
              valor: s.id,
              etiqueta: `${i + 1}. ${s.titulo}${s.estaCompleta(postulacion) ? ' ✓' : ''}`,
            }))}
          />
        </div>

        {/* Escritorio: menú lateral como línea de tiempo */}
        <nav
          aria-label="Secciones del formulario"
          className="hidden h-fit flex-col gap-space-md rounded-3xl bg-surface-container-lowest p-4 shadow-sm lg:flex"
        >
          <ol className="relative flex flex-col gap-1">
            <span
              aria-hidden="true"
              className="absolute bottom-5 left-[19px] top-5 w-0.5 bg-outline-variant"
            />
            {SECCIONES.map((s, i) => {
              const completa = s.estaCompleta(postulacion)
              const esActual = s.id === actualId
              return (
                <li key={s.id} className="relative flex items-center gap-2">
                  <span
                    aria-hidden="true"
                    className={cx(
                      'relative z-10 flex h-[22px] w-[22px] shrink-0 items-center justify-center rounded-full border-2 bg-surface-container-lowest ml-[9px]',
                      completa
                        ? 'border-primary bg-primary text-on-primary'
                        : esActual
                          ? 'border-primary'
                          : 'border-outline-variant',
                    )}
                  >
                    {completa && <Icon nombre="check" className="text-[14px] font-bold" />}
                    {!completa && esActual && <span className="h-2 w-2 rounded-full bg-primary" />}
                  </span>
                  <button
                    type="button"
                    aria-current={esActual ? 'step' : undefined}
                    onClick={() => irA(s.id)}
                    className={cx(
                      'flex flex-1 items-center gap-3 rounded-xl px-3 py-2.5 text-left text-body-md focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
                      esActual
                        ? 'bg-primary text-on-primary shadow-md'
                        : 'text-on-surface hover:bg-surface-container',
                    )}
                  >
                    <Icon nombre={s.icono} relleno className="text-[22px]" />
                    <span className={cx('flex-1 leading-tight', esActual && 'font-semibold')}>
                      {i + 1}. {s.titulo}
                    </span>
                    {seccionesConCambios.includes(s.id) && (
                      <span
                        title="Cambios sin guardar"
                        aria-label="Cambios sin guardar"
                        className="h-2 w-2 rounded-full bg-[#fcd116]"
                      />
                    )}
                    {esActual ? (
                      <Icon nombre="chevron_right" className="text-[22px]" />
                    ) : (
                      completa && (
                        <Icon nombre="check_circle" relleno className="text-[20px] text-success" />
                      )
                    )}
                    <span className="sr-only">
                      {completa ? 'Sección completa' : 'Sección incompleta'}
                    </span>
                  </button>
                </li>
              )
            })}
          </ol>
          <p className="flex items-start gap-2 rounded-xl bg-primary-fixed/60 p-3 text-label-md text-primary">
            <Icon nombre="info" relleno className="text-[20px]" />
            Puedes completar las secciones en el orden que prefieras.
          </p>
        </nav>

        {/* Las secciones manejan su propio formulario: el contenedor detecta aquí lo que el
            postulante edita (texto, listas, archivos, opciones Sí/No y casillas) para el aviso
            de cambios sin guardar. La marca se quita cuando la sección guarda (store).
            No se escucha `onInput`: en una lista el navegador lanza `input` antes que `change`,
            y marcar con el primero redibuja la página y React devuelve la lista a su valor
            anterior, así que la primera elección se perdía. `onChange` de React ya cubre texto,
            listas, fechas y archivos. */}
        <div
          className="min-w-0 overflow-hidden rounded-3xl bg-surface-container-lowest shadow-[0_2px_14px_rgba(0,38,77,0.06)]"
          onChange={() => marcarCambios(actual.id, true)}
          onClick={(e) => {
            if ((e.target as Element).closest('[role="radio"], [role="checkbox"]')) {
              marcarCambios(actual.id, true)
            }
          }}
        >
          <EncabezadoSeccion titulo={actual.encabezado ?? actual.titulo} icono={actual.icono} />
          <div className={CUERPO_SECCION}>
            <Seccion
              key={actual.id}
              titulo={actual.titulo}
              datos={postulacion}
              guardando={guardando}
              onGuardar={(cambios) => guardarSeccion(actual.id, cambios)}
              onCambios={(c) => marcarCambios(actual.id, c)}
              incompletas={SECCIONES.filter(
                (s) => s.id !== 'cierre' && !s.estaCompleta(postulacion),
              ).map(({ id, titulo }) => ({ id, titulo }))}
              onIrA={irA}
            />
          </div>
        </div>
      </div>

      <ModalConfirmacion
        abierto={destinoPendiente !== null}
        titulo="Cambios sin guardar"
        mensaje="Tiene cambios sin guardar en esta sección. Si continúa, se perderán."
        textoConfirmar="Descartar y continuar"
        peligroso
        onConfirmar={descartarYContinuar}
        onCancelar={() => setDestinoPendiente(null)}
      />
    </main>
  )
}
