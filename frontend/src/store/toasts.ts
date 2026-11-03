import { useSyncExternalStore } from 'react'
import type { TipoAlert } from '@/components/ui/Alert'

export interface Toast {
  id: number
  tipo: TipoAlert
  mensaje: string
}

// Cuánto tiempo se ve un toast antes de quitarse solo.
export const DURACION_TOAST_MS = 4000
// Máximo de toasts a la vez: si llega otro, se quita el más antiguo.
const MAXIMO_TOASTS = 3

const escuchas = new Set<() => void>()
const temporizadores = new Map<number, ReturnType<typeof setTimeout>>()

// Se reemplaza el arreglo entero en cada cambio: useSyncExternalStore necesita recibir la
// misma referencia mientras no cambie nada.
let toasts: Toast[] = []
let siguienteId = 1

function poner(nuevos: Toast[]) {
  toasts = nuevos
  escuchas.forEach((escucha) => escucha())
}

function programarCierre(id: number, duracion: number) {
  clearTimeout(temporizadores.get(id))
  temporizadores.set(
    id,
    setTimeout(() => cerrarToast(id), duracion),
  )
}

// Muestra un aviso flotante que se quita solo (por defecto, a los 4 s). Si ya se está viendo
// el mismo mensaje, no se apila otro: solo se reinicia su tiempo, para que guardar varias
// veces seguidas no llene la pantalla de avisos iguales.
export function mostrarToast(
  mensaje: string,
  tipo: TipoAlert = 'success',
  duracion = DURACION_TOAST_MS,
): number {
  const repetido = toasts.find((t) => t.tipo === tipo && t.mensaje === mensaje)
  if (repetido) {
    programarCierre(repetido.id, duracion)
    return repetido.id
  }
  const id = siguienteId++
  const lista = [...toasts, { id, tipo, mensaje }]
  for (const sobrante of lista.splice(0, Math.max(0, lista.length - MAXIMO_TOASTS))) {
    clearTimeout(temporizadores.get(sobrante.id))
    temporizadores.delete(sobrante.id)
  }
  poner(lista)
  programarCierre(id, duracion)
  return id
}

export function cerrarToast(id: number) {
  clearTimeout(temporizadores.get(id))
  temporizadores.delete(id)
  if (toasts.some((t) => t.id === id)) poner(toasts.filter((t) => t.id !== id))
}

// Quita todos los toasts (para las pruebas).
export function reiniciarToasts() {
  temporizadores.forEach((temporizador) => clearTimeout(temporizador))
  temporizadores.clear()
  poner([])
}

function suscribir(escucha: () => void) {
  escuchas.add(escucha)
  return () => escuchas.delete(escucha)
}

export function useToasts() {
  return useSyncExternalStore(suscribir, () => toasts)
}
