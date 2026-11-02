import type { InternalAxiosRequestConfig } from 'axios'
import { marcarMerito } from './adminPostulaciones'
import { http } from './http'

afterEach(() => {
  delete http.defaults.adapter
})

describe('servicio de postulaciones del panel', () => {
  it('marcarMerito envía PUT /admin/postulaciones/{id}/merito con supero', async () => {
    const peticiones: InternalAxiosRequestConfig[] = []
    const respuesta = { id: 'p1', supero_merito: true, estado_visible: 'Evaluación Técnica' }
    http.defaults.adapter = (config) => {
      peticiones.push(config)
      return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data: respuesta })
    }

    const r = await marcarMerito(new AbortController().signal, 'p1', true)

    expect(r).toEqual(respuesta)
    expect(peticiones[0]).toMatchObject({ method: 'put', url: '/admin/postulaciones/p1/merito' })
    expect(JSON.parse(peticiones[0].data)).toEqual({ supero: true })
  })
})
