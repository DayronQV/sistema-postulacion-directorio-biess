import type { Domicilio } from '@/services/postulacion'
import { useCatalogosGeo } from '../useCatalogosGeo'

export interface Ubicacion {
  pais: string
  provincia: string
  ciudad: string
}

// Nombres de país, provincia y ciudad del domicilio. La API guarda ids de catálogo cuando
// el país es Ecuador y texto libre en otro país; el documento impreso debe mostrar nombres.
export function useUbicacion(dom: Domicilio | undefined): Ubicacion {
  const paisId = dom?.pais_id ? String(dom.pais_id) : ''
  const provinciaId = dom?.provincia_id ? String(dom.provincia_id) : ''
  const { paises, provincias, cantones, esEcuador } = useCatalogosGeo(paisId, provinciaId)

  const nombre = (items: { id: number; nombre: string }[], id: number | null | undefined) =>
    items.find((i) => i.id === id)?.nombre ?? ''

  return {
    pais: nombre(paises.items, dom?.pais_id),
    provincia: esEcuador
      ? nombre(provincias.items, dom?.provincia_id)
      : (dom?.provincia_texto ?? ''),
    ciudad: esEcuador ? nombre(cantones.items, dom?.canton_id) : (dom?.ciudad_texto ?? ''),
  }
}
