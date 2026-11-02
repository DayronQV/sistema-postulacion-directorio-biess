import { useCallback, useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { eliminarDocumento } from '@/services/documentos'
import { guardarPostulacion, obtenerPostulacion, type PublicacionAcademica } from '@/services/postulacion'
import type { PublicacionesProps } from './secciones/Publicaciones'

const cargar = () => obtenerPostulacion()

export async function guardarPublicaciones(tiene: boolean, publicaciones: PublicacionAcademica[]): Promise<void> {
  const actual = await obtenerPostulacion()
  const nuevas = tiene ? publicaciones : []
  await guardarPostulacion({ ...actual, tiene_publicaciones: tiene, publicaciones: nuevas })
  const usados = new Set(nuevas.map((fila) => fila.documento_id))
  const huerfanos = new Set((actual.publicaciones ?? []).map((fila) => fila.documento_id).filter((id) => id && !usados.has(id)))
  await Promise.allSettled([...huerfanos].map((id) => eliminarDocumento(new AbortController().signal, id)))
}

export function usePublicaciones() {
  const { data, cargando, error, ejecutar } = useApi(cargar)
  useEffect(() => { void ejecutar() }, [ejecutar])
  const props = useMemo<PublicacionesProps | null>(() => data && { tienePublicaciones: data.tiene_publicaciones, valor: data.publicaciones ?? [], onGuardar: guardarPublicaciones }, [data])
  const recargar = useCallback(() => void ejecutar(), [ejecutar])
  return { cargando, error, recargar, props }
}
