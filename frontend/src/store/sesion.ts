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
let actual: Sesion | null = null
let temporizador: ReturnType<typeof setTimeout> | undefined

// setTimeout no admite plazos mayores a ~24,8 días.
const PLAZO_MAXIMO_MS = 2 ** 31 - 1

function emitir() {
  escuchas.forEach((f) => f())
}

// Programa el cierre de la sesión cuando vence el token, para que las rutas protegidas
// redirijan al login aunque el usuario no haga nada.
function programarVencimiento(sesion: Sesion) {
  clearTimeout(temporizador)
  temporizador = setTimeout(
    cerrarSesion,
    Math.min(Math.max(sesion.expiraEn - Date.now(), 0), PLAZO_MAXIMO_MS),
  )
}

// El rol se lee del claim `rol` del JWT (backend/internal/auth/jwt.go) solo para decidir qué
// pantallas mostrar; la autorización real la hace el servidor en cada petición.
function rolDelToken(token: string): Rol | undefined {
  try {
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    const rol = payload.rol
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
  programarVencimiento(actual)
  emitir()
}

export function cerrarSesion() {
  clearTimeout(temporizador)
  temporizador = undefined
  if (!actual) return
  actual = null
  try {
    sessionStorage.removeItem(CLAVE)
  } catch {
    // nada que limpiar
  }
  emitir()
}

// Sin efectos secundarios: React la usa como getSnapshot y puede llamarla varias veces por render.
export function obtenerSesion(): Sesion | null {
  return actual
}

// Al cargar la página se recupera la sesión guardada y se programa su vencimiento.
actual = leerAlmacenada()
if (actual) programarVencimiento(actual)

// Para el cliente HTTP (fuera del render): además del temporizador, no envía un token vencido
// si el navegador suspendió la pestaña y el temporizador todavía no corrió.
export function obtenerToken(): string | null {
  if (actual && actual.expiraEn <= Date.now()) cerrarSesion()
  return actual?.token ?? null
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
