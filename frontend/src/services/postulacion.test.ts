import type { InternalAxiosRequestConfig } from 'axios'
import { http } from './http'
import {
  guardarPostulacion,
  obtenerPostulacion,
  validarCargo,
  type Postulacion,
} from './postulacion'

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

describe('servicio de postulación', () => {
  it('obtenerPostulacion envía GET /postulacion y devuelve el cuerpo', async () => {
    const peticiones = capturar({ id: 'p-1', familiares: [] })
    const p = await obtenerPostulacion()

    expect(p).toEqual({ id: 'p-1', familiares: [] })
    expect(peticiones[0]).toMatchObject({ method: 'get', url: '/postulacion' })
  })

  it('guardarPostulacion envía PUT /postulacion con el formulario entero', async () => {
    const formulario = { id: 'p-1', cargo: 'afiliados', familiares: [] } as unknown as Postulacion
    const peticiones = capturar(formulario)
    const guardada = await guardarPostulacion(formulario)

    expect(guardada).toEqual(formulario)
    expect(peticiones[0]).toMatchObject({ method: 'put', url: '/postulacion' })
    expect(JSON.parse(peticiones[0].data)).toEqual(formulario)
  })

  it('validarCargo envía POST /postulacion/cargo/validar con el cargo y devuelve el resultado', async () => {
    const resultado = {
      cargo: 'jubilados',
      elegible: true,
      mensaje: 'Postulante acredita jubilación',
    }
    const peticiones = capturar(resultado)
    const devuelto = await validarCargo('jubilados')

    expect(devuelto).toEqual(resultado)
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/postulacion/cargo/validar' })
    expect(JSON.parse(peticiones[0].data)).toEqual({ cargo: 'jubilados' })
  })
})
