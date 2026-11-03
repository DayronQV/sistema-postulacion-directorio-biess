import { useCallback, useEffect, useRef, useState } from 'react'
import { ApiError } from '@/services/errores'
import {
  finalizarEvaluacion,
  guardarRespuesta,
  iniciarEvaluacion,
  obtenerPregunta,
  type PreguntaExamen,
  type ResultadoExamen,
} from '@/services/evaluacion'
import { useReloj } from './useReloj'

export type FaseExamen = 'inicio' | 'examen' | 'resultado'

function comoApiError(e: unknown): ApiError {
  return e instanceof ApiError ? e : new ApiError(0, 'UNKNOWN', 'Ocurrió un error inesperado.')
}

// Estado y acciones del examen: iniciar, mostrar las preguntas una por una, guardar cada
// respuesta, el reloj con cierre automático y la calificación. El servidor decide cuándo
// termina el examen; aquí se reacciona a su tiempo restante y a sus 409.
export function useExamen() {
  const [fase, setFase] = useState<FaseExamen>('inicio')
  const [iniciando, setIniciando] = useState(false)
  const [errorInicio, setErrorInicio] = useState<ApiError | null>(null)

  const [total, setTotal] = useState(0)
  const [numero, setNumero] = useState(1)
  const [pregunta, setPregunta] = useState<PreguntaExamen | null>(null)
  const [cargandoPregunta, setCargandoPregunta] = useState(false)
  // Opción marcada por número de pregunta; sirve para el avance y para mostrarla al volver.
  const [respondidas, setRespondidas] = useState<Record<number, number>>({})
  const [guardando, setGuardando] = useState(false)
  const [errorExamen, setErrorExamen] = useState('')

  const [finalizando, setFinalizando] = useState(false)
  const [resultado, setResultado] = useState<ResultadoExamen | null>(null)
  const finalizada = useRef(false)

  const finalizarRef = useRef<() => Promise<void>>(async () => undefined)
  const reloj = useReloj(() => void finalizarRef.current())
  const { sincronizar, detener } = reloj
  // Respuesta que se está guardando: finalizar la espera, para que entre en la nota.
  const guardadoEnCurso = useRef<Promise<void> | null>(null)

  const finalizar = useCallback(async () => {
    if (finalizada.current) return
    setFinalizando(true)
    setErrorExamen('')
    try {
      await guardadoEnCurso.current?.catch(() => undefined)
      const calificacion = await finalizarEvaluacion()
      finalizada.current = true
      detener()
      setResultado(calificacion)
      setFase('resultado')
    } catch (e) {
      setErrorExamen(comoApiError(e).message)
    } finally {
      setFinalizando(false)
    }
  }, [detener])
  useEffect(() => {
    finalizarRef.current = finalizar
  }, [finalizar])

  const cargarPregunta = useCallback(
    async (n: number) => {
      setCargandoPregunta(true)
      setErrorExamen('')
      try {
        const datos = await obtenerPregunta(n)
        setPregunta(datos)
        setNumero(datos.numero)
        setTotal(datos.total)
        sincronizar(datos.tiempo_restante_segundos)
        const elegida = datos.opcion_elegida
        if (elegida !== null) setRespondidas((r) => ({ ...r, [datos.numero]: elegida }))
      } catch (e) {
        const error = comoApiError(e)
        // 409: el examen ya terminó o se agotó el tiempo; se califica con lo guardado.
        if (error.status === 409) await finalizar()
        else setErrorExamen(error.message)
      } finally {
        setCargandoPregunta(false)
      }
    },
    [finalizar, sincronizar],
  )

  const iniciar = useCallback(async () => {
    setIniciando(true)
    setErrorInicio(null)
    try {
      const inicio = await iniciarEvaluacion()
      setTotal(inicio.total_preguntas)
      sincronizar(inicio.tiempo_restante_segundos)
      setFase('examen')
      await cargarPregunta(1)
      // Si retoma un examen en curso (recargó la página), trae lo que ya respondió en las
      // demás preguntas, para que el avance y el aviso de "sin responder" sean correctos.
      if (inicio.reanudado) {
        const numeros = Array.from({ length: inicio.total_preguntas - 1 }, (_, i) => i + 2)
        const leidas = await Promise.allSettled(numeros.map((n) => obtenerPregunta(n)))
        const previas: Record<number, number> = {}
        leidas.forEach((r) => {
          if (r.status === 'fulfilled' && r.value.opcion_elegida !== null)
            previas[r.value.numero] = r.value.opcion_elegida
        })
        setRespondidas((r) => ({ ...previas, ...r }))
      }
    } catch (e) {
      setErrorInicio(comoApiError(e))
    } finally {
      setIniciando(false)
    }
  }, [cargarPregunta, sincronizar])

  const responder = useCallback(
    async (opcion: number) => {
      setGuardando(true)
      setErrorExamen('')
      const guardado = guardarRespuesta(numero, opcion)
      guardadoEnCurso.current = guardado
      try {
        await guardado
        setRespondidas((r) => ({ ...r, [numero]: opcion }))
      } catch (e) {
        const error = comoApiError(e)
        if (error.status === 409) await finalizar()
        else setErrorExamen(error.message)
      } finally {
        if (guardadoEnCurso.current === guardado) guardadoEnCurso.current = null
        setGuardando(false)
      }
    },
    [numero, finalizar],
  )

  const irA = useCallback(
    (n: number) => {
      if (n < 1 || n > total || n === numero || guardando || cargandoPregunta) return
      void cargarPregunta(n)
    },
    [total, numero, guardando, cargandoPregunta, cargarPregunta],
  )

  return {
    fase,
    iniciando,
    errorInicio,
    iniciar,
    total,
    numero,
    pregunta,
    cargandoPregunta,
    respondidas,
    guardando,
    errorExamen,
    responder,
    irA,
    restante: reloj.restante,
    finalizando,
    finalizar,
    resultado,
  }
}
