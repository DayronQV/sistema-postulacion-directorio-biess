import { useCallback, useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type ReferenciaPersonal,
} from '@/services/postulacion'
import type { ReferenciasProps } from './secciones/Referencias'

// useApi pasa la señal de cancelación como primer argumento; esta llamada la ignora.
const cargar = () => obtenerPostulacion()

// Guarda la sección de referencias. Igual que guardarFamiliares: PUT /postulacion
// reemplaza el estado completo, así que se lee la postulación justo antes, se cambia solo
// esta sección y se envía el formulario entero para no borrar las demás.
export async function guardarReferencias(referencias: ReferenciaPersonal[]): Promise<void> {
  const actual = await obtenerPostulacion()
  await guardarPostulacion({ ...actual, referencias })
}

// Conecta la sección Referencias con la API: carga la postulación y entrega las props del
// componente. El contenedor del formulario lo monta así:
//
//	const { cargando, error, recargar, props } = useReferencias()
//	{props && <Referencias {...props} />}
export function useReferencias() {
  const { data, cargando, error, ejecutar } = useApi(cargar)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  const props = useMemo<ReferenciasProps | null>(
    () => data && { valor: data.referencias ?? [], onGuardar: guardarReferencias },
    [data],
  )
  const recargar = useCallback(() => void ejecutar(), [ejecutar])

  return { cargando, error, recargar, props }
}
