import { useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { guardarPostulacion, obtenerPostulacion } from '@/services/postulacion'
import { motivoDeBloqueo } from './informacionPersonal'
import type { InformacionPersonalProps } from './secciones/InformacionPersonal'

const cargarPostulacion = () => obtenerPostulacion()

// PUT /postulacion reemplaza el estado completo. Los datos personales no se editan, así que
// "mezclar la sección" es reenviar la postulación tal como está en el servidor; se lee en el
// momento de guardar para no pisar lo que otra sección guardó después de cargar la página.
// El servidor también hace aquí la validación de nacionalidad: 422 y cierre si no es
// ecuatoriana, 409 si la postulación ya no está en borrador.
async function guardarInformacionPersonal() {
  const actual = await obtenerPostulacion()
  return guardarPostulacion(actual)
}

// Conecta la sección Información personal con la API. El contenedor del formulario lo monta
// así, y con `bloqueo` decide si muestra el resto de las secciones:
//
//	const { cargando, errorCarga, recargar, bloqueo, props } = useInformacionPersonal()
//	{bloqueo ? <PostulacionBloqueada motivo={bloqueo} /> : props && <InformacionPersonal {...props} />}
export function useInformacionPersonal() {
  const carga = useApi(cargarPostulacion)
  const guardado = useApi(guardarInformacionPersonal)
  const { ejecutar: cargar } = carga
  const { ejecutar: guardar } = guardado

  useEffect(() => {
    void cargar()
  }, [cargar])

  // Si el servidor rechaza el guardado por la regla de nacionalidad (422) o porque la
  // postulación ya no es editable (409), se vuelve a leer para mostrar el bloqueo.
  const estadoError = guardado.error?.status
  useEffect(() => {
    if (estadoError === 422 || estadoError === 409) void cargar()
  }, [estadoError, cargar])

  const postulacion = guardado.data ?? carga.data
  const props = useMemo<InformacionPersonalProps | null>(
    () =>
      postulacion && {
        informacion: postulacion.informacion_personal,
        onGuardar: async () => {
          await guardar()
        },
        guardando: guardado.cargando,
        guardado: guardado.data !== null && !guardado.cargando,
        error: guardado.error,
      },
    [postulacion, guardar, guardado.cargando, guardado.data, guardado.error],
  )

  return {
    // Antes del primer efecto todavía no hay datos ni error: también es "cargando".
    cargando: carga.cargando || (carga.data === null && carga.error === null),
    errorCarga: carga.error,
    recargar: cargar,
    bloqueo: postulacion ? motivoDeBloqueo(postulacion) : null,
    codigo: postulacion?.codigo,
    props,
  }
}
