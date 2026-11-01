import type { FieldValues, Path, UseFormSetError } from 'react-hook-form'
import type { ApiError } from '@/services/errores'

// Pasa los errores por campo del servidor (`fields`, con los nombres de la API) a los
// campos del formulario. `nombres` traduce el nombre de la API al del formulario.
// Devuelve true si se aplicó al menos uno.
export function aplicarErroresDeCampo<T extends FieldValues>(
  error: ApiError,
  setError: UseFormSetError<T>,
  nombres: Partial<Record<string, Path<T>>>,
): boolean {
  let aplicado = false
  for (const [campoApi, mensaje] of Object.entries(error.fields ?? {})) {
    const campo = nombres[campoApi]
    if (campo) {
      setError(campo, { type: 'server', message: mensaje })
      aplicado = true
    }
  }
  return aplicado
}
