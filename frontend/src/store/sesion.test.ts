import { vi } from 'vitest'
import { cerrarSesion, iniciarSesion, obtenerSesion, obtenerToken } from './sesion'

function jwt(payload: object) {
  return `x.${btoa(JSON.stringify(payload))}.y`
}

afterEach(() => {
  cerrarSesion()
  vi.useRealTimers()
})

describe('sesión', () => {
  it('guarda y borra el token', () => {
    iniciarSesion('abc', 600)
    expect(obtenerToken()).toBe('abc')
    cerrarSesion()
    expect(obtenerToken()).toBeNull()
    expect(sessionStorage.getItem('biess.sesion')).toBeNull()
  })

  it('lee el rol del payload del JWT', () => {
    iniciarSesion(jwt({ rol: 'analista' }), 600)
    expect(obtenerSesion()?.rol).toBe('analista')
  })

  it('ignora un rol desconocido o un token que no es JWT', () => {
    iniciarSesion(jwt({ rol: 'otro' }), 600)
    expect(obtenerSesion()?.rol).toBeUndefined()
    iniciarSesion('no-es-jwt', 600)
    expect(obtenerSesion()?.rol).toBeUndefined()
  })

  it('cierra la sesión cuando vence el token', () => {
    vi.useFakeTimers()
    iniciarSesion('abc', 60)
    vi.advanceTimersByTime(61_000)
    expect(obtenerToken()).toBeNull()
  })
})

describe('obtenerSesion', () => {
  it('no tiene efectos: devuelve el mismo objeto en llamadas seguidas', () => {
    iniciarSesion('abc', 600)
    expect(obtenerSesion()).toBe(obtenerSesion())
  })
})
