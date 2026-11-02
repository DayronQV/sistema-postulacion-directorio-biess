import { useCallback, useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { eliminarDocumento } from '@/services/documentos'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type ExperienciaProfesional,
} from '@/services/postulacion'
import type { ExperienciaProps } from './secciones/Experiencia'

const cargar = () => obtenerPostulacion()

export async function guardarExperiencia(experiencia: ExperienciaProfesional[]): Promise<void> {
  const actual = await obtenerPostulacion()
  await guardarPostulacion({ ...actual, experiencia })

  const usados = new Set(experiencia.map((fila) => fila.documento_id))
  const huerfanos = new Set(
    (actual.experiencia ?? [])
      .map((fila) => fila.documento_id)
      .filter((documentoId) => documentoId && !usados.has(documentoId)),
  )
  await Promise.allSettled(
    [...huerfanos].map((documentoId) =>
      eliminarDocumento(new AbortController().signal, documentoId),
    ),
  )
}

export function useExperiencia() {
  const { data, cargando, error, ejecutar } = useApi(cargar)
  useEffect(() => { void ejecutar() }, [ejecutar])
  const props = useMemo<ExperienciaProps | null>(
    () => data && { valor: data.experiencia ?? [], onGuardar: guardarExperiencia },
    [data],
  )
  const recargar = useCallback(() => void ejecutar(), [ejecutar])
  return { cargando, error, recargar, props }
}
