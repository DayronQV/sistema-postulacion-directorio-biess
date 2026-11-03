import { useState } from 'react'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Icon from '@/components/ui/Icon'
import { ModalConfirmacion } from '@/components/ui/Modal'
import InstruccionesExamen from '@/features/evaluacion/InstruccionesExamen'
import PreguntaActual from '@/features/evaluacion/PreguntaActual'
import ResultadoExamen from '@/features/evaluacion/ResultadoExamen'
import { useBloqueosExamen } from '@/features/evaluacion/useBloqueosExamen'
import { useExamen } from '@/features/evaluacion/useExamen'
import { formatoReloj } from '@/features/evaluacion/useReloj'
import { cx } from '@/lib/cx'

// Último minuto: el reloj se marca en rojo.
const SEGUNDOS_DE_AVISO = 60

// Evaluación teórica del postulante habilitado (EF-14): instrucciones, preguntas una por una
// con el reloj fijo arriba, finalización con confirmación y resultado. Durante el examen se
// bloquean «Atrás», copiar, cortar, pegar y el menú contextual.
export default function Evaluacion() {
  const examen = useExamen()
  const [confirmando, setConfirmando] = useState(false)
  useBloqueosExamen(examen.fase === 'examen')

  const respondidas = Object.keys(examen.respondidas).length
  const sinResponder = examen.total - respondidas
  const urgente = examen.restante !== null && examen.restante <= SEGUNDOS_DE_AVISO

  return (
    <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      {examen.fase === 'inicio' && (
        <InstruccionesExamen
          iniciando={examen.iniciando}
          error={examen.errorInicio}
          onIniciar={() => void examen.iniciar()}
          onVerResultado={() => void examen.finalizar()}
          cargandoResultado={examen.finalizando}
        />
      )}

      {examen.fase === 'examen' && (
        <>
          <div
            className={cx(
              'sticky top-16 z-30 flex items-center justify-between gap-3 rounded-xl px-5 py-3 shadow-md',
              urgente ? 'bg-error text-on-error' : 'bg-primary-container text-on-primary',
            )}
          >
            <span className="flex items-center gap-2 text-body-bold">
              <Icon nombre="timer" className="text-[22px]" />
              Tiempo restante
            </span>
            <span
              role="timer"
              aria-label="Tiempo restante"
              className="font-mono text-headline-md font-bold tabular-nums"
            >
              {examen.restante !== null ? formatoReloj(examen.restante) : '--:--'}
            </span>
          </div>

          {examen.errorExamen && (
            <Alert tipo="error" titulo="No se pudo completar la acción">
              {examen.errorExamen}
            </Alert>
          )}

          <PreguntaActual
            pregunta={examen.pregunta}
            cargando={examen.cargandoPregunta}
            guardando={examen.guardando}
            elegida={examen.respondidas[examen.numero]}
            onElegir={(opcion) => void examen.responder(opcion)}
            onAnterior={() => examen.irA(examen.numero - 1)}
            onSiguiente={() => examen.irA(examen.numero + 1)}
          />

          <nav aria-label="Preguntas" className="flex flex-wrap gap-2">
            {Array.from({ length: examen.total }, (_, i) => i + 1).map((n) => {
              const hecha = examen.respondidas[n] !== undefined
              return (
                <button
                  key={n}
                  type="button"
                  aria-label={`Ir a la pregunta ${n}${hecha ? ' (respondida)' : ''}`}
                  aria-current={n === examen.numero ? 'step' : undefined}
                  onClick={() => examen.irA(n)}
                  className={cx(
                    'h-10 w-10 rounded-lg text-label-md focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary',
                    n === examen.numero
                      ? 'bg-primary-container text-on-primary'
                      : hecha
                        ? 'bg-secondary-container text-on-secondary-container'
                        : 'border border-outline-variant text-on-surface hover:bg-surface-container',
                  )}
                >
                  {n}
                </button>
              )
            })}
          </nav>

          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-body-md text-on-surface-variant">
              {respondidas} de {examen.total} respondidas
            </p>
            <Button icono="task_alt" onClick={() => setConfirmando(true)}>
              Finalizar y Calificar
            </Button>
          </div>

          <ModalConfirmacion
            abierto={confirmando}
            titulo="Finalizar y calificar"
            mensaje={
              <div className="flex flex-col gap-space-sm">
                <p>
                  Una vez finalizado no podrá cambiar sus respuestas. Se calificará con lo que ha
                  respondido.
                </p>
                {sinResponder > 0 && (
                  <p className="font-semibold">
                    Le{' '}
                    {sinResponder === 1 ? 'falta 1 pregunta' : `faltan ${sinResponder} preguntas`}{' '}
                    por responder; contará como no acertada.
                  </p>
                )}
              </div>
            }
            textoConfirmar="Finalizar"
            cargando={examen.finalizando}
            onConfirmar={() => {
              void examen.finalizar().then(() => setConfirmando(false))
            }}
            onCancelar={() => setConfirmando(false)}
          />
        </>
      )}

      {examen.fase === 'resultado' && examen.resultado && (
        <ResultadoExamen resultado={examen.resultado} />
      )}
    </main>
  )
}
