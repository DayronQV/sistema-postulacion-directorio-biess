import { useSyncExternalStore } from 'react'

export type Rol = 'postulante' | 'analista' | 'administrador'

export interface Sesion {
  token: string
  expiraEn: number // marca de tiempo en ms
  rol?: Rol
  // Primer ingreso: el postulante aún no tiene postulación y solo puede abrir el
  // formulario hasta que se cree (ver marcarPostulacionCreada).
  primerIngreso?: boolean
}

export interface OpcionesSesion {
  primerIngreso?: boolean
}

const CLAVE = 'biess.sesion'
const escuchas = new Set<() => void>()

// El token vive solo en la pestaña (sessionStorage). Se guarda una referencia para que
// useSyncExternalStore reciba siempre el mismo objeto mientras la sesión no cambie.
let actual: Sesion | null = null
let temporizador: ReturnType<typeof setTimeout> | undefined
// true si la última sesión terminó por vencimiento del token (o un 401), no porque el
// usuario la cerró: el login lo usa para avisar "Su sesión expiró".
let expirada = false

// setTimeout no admite plazos mayores a ~24,8 días.
const PLAZO_MAXIMO_MS = 2 ** 31 - 1

function leerAlmacenada(): Sesion | null {
  try {
    const crudo = sessionStorage.getItem(CLAVE)
    if (!crudo) return null
    const sesion = JSON.parse(crudo) as Sesion
    if (sesion.expiraEn > Date.now()) return sesion
    // Venció mientras la pestaña estaba cerrada o recargándose.
    sessionStorage.removeItem(CLAVE)
    expirada = true
    return null
  } catch {
    return null
  }
}

function guardar(sesion: Sesion) {
  try {
    sessionStorage.setItem(CLAVE, JSON.stringify(sesion))
  } catch {
    // sin almacenamiento: la sesión dura lo que dure la pestaña en memoria
  }
}

function emitir() {
  escuchas.forEach((f) => f())
}

// Programa el cierre de la sesión cuando vence el token, para que las rutas protegidas
// redirijan al login aunque el usuario no haga nada.
function programarVencimiento(sesion: Sesion) {
  clearTimeout(temporizador)
  temporizador = setTimeout(
    expirarSesion,
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

export function iniciarSesion(
  token: string,
  expiresInSegundos: number,
  { primerIngreso = false }: OpcionesSesion = {},
) {
  actual = {
    token,
    expiraEn: Date.now() + expiresInSegundos * 1000,
    rol: rolDelToken(token),
    primerIngreso,
  }
  expirada = false
  guardar(actual)
  programarVencimiento(actual)
  emitir()
}

function terminar(porVencimiento: boolean) {
  clearTimeout(temporizador)
  temporizador = undefined
  if (!actual) return
  actual = null
  expirada = porVencimiento
  try {
    sessionStorage.removeItem(CLAVE)
  } catch {
    // nada que limpiar
  }
  emitir()
}

// Cierre pedido por el usuario (botón Cerrar sesión).
export function cerrarSesion() {
  terminar(false)
}

// Cierre porque el token venció o el servidor lo rechazó (401).
export function expirarSesion() {
  terminar(true)
}

// true si la última sesión terminó por vencimiento y aún no se volvió a iniciar sesión.
export function sesionExpirada(): boolean {
  return expirada
}

// La postulación ya existe (GET /postulacion la crea en el primer ingreso): se levanta la
// restricción del primer ingreso y el postulante puede abrir las demás rutas privadas.
export function marcarPostulacionCreada() {
  if (!actual?.primerIngreso) return
  actual = { ...actual, primerIngreso: false }
  guardar(actual)
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
  if (actual && actual.expiraEn <= Date.now()) expirarSesion()
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
