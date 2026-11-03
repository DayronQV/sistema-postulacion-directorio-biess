import type { InternalAxiosRequestConfig } from 'axios'
import { obtenerDashboard } from './adminDashboard'
import { http } from './http'

afterEach(() => {
  delete http.defaults.adapter
})

it('consulta GET /admin/dashboard', async () => {
  const peticiones: InternalAxiosRequestConfig[] = []
  const respuesta = {
    por_cargo: { afiliados: 3, jubilados: 2, empleador: 1 },
    por_estado: { borrador: 1, enviada: 4, cerrada: 1 },
    total: 6,
  }
  http.defaults.adapter = (config) => {
    peticiones.push(config)
    return Promise.resolve({ status: 200, statusText: '', headers: {}, config, data: respuesta })
  }

  await expect(obtenerDashboard(new AbortController().signal)).resolves.toEqual(respuesta)
  expect(peticiones[0]).toMatchObject({ method: 'get', url: '/admin/dashboard' })
})
