import { vi } from 'vitest'
import {
  cerrarSesion,
  expirarSesion,
  iniciarSesion,
  marcarPostulacionCreada,
  obtenerSesion,
  obtenerToken,
  sesionExpirada,
} from './sesion'

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

describe('motivo del cierre de sesión', () => {
  it('expirarSesion marca la sesión como expirada y cerrarSesion no', () => {
    iniciarSesion('abc', 600)
    expirarSesion()
    expect(obtenerToken()).toBeNull()
    expect(sesionExpirada()).toBe(true)

    iniciarSesion('abc', 600)
    expect(sesionExpirada()).toBe(false)
    cerrarSesion()
    expect(sesionExpirada()).toBe(false)
  })

  it('el vencimiento del token cuenta como expiración', () => {
    vi.useFakeTimers()
    iniciarSesion('abc', 60)
    vi.advanceTimersByTime(61_000)
    expect(sesionExpirada()).toBe(true)
  })

  it('una sesión guardada que venció mientras la pestaña estaba cerrada se descarta como expirada', async () => {
    sessionStorage.setItem(
      'biess.sesion',
      JSON.stringify({ token: 'viejo', expiraEn: Date.now() - 1000 }),
    )
    vi.resetModules()
    const recargado = await import('./sesion')
    expect(recargado.obtenerSesion()).toBeNull()
    expect(recargado.sesionExpirada()).toBe(true)
    expect(sessionStorage.getItem('biess.sesion')).toBeNull()
  })
})

describe('primer ingreso', () => {
  it('se guarda con la sesión y se levanta al crear la postulación', () => {
    iniciarSesion('abc', 600, { primerIngreso: true })
    expect(obtenerSesion()?.primerIngreso).toBe(true)

    marcarPostulacionCreada()
    expect(obtenerSesion()?.primerIngreso).toBe(false)
    expect(JSON.parse(sessionStorage.getItem('biess.sesion') ?? '{}').primerIngreso).toBe(false)
  })

  it('sin primer ingreso marcarPostulacionCreada no cambia la sesión', () => {
    iniciarSesion('abc', 600)
    const antes = obtenerSesion()
    marcarPostulacionCreada()
    expect(obtenerSesion()).toBe(antes)
  })
})
