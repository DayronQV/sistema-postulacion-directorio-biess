import { useSyncExternalStore } from 'react'
import {
  enviarPostulacion,
  guardarPostulacion,
  obtenerPostulacion,
  suscribirGuardado,
  type DatosPostulacion,
  type Postulacion,
} from '@/services/postulacion'
import { ApiError } from '@/services/errores'
import { mostrarToast } from '@/store/toasts'

export interface EstadoFormulario {
  postulacion: Postulacion | null
  cargando: boolean
  errorCarga: ApiError | null
  guardando: boolean
  errorGuardado: ApiError | null
  seccionesConCambios: string[] // secciones con cambios sin guardar
}

const inicial: EstadoFormulario = {
  postulacion: null,
  cargando: false,
  errorCarga: null,
  guardando: false,
  errorGuardado: null,
  seccionesConCambios: [],
}

let estado: EstadoFormulario = inicial
const escuchas = new Set<() => void>()
let cargaEnCurso: AbortController | null = null

function poner(parcial: Partial<EstadoFormulario>) {
  estado = { ...estado, ...parcial }
  escuchas.forEach((f) => f())
}

function comoApiError(e: unknown): ApiError {
  return e instanceof ApiError ? e : new ApiError(0, 'UNKNOWN', 'Ocurrió un error inesperado.')
}

export async function cargarPostulacion(): Promise<void> {
  cargaEnCurso?.abort()
  const control = new AbortController()
  cargaEnCurso = control
  poner({ cargando: true, errorCarga: null })
  try {
    const postulacion = await obtenerPostulacion(control.signal)
    if (control.signal.aborted) return
    poner({ postulacion, cargando: false, seccionesConCambios: [] })
  } catch (e) {
    if (control.signal.aborted) return
    poner({ cargando: false, errorCarga: comoApiError(e) })
  }
}

// Vuelve a leer la postulación sin mostrar el cargador ni tocar los cambios sin guardar. Las
// secciones que se conectan solas a la API guardan por su cuenta: al salir de ellas se
// refrescan el avance y las marcas de sección completa.
export async function refrescarPostulacion(): Promise<void> {
  try {
    const postulacion = await obtenerPostulacion()
    poner({ postulacion })
  } catch {
    // el refresco es opcional: se conserva lo que ya está en pantalla
  }
}

// Guarda los cambios de una sección. Devuelve true si el servidor los aceptó; si no, deja el
// error en `errorGuardado` y conserva lo que el usuario escribió.
export async function guardarSeccion(
  id: string,
  cambios: Partial<DatosPostulacion>,
): Promise<boolean> {
  if (!estado.postulacion) return false
  const combinada: Postulacion = { ...estado.postulacion, ...cambios }
  poner({ guardando: true, errorGuardado: null })
  try {
    const respuesta = await guardarPostulacion(combinada)
    // Si la respuesta no trae la postulación completa, se conserva lo guardado.
    const postulacion = respuesta && Object.keys(respuesta).length > 0 ? respuesta : combinada
    poner({
      postulacion: { ...combinada, ...postulacion },
      guardando: false,
      seccionesConCambios: estado.seccionesConCambios.filter((s) => s !== id),
    })
    // Inhabilidades y el cierre guardan por aquí y no tenían ningún aviso de éxito.
    mostrarToast('La información se guardó correctamente.')
    return true
  } catch (e) {
    poner({ guardando: false, errorGuardado: comoApiError(e) })
    return false
  }
}

// Envía la postulación. Devuelve null si el servidor la aceptó (el estado queda "enviada") o el
// error: 422 con las secciones incompletas en `fields`, 409 si ya se había enviado.
export async function enviarFormulario(): Promise<ApiError | null> {
  poner({ guardando: true, errorGuardado: null })
  try {
    const enviada = await enviarPostulacion()
    poner({
      postulacion: { ...estado.postulacion, ...enviada },
      guardando: false,
      seccionesConCambios: [],
    })
    return null
  } catch (e) {
    poner({ guardando: false })
    const error = comoApiError(e)
    // 409: ya estaba enviada (por ejemplo, desde otra pestaña). Se relee para mostrarlo.
    if (error.code === 'CONFLICT') await refrescarPostulacion()
    return error
  }
}

export function marcarCambios(id: string, conCambios: boolean) {
  const ya = estado.seccionesConCambios.includes(id)
  if (conCambios === ya) return
  poner({
    seccionesConCambios: conCambios
      ? [...estado.seccionesConCambios, id]
      : estado.seccionesConCambios.filter((s) => s !== id),
  })
}

export function limpiarErrorGuardado() {
  if (estado.errorGuardado) poner({ errorGuardado: null })
}

export function reiniciarFormulario() {
  cargaEnCurso?.abort()
  estado = inicial
  escuchas.forEach((f) => f())
}

function suscribir(escucha: () => void) {
  escuchas.add(escucha)
  return () => {
    escuchas.delete(escucha)
  }
}

// Las secciones guardan con sus propios hooks: cada PUT exitoso actualiza la postulación del
// contenedor (avance y marcas de completa al instante) y quita las marcas de cambios sin
// guardar. Se queda con la respuesta solo si trae la postulación completa.
suscribirGuardado((guardada) => {
  if (!estado.postulacion) return
  const completa = guardada && typeof guardada === 'object' && 'id' in guardada
  poner({
    postulacion: completa ? { ...estado.postulacion, ...guardada } : estado.postulacion,
    seccionesConCambios: [],
  })
})

export function useFormulario(): EstadoFormulario {
  return useSyncExternalStore(suscribir, () => estado)
}
