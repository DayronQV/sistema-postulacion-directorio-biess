import { AxiosError, type AxiosAdapter, type InternalAxiosRequestConfig } from 'axios'
import { cerrarSesion, iniciarSesion, obtenerToken } from '@/store/sesion'
import { ApiError } from './errores'
import { http } from './http'

// Adaptador falso: responde con el estado y cuerpo indicados, sin red.
function simular(status: number, data: unknown): AxiosAdapter {
  return (config: InternalAxiosRequestConfig) => {
    const respuesta = { status, data, statusText: '', headers: {}, config }
    if (status >= 400) {
      return Promise.reject(new AxiosError('fallo', 'ERR_BAD_REQUEST', config, null, respuesta))
    }
    return Promise.resolve(respuesta)
  }
}

afterEach(() => {
  cerrarSesion()
  delete http.defaults.adapter
})

describe('cliente http', () => {
  it('envía el token como Bearer cuando hay sesión', async () => {
    iniciarSesion('abc', 600)
    let auth: unknown
    http.defaults.adapter = (config) => {
      auth = config.headers.Authorization
      return simular(200, {})(config)
    }
    await http.get('/postulacion')
    expect(auth).toBe('Bearer abc')
  })

  it('no envía Authorization sin sesión', async () => {
    let auth: unknown = 'sin-llamar'
    http.defaults.adapter = (config) => {
      auth = config.headers.Authorization
      return simular(200, {})(config)
    }
    await http.post('/auth/login', {})
    expect(auth).toBeUndefined()
  })

  it('normaliza el error del contrato en un ApiError', async () => {
    http.defaults.adapter = simular(429, {
      error: { code: 'LOCKED', message: 'Cuenta bloqueada', retry_after_seconds: 540 },
    })
    const error = await http.post('/auth/login', {}).catch((e) => e)
    expect(error).toBeInstanceOf(ApiError)
    expect(error).toMatchObject({
      status: 429,
      code: 'LOCKED',
      message: 'Cuenta bloqueada',
      retryAfterSeconds: 540,
    })
  })

  it('conserva los errores por campo', async () => {
    http.defaults.adapter = simular(422, {
      error: {
        code: 'VALIDATION_ERROR',
        message: 'Datos inválidos',
        fields: { email: 'inválido' },
      },
    })
    const error = await http.post('/auth/registro', {}).catch((e) => e)
    expect(error.fields).toEqual({ email: 'inválido' })
  })

  it('cierra la sesión ante un 401 con token', async () => {
    iniciarSesion('vencido', 600)
    http.defaults.adapter = simular(401, {
      error: { code: 'UNAUTHORIZED', message: 'Sesión vencida' },
    })
    await http.get('/postulacion').catch(() => undefined)
    expect(obtenerToken()).toBeNull()
  })

  it('un 401 sin sesión (credenciales malas) no hace nada más', async () => {
    http.defaults.adapter = simular(401, { error: { code: 'UNAUTHORIZED', message: 'Incorrecto' } })
    const error = await http.post('/auth/login', {}).catch((e) => e)
    expect(error.code).toBe('UNAUTHORIZED')
  })

  it('lee el error del servidor aunque la respuesta sea un Blob (descargas)', async () => {
    const cuerpo = new Blob(
      [JSON.stringify({ error: { code: 'NOT_FOUND', message: 'Documento no encontrado' } })],
      { type: 'application/json' },
    )
    http.defaults.adapter = simular(404, cuerpo)
    const error = await http
      .get('/postulacion/documentos/x', { responseType: 'blob' })
      .catch((e) => e)
    expect(error).toMatchObject({
      status: 404,
      code: 'NOT_FOUND',
      message: 'Documento no encontrado',
    })
  })

  it('devuelve NETWORK_ERROR si no hay respuesta del servidor', async () => {
    http.defaults.adapter = (config) =>
      Promise.reject(new AxiosError('Network Error', 'ERR_NETWORK', config))
    const error = await http.get('/x').catch((e) => e)
    expect(error).toMatchObject({ status: 0, code: 'NETWORK_ERROR' })
  })
})
