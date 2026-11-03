import { http } from './http'

// Evaluación teórica del postulante (docs/api.md, Evaluación del postulante). El servidor es
// la autoridad del tiempo: todas las respuestas traen el tiempo restante y la hora de cierre.

export interface InicioEvaluacion {
  total_preguntas: number
  inicio: string // RFC 3339, UTC
  cierre: string
  tiempo_restante_segundos: number
}

export interface PreguntaExamen {
  numero: number // desde 1
  total: number
  id: number
  enunciado: string
  opciones: string[] // en el orden que le tocó a este postulante; nunca trae la correcta
  cierre: string
  tiempo_restante_segundos: number
  opcion_elegida: number | null // posición desde 0; null si todavía no respondió
}

export interface ResultadoExamen {
  nota: number // porcentaje de aciertos sobre 100
  aciertos: number
  total: number
  finalizado_at: string
}

// POST /evaluacion/iniciar: 201 si crea el examen, 200 si ya tenía uno en curso (con el tiempo
// que le queda; `reanudado` distingue ese caso). 403 si no está habilitado o fuera del horario
// (con `fields.inicia_en` o `fields.cierra_en`), 409 si ya rindió, 422 si el banco no alcanza.
export async function iniciarEvaluacion(signal?: AbortSignal) {
  const { data, status } = await http.post<InicioEvaluacion>('/evaluacion/iniciar', undefined, {
    signal,
  })
  return { ...data, reanudado: status === 200 }
}

// GET /evaluacion/pregunta/{n}: 404 si no inició o no existe; 409 si ya terminó o se agotó el tiempo.
export async function obtenerPregunta(n: number, signal?: AbortSignal) {
  const { data } = await http.get<PreguntaExamen>(`/evaluacion/pregunta/${n}`, { signal })
  return data
}

// PUT /evaluacion/respuesta/{n}: `opcion` es la posición mostrada, desde 0. 204 sin cuerpo.
// Se puede cambiar mientras el examen siga abierto; 409 si ya terminó.
export async function guardarRespuesta(n: number, opcion: number, signal?: AbortSignal) {
  await http.put(`/evaluacion/respuesta/${n}`, { opcion }, { signal })
}

// POST /evaluacion/finalizar: califica con lo guardado y manda el reporte al correo. Es
// idempotente: si ya estaba finalizado devuelve el mismo resultado.
export async function finalizarEvaluacion(signal?: AbortSignal) {
  const { data } = await http.post<ResultadoExamen>('/evaluacion/finalizar', undefined, {
    signal,
  })
  return data
}
