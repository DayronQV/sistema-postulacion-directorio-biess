import { useEffect, useState } from 'react'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Icon from '@/components/ui/Icon'
import Loader from '@/components/ui/Loader'
import { ModalConfirmacion } from '@/components/ui/Modal'
import Select from '@/components/ui/Select'
import { cx } from '@/lib/cx'
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
      <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h1 className="text-headline-lg text-primary">Formulario de postulación</h1>
          <Badge tono={postulacion.supero_merito ? 'primary' : 'success'}>
            {postulacion.estado_visible ?? 'Enviada'}
          </Badge>
        </div>
        {postulacion.supero_merito && (
          <Alert tipo="success" titulo="Superó la Fase de Méritos">
            Su postulación pasó a Evaluación Técnica. Se le comunicará cómo continuar con la
            evaluación teórica.
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
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <header className="flex flex-col gap-space-md">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex flex-col gap-1">
            <h1 className="text-headline-lg text-primary">Formulario de postulación</h1>
            {postulacion.codigo && (
              <p className="text-body-md text-on-surface-variant">
                Código único:{' '}
                <span className="font-mono font-semibold text-on-surface">
                  {postulacion.codigo}
                </span>
              </p>
            )}
          </div>
          {/* Aquí siempre es borrador: la enviada y la cerrada tienen su propia pantalla. */}
          <Badge tono="warning">Borrador</Badge>
        </div>
        <div className="flex flex-col gap-1.5">
          <div className="flex justify-between text-label-md text-on-surface-variant">
            <span>
              {completas} de {SECCIONES.length} secciones completas
            </span>
            <span>{porcentaje}%</span>
          </div>
          <div
            role="progressbar"
            aria-label="Avance del formulario"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={porcentaje}
            className="h-2 overflow-hidden rounded-full bg-surface-container-high"
          >
            <div
              className="h-full rounded-full bg-secondary-container transition-all"
              style={{ width: `${porcentaje}%` }}
            />
          </div>
        </div>
      </header>

      {postulacion.periodo_abierto === false && (
        <Alert tipo="warning" titulo="El periodo de postulación no está habilitado">
          Puede revisar su postulación, pero no guardar cambios ni enviarla hasta que el BIESS
          habilite el periodo.
        </Alert>
      )}

      <div className="grid grid-cols-1 gap-space-lg lg:grid-cols-[17rem_1fr]">
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

        {/* Escritorio: menú lateral */}
        <nav
          aria-label="Secciones del formulario"
          className="hidden h-fit flex-col gap-1 rounded-xl bg-surface-container-lowest p-3 shadow-sm lg:flex"
        >
          {SECCIONES.map((s, i) => {
            const completa = s.estaCompleta(postulacion)
            const esActual = s.id === actualId
            return (
              <button
                key={s.id}
                type="button"
                aria-current={esActual ? 'step' : undefined}
                onClick={() => irA(s.id)}
                className={cx(
                  'flex items-center gap-3 rounded-lg px-3 py-2.5 text-left text-body-md focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
                  esActual
                    ? 'bg-primary-container text-on-primary'
                    : 'text-on-surface hover:bg-surface-container',
                )}
              >
                <Icon nombre={s.icono} className="text-[20px]" />
                <span className="flex-1">
                  {i + 1}. {s.titulo}
                </span>
                {seccionesConCambios.includes(s.id) && (
                  <span
                    title="Cambios sin guardar"
                    aria-label="Cambios sin guardar"
                    className="h-2 w-2 rounded-full bg-secondary-container"
                  />
                )}
                <Icon
                  nombre={completa ? 'check_circle' : 'radio_button_unchecked'}
                  relleno={completa}
                  className={cx('text-[18px]', completa && !esActual && 'text-success')}
                />
                <span className="sr-only">
                  {completa ? 'Sección completa' : 'Sección incompleta'}
                </span>
              </button>
            )
          })}
        </nav>

        {/* Las secciones manejan su propio formulario: el contenedor detecta aquí lo que el
            postulante edita (texto, listas, archivos, opciones Sí/No y casillas) para el aviso
            de cambios sin guardar. La marca se quita cuando la sección guarda (store). */}
        <div
          className="min-w-0"
          onInput={() => marcarCambios(actual.id, true)}
          onChange={() => marcarCambios(actual.id, true)}
          onClick={(e) => {
            if ((e.target as Element).closest('[role="radio"], [role="checkbox"]')) {
              marcarCambios(actual.id, true)
            }
          }}
        >
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
