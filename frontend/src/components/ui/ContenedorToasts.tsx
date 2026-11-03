import { cerrarToast, useToasts } from '@/store/toasts'
import Alert from './Alert'

// Avisos flotantes (toasts), arriba a la derecha y por encima de la barra superior y de las
// ventanas. Cada uno se quita solo (ver DURACION_TOAST_MS) o con su botón de cerrar. Se monta
// una sola vez, en App; para mostrar uno se llama a mostrarToast.
export default function ContenedorToasts() {
  const toasts = useToasts()
  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed inset-x-4 top-20 z-[70] flex flex-col items-end gap-2 sm:inset-x-auto sm:right-6 sm:w-80"
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className="toast-entrada pointer-events-auto w-full max-w-sm rounded-xl shadow-lg ring-1 ring-black/5"
        >
          <Alert tipo={toast.tipo} titulo={toast.mensaje} onCerrar={() => cerrarToast(toast.id)} />
        </div>
      ))}
    </div>
  )
}
