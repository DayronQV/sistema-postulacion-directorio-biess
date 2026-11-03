import type { InternalAxiosRequestConfig } from 'axios'
import { actualizarFase, listarFases, type Fase } from './adminFases'
import { http } from './http'

afterEach(() => {
  delete http.defaults.adapter
})

const fase: Fase = {
  id: 1,
  nombre: 'Postulación',
  orden: 1,
  publicada: true,
  habilita_formulario: true,
  fecha_inicio: '2026-11-01T05:00:00Z',
  fecha_fin: null,
}

function capturar(data: unknown) {
  const peticiones: InternalAxiosRequestConfig[] = []
  http.defaults.adapter = (config) => {
    peticiones.push(config)
    return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data })
  }
  return peticiones
}

describe('servicio de fases del panel', () => {
  it('listarFases envía GET /admin/fases', async () => {
    const peticiones = capturar([fase])
    expect(await listarFases(new AbortController().signal)).toEqual([fase])
    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/admin/fases' })
  })

  it('actualizarFase envía PUT /admin/fases/{id} con los cambios', async () => {
    const peticiones = capturar(fase)
    const cambios = {
      publicada: true,
      habilita_formulario: true,
      fecha_inicio: fase.fecha_inicio,
      fecha_fin: null,
    }
    expect(await actualizarFase(new AbortController().signal, 1, cambios)).toEqual(fase)
    expect(peticiones[0]).toMatchObject({ method: 'put', url: '/admin/fases/1' })
    expect(JSON.parse(peticiones[0].data)).toEqual(cambios)
  })
})
