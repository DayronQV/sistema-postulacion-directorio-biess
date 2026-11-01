import { useSyncExternalStore } from 'react'

export type Rol = 'postulante' | 'analista' | 'administrador'

export interface Sesion {
  token: string
  expiraEn: number // marca de tiempo en ms
  rol?: Rol
}

const CLAVE = 'biess.sesion'
const escuchas = new Set<() => void>()

function leerAlmacenada(): Sesion | null {
  try {
    const crudo = sessionStorage.getItem(CLAVE)
    if (!crudo) return null
    const sesion = JSON.parse(crudo) as Sesion
    return sesion.expiraEn > Date.now() ? sesion : null
  } catch {
    return null
  }
}

// El token vive solo en la pestaña (sessionStorage). Se guarda una referencia para que
// useSyncExternalStore reciba siempre el mismo objeto mientras la sesión no cambie.
let actual: Sesion | null = leerAlmacenada()

function emitir() {
  escuchas.forEach((f) => f())
}

// El rol se lee del payload del JWT solo para decidir qué pantallas mostrar;
// la autorización real la hace el servidor en cada petición.
function rolDelToken(token: string): Rol | undefined {
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    const rol = payload.rol ?? payload.role
    return rol === 'postulante' || rol === 'analista' || rol === 'administrador' ? rol : undefined
  } catch {
    return undefined
  }
}

export function iniciarSesion(token: string, expiresInSegundos: number) {
  actual = { token, expiraEn: Date.now() + expiresInSegundos * 1000, rol: rolDelToken(token) }
  try {
    sessionStorage.setItem(CLAVE, JSON.stringify(actual))
  } catch {
    // sin almacenamiento: la sesión dura lo que dure la pestaña en memoria
  }
  emitir()
}

export function cerrarSesion() {
  if (!actual) return
  actual = null
  try {
    sessionStorage.removeItem(CLAVE)
  } catch {
    // nada que limpiar
  }
  emitir()
}

export function obtenerSesion(): Sesion | null {
  if (actual && actual.expiraEn <= Date.now()) cerrarSesion()
  return actual
}

export function obtenerToken(): string | null {
  return obtenerSesion()?.token ?? null
}

function suscribir(escucha: () => void) {
  escuchas.add(escucha)
  return () => {
    escuchas.delete(escucha)
  }
}

export function useSesion(): Sesion | null {
  return useSyncExternalStore(suscribir, obtenerSesion)
}
