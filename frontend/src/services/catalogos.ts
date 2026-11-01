import { http } from './http'

// Catálogos geográficos (docs/api.md, sección Catálogos). Son públicos y el servidor los
// guarda en caché. Reciben la señal de cancelación como primer argumento para usarse con
// useApi.

export interface ItemCatalogo {
  id: number
  nombre: string
}

// GET /catalogos/paises: Ecuador va primero.
export async function listarPaises(signal: AbortSignal) {
  const { data } = await http.get<ItemCatalogo[]>('/catalogos/paises', { signal })
  return data
}

// GET /catalogos/paises/{id}/provincias
export async function listarProvincias(signal: AbortSignal, paisId: number) {
  const { data } = await http.get<ItemCatalogo[]>(`/catalogos/paises/${paisId}/provincias`, {
    signal,
  })
  return data
}

// GET /catalogos/provincias/{id}/cantones
export async function listarCantones(signal: AbortSignal, provinciaId: number) {
  const { data } = await http.get<ItemCatalogo[]>(`/catalogos/provincias/${provinciaId}/cantones`, {
    signal,
  })
  return data
}
