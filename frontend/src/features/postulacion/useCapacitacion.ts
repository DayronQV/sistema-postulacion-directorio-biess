import { useCallback, useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { eliminarDocumento } from '@/services/documentos'
import { guardarPostulacion, obtenerPostulacion, type Capacitacion } from '@/services/postulacion'
import type { CapacitacionProps } from './secciones/Capacitacion'

// useApi pasa la señal de cancelación como primer argumento; esta llamada la ignora.
const cargar = () => obtenerPostulacion()

// Guarda la sección de capacitación. PUT /postulacion reemplaza el estado completo, así
// que se lee la postulación justo antes, se cambia solo esta sección y se envía entera.
// Después del PUT ya se pueden borrar los respaldos que quedaron sin referencia (antes el
// servidor respondería 409 porque todavía estaban en uso).
export async function guardarCapacitaciones(capacitaciones: Capacitacion[]): Promise<void> {
  const actual = await obtenerPostulacion()
  await guardarPostulacion({ ...actual, capacitaciones })

  const usados = new Set(capacitaciones.map((fila) => fila.documento_id))
  const huerfanos = new Set(
    (actual.capacitaciones ?? [])
      .map((fila) => fila.documento_id)
      .filter((id) => id && !usados.has(id)),
  )
  await Promise.allSettled(
    [...huerfanos].map((id) => eliminarDocumento(new AbortController().signal, id)),
  )
}

// Conecta la sección Capacitación con la API: carga la postulación y entrega las props.
export function useCapacitacion() {
  const { data, cargando, error, ejecutar } = useApi(cargar)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  const props = useMemo<CapacitacionProps | null>(
    () => data && { valor: data.capacitaciones ?? [], onGuardar: guardarCapacitaciones },
    [data],
  )
  const recargar = useCallback(() => void ejecutar(), [ejecutar])

  return { cargando, error, recargar, props }
}
