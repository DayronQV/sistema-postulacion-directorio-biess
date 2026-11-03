import type { InternalAxiosRequestConfig } from 'axios'
import { http } from './http'
import {
  actualizarPregunta,
  cargarPreguntasCsv,
  crearPregunta,
  descargarReporte,
  eliminarPregunta,
  guardarConfig,
  listarPreguntas,
  listarResultados,
  obtenerConfig,
} from './adminEvaluacion'

// Adaptador falso: guarda la petición y responde con `data`, sin red.
function capturar(data: unknown, status = 200, headers: Record<string, string> = {}) {
  const peticiones: InternalAxiosRequestConfig[] = []
  http.defaults.adapter = (config) => {
    peticiones.push(config)
    return Promise.resolve({ status, statusText: '', headers, config, data })
  }
  return peticiones
}

afterEach(() => {
  delete http.defaults.adapter
})

const pregunta = { enunciado: '¿Capital?', opciones: ['Quito', 'Cuenca'], respuesta_correcta: 0 }

describe('servicio de evaluación (administración)', () => {
  it('listarPreguntas envía la búsqueda y la paginación', async () => {
    const peticiones = capturar({ items: [], total: 0, pagina: 2, por_pagina: 20 })
    await listarPreguntas(new AbortController().signal, { buscar: 'capital', pagina: 2 })

    expect(peticiones[0]).toMatchObject({
      method: 'get',
      url: '/admin/evaluacion/preguntas',
      params: { buscar: 'capital', pagina: 2, por_pagina: 20 },
    })
  })

  it('listarPreguntas omite la búsqueda vacía', async () => {
    const peticiones = capturar({ items: [], total: 0, pagina: 1, por_pagina: 20 })
    await listarPreguntas(new AbortController().signal, { buscar: '' })

    expect(peticiones[0].params.buscar).toBeUndefined()
  })

  it('crearPregunta envía POST con la pregunta', async () => {
    const peticiones = capturar({ id: 7, ...pregunta }, 201)
    const creada = await crearPregunta(new AbortController().signal, pregunta)

    expect(creada.id).toBe(7)
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/admin/evaluacion/preguntas' })
    expect(JSON.parse(peticiones[0].data)).toEqual(pregunta)
  })

  it('actualizarPregunta envía PUT a la pregunta', async () => {
    const peticiones = capturar({ id: 7, ...pregunta })
    await actualizarPregunta(new AbortController().signal, 7, pregunta)

    expect(peticiones[0]).toMatchObject({ method: 'put', url: '/admin/evaluacion/preguntas/7' })
  })

  it('eliminarPregunta envía DELETE', async () => {
    const peticiones = capturar(null, 204)
    await eliminarPregunta(new AbortController().signal, 7)

    expect(peticiones[0]).toMatchObject({ method: 'delete', url: '/admin/evaluacion/preguntas/7' })
  })

  it('cargarPreguntasCsv envía el archivo como multipart en el campo `archivo`', async () => {
    const peticiones = capturar({ creadas: 3 }, 201)
    const archivo = new File(['enunciado,opciones,respuesta_correcta\n'], 'preguntas.csv', {
      type: 'text/csv',
    })
    const { creadas } = await cargarPreguntasCsv(new AbortController().signal, archivo)

    expect(creadas).toBe(3)
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/admin/evaluacion/preguntas' })
    expect(peticiones[0].data).toBeInstanceOf(FormData)
    expect((peticiones[0].data as FormData).get('archivo')).toBeInstanceOf(File)
  })

  it('obtenerConfig y guardarConfig usan /admin/evaluacion/config', async () => {
    const config = { duracion_minutos: 30, num_preguntas: 10, total_banco: 40, configurada: true }
    const peticiones = capturar(config)
    await obtenerConfig(new AbortController().signal)
    await guardarConfig(new AbortController().signal, { duracion_minutos: 45, num_preguntas: 12 })

    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/admin/evaluacion/config' })
    expect(peticiones[1]).toMatchObject({ method: 'put', url: '/admin/evaluacion/config' })
    expect(JSON.parse(peticiones[1].data)).toEqual({ duracion_minutos: 45, num_preguntas: 12 })
  })

  it('listarResultados devuelve la lista de `items`', async () => {
    capturar({ items: [{ intento_id: 'i-1' }] })
    expect(await listarResultados(new AbortController().signal)).toEqual([{ intento_id: 'i-1' }])
  })

  it('descargarReporte pide el PDF y toma el nombre de Content-Disposition', async () => {
    const peticiones = capturar(new Blob(['%PDF']), 200, {
      'content-disposition': 'attachment; filename="reporte-BIESS-1.pdf"',
    })
    const archivo = await descargarReporte(new AbortController().signal, 'i 1')

    expect(peticiones[0]).toMatchObject({
      url: '/admin/evaluacion/resultados/i%201/reporte',
      responseType: 'blob',
    })
    expect(archivo.nombre).toBe('reporte-BIESS-1.pdf')
  })

  it('descargarReporte usa un nombre por defecto si el servidor no manda uno', async () => {
    capturar(new Blob(['%PDF']))
    expect((await descargarReporte(new AbortController().signal, 'i-1')).nombre).toBe(
      'reporte-evaluacion.pdf',
    )
  })
})
