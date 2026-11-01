import type { InternalAxiosRequestConfig } from 'axios'
import { registrarCuenta, validarIdentidad } from './auth'
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

describe('servicio de autenticación', () => {
  it('validarIdentidad envía POST /auth/identidad y devuelve el ticket', async () => {
    const peticiones = capturar({ ticket: 'abc' })
    const signal = new AbortController().signal
    const ticket = await validarIdentidad(signal, {
      cedula: '1712345675',
      codigo_dactilar: 'V1234V5678',
    })

    expect(ticket).toBe('abc')
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/auth/identidad', signal })
    expect(JSON.parse(peticiones[0].data)).toEqual({
      cedula: '1712345675',
      codigo_dactilar: 'V1234V5678',
    })
  })

  it('registrarCuenta envía POST /auth/registro y devuelve el id', async () => {
    const peticiones = capturar({ id: 'id-1' })
    const id = await registrarCuenta(new AbortController().signal, {
      ticket: 't',
      email: 'ana@correo.com',
      password: 'Clave12345',
    })

    expect(id).toBe('id-1')
    expect(peticiones[0]).toMatchObject({ method: 'post', url: '/auth/registro' })
    expect(JSON.parse(peticiones[0].data)).toEqual({
      ticket: 't',
      email: 'ana@correo.com',
      password: 'Clave12345',
    })
  })
})
