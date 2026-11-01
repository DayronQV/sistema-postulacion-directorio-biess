import { useCallback, useEffect } from 'react'
import { useApi } from '@/hooks/useApi'
import {
  listarCantones,
  listarPaises,
  listarProvincias,
  type ItemCatalogo,
} from '@/services/catalogos'
import type { ApiError } from '@/services/errores'
import { PAIS_ECUADOR } from './domicilio'

export interface ListaCatalogo {
  items: ItemCatalogo[]
  cargando: boolean
  error: ApiError | null
  reintentar: () => void
}

type Cargar = (signal: AbortSignal, id: number) => Promise<ItemCatalogo[]>

// Carga una lista que depende de `padreId` (null = no corresponde cargarla). Recuerda de
// qué padre es lo que tiene: así, al cambiar el padre, la lista anterior nunca se muestra
// mientras llega la nueva. `cargar` debe ser estable (función de módulo).
function useLista(padreId: number | null, cargar: Cargar): ListaCatalogo {
  const llamada = useCallback(
    async (signal: AbortSignal, id: number) => ({ id, items: await cargar(signal, id) }),
    [cargar],
  )
  const { data, error, ejecutar } = useApi(llamada)

  useEffect(() => {
    if (padreId !== null) void ejecutar(padreId)
  }, [padreId, ejecutar])

  const items = padreId !== null && data?.id === padreId ? data.items : null
  const errorVigente = padreId !== null ? error : null
  return {
    items: items ?? [],
    cargando: padreId !== null && items === null && errorVigente === null,
    error: errorVigente,
    reintentar: () => {
      if (padreId !== null) void ejecutar(padreId)
    },
  }
}

const cargarPaises: Cargar = (signal) => listarPaises(signal)

// Listas en cascada país → provincia → cantón. Las provincias solo se piden si el país es
// Ecuador y los cantones solo cuando hay provincia.
export function useCatalogosGeo(paisId: string, provinciaId: string) {
  const paises = useLista(0, cargarPaises)
  const pais = paises.items.find((p) => String(p.id) === paisId)
  const esEcuador = pais?.nombre === PAIS_ECUADOR

  const provincias = useLista(esEcuador ? Number(paisId) : null, listarProvincias)
  const cantones = useLista(esEcuador && provinciaId ? Number(provinciaId) : null, listarCantones)

  return { paises, provincias, cantones, esEcuador, hayPais: pais !== undefined }
}
