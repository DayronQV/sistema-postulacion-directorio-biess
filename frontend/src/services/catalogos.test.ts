import type { InternalAxiosRequestConfig } from 'axios'
import { listarCantones, listarPaises, listarProvincias } from './catalogos'
import { http } from './http'

// Adaptador falso: guarda la petición y responde con `data`, sin red.
function capturar(data: unknown) {
  const peticiones: InternalAxiosRequestConfig[] = []
  http.defaults.adapter = (config) => {
    peticiones.push(config)
    return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data })
  }
  return peticiones
}

afterEach(() => {
  delete http.defaults.adapter
})

describe('servicio de catálogos', () => {
  it('listarPaises pide GET /catalogos/paises', async () => {
    const peticiones = capturar([{ id: 52, nombre: 'Ecuador' }])
    const signal = new AbortController().signal

    expect(await listarPaises(signal)).toEqual([{ id: 52, nombre: 'Ecuador' }])
    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/catalogos/paises', signal })
  })

  it('listarProvincias pide las provincias del país', async () => {
    const peticiones = capturar([{ id: 17, nombre: 'Pichincha' }])

    expect(await listarProvincias(new AbortController().signal, 52)).toHaveLength(1)
    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/catalogos/paises/52/provincias' })
  })

  it('listarCantones pide los cantones de la provincia', async () => {
    const peticiones = capturar([{ id: 170, nombre: 'Quito' }])

    expect(await listarCantones(new AbortController().signal, 17)).toHaveLength(1)
    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/catalogos/provincias/17/cantones' })
  })
})
