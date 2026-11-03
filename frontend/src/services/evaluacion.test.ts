import type { InternalAxiosRequestConfig } from 'axios'
import { http } from './http'
import {
  finalizarEvaluacion,
  guardarRespuesta,
  iniciarEvaluacion,
  obtenerPregunta,
} from './evaluacion'

// Adaptador falso: guarda la petición y responde con `data`, sin red.
function capturar(data: unknown, status = 200) {
  const peticiones: InternalAxiosRequestConfig[] = []
  http.defaults.adapter = (config) => {
    peticiones.push(config)
    return Promise.resolve({ status, statusText: '', headers: {}, config, data })
  }
  return peticiones
}

afterEach(() => {
  delete http.defaults.adapter
})

describe('servicio de evaluación (postulante)', () => {
  it('iniciarEvaluacion envía POST /evaluacion/iniciar sin cuerpo', async () => {
    const peticiones = capturar(
      { total_preguntas: 10, inicio: 'a', cierre: 'b', tiempo_restante_segundos: 1800 },
      201,
    )
    const inicio = await iniciarEvaluacion()

    expect(inicio.tiempo_restante_segundos).toBe(1800)
    expect(inicio.reanudado).toBe(false)
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/evaluacion/iniciar' })
    expect(peticiones[0].data).toBeUndefined()
  })

  it('iniciarEvaluacion marca como reanudado el examen que ya estaba en curso (200)', async () => {
    capturar({ total_preguntas: 10, inicio: 'a', cierre: 'b', tiempo_restante_segundos: 900 }, 200)
    const inicio = await iniciarEvaluacion()

    expect(inicio.reanudado).toBe(true)
  })

  it('obtenerPregunta pide GET /evaluacion/pregunta/{n}', async () => {
    const peticiones = capturar({ numero: 3, total: 10 })
    await obtenerPregunta(3)

    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/evaluacion/pregunta/3' })
  })

  it('guardarRespuesta envía PUT con la posición de la opción', async () => {
    const peticiones = capturar(null, 204)
    await guardarRespuesta(2, 1)

    expect(peticiones[0]).toMatchObject({ method: 'put', url: '/evaluacion/respuesta/2' })
    expect(JSON.parse(peticiones[0].data)).toEqual({ opcion: 1 })
  })

  it('finalizarEvaluacion envía POST y devuelve la calificación', async () => {
    const peticiones = capturar({ nota: 80, aciertos: 8, total: 10, finalizado_at: 'x' })
    const resultado = await finalizarEvaluacion()

    expect(resultado.nota).toBe(80)
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/evaluacion/finalizar' })
  })
})
