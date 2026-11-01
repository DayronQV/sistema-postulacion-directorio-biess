import { useCallback, useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { eliminarDocumento } from '@/services/documentos'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type FormacionAcademica,
} from '@/services/postulacion'
import type { FormacionProps } from './secciones/Formacion'

const cargar = () => obtenerPostulacion()

export async function guardarFormacion(formacion: FormacionAcademica[]): Promise<void> {
  const actual = await obtenerPostulacion()
  await guardarPostulacion({ ...actual, formacion })

  // Después de cambiar las referencias, los respaldos eliminados o reemplazados ya
  // pueden borrarse. Antes del PUT el servidor responde 409 porque todavía están en uso.
  const usados = new Set(formacion.map((fila) => fila.documento_id))
  const huerfanos = new Set(
    (actual.formacion ?? [])
      .map((fila) => fila.documento_id)
      .filter((documentoId) => documentoId && !usados.has(documentoId)),
  )
  await Promise.allSettled(
    [...huerfanos].map((documentoId) =>
      eliminarDocumento(new AbortController().signal, documentoId),
    ),
  )
}

export function useFormacion() {
  const { data, cargando, error, ejecutar } = useApi(cargar)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  const props = useMemo<FormacionProps | null>(
    () => data && { valor: data.formacion ?? [], onGuardar: guardarFormacion },
    [data],
  )
  const recargar = useCallback(() => void ejecutar(), [ejecutar])

  return { cargando, error, recargar, props }
}
