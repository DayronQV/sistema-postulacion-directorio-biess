import { useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { guardarPostulacion, obtenerPostulacion } from '@/services/postulacion'
import {
  DOMICILIO_VACIO,
  domicilioParaApi,
  valoresDesdeApi,
  type ValoresDomicilio,
} from './domicilio'

const cargarPostulacion = () => obtenerPostulacion()

// PUT /postulacion reemplaza el estado completo: enviar solo el domicilio borraría las
// demás secciones. Por eso se lee la postulación en el momento de guardar (no la del
// montaje: otra sección pudo guardar mientras tanto), se cambia únicamente el domicilio y
// se envía el formulario entero.
async function guardarDomicilio(_signal: AbortSignal, valores: ValoresDomicilio) {
  const actual = await obtenerPostulacion()
  return guardarPostulacion({ ...actual, domicilio: domicilioParaApi(valores) })
}

// Conecta la sección de domicilio con la API: carga la postulación y guarda el domicilio.
export function useDomicilio() {
  const carga = useApi(cargarPostulacion)
  const guardado = useApi(guardarDomicilio)
  const { ejecutar: cargar } = carga
  const { ejecutar: guardar } = guardado

  useEffect(() => {
    void cargar()
  }, [cargar])

  const inicial = carga.data
  const valoresIniciales = useMemo(
    () => (inicial ? valoresDesdeApi(inicial.domicilio) : DOMICILIO_VACIO),
    [inicial],
  )

  return {
    // Antes del primer efecto todavía no hay datos ni error: también es "cargando".
    cargando: carga.cargando || (carga.data === null && carga.error === null),
    errorCarga: carga.error,
    recargar: cargar,
    valoresIniciales,
    correoPrincipal: (guardado.data ?? carga.data)?.informacion_personal.email ?? '',
    onGuardar: async (valores: ValoresDomicilio) => {
      await guardar(valores)
    },
    guardando: guardado.cargando,
    guardado: guardado.data !== null && !guardado.cargando,
    errorGuardado: guardado.error,
  }
}
