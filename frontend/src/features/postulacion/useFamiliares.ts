import { useCallback, useEffect, useMemo } from 'react'
import { useApi } from '@/hooks/useApi'
import { guardarPostulacion, obtenerPostulacion } from '@/services/postulacion'
import type { FamiliaresProps, ValoresFamiliares } from './secciones/Familiares'

// useApi pasa la señal de cancelación como primer argumento; estas llamadas la ignoran.
const cargar = () => obtenerPostulacion()

// Guarda la sección de familiares. PUT /postulacion reemplaza el estado completo, así que
// se lee la postulación justo antes, se cambia solo esta sección y se envía el formulario
// entero: enviar solo los familiares borraría el domicilio y las demás secciones. La
// lectura es nueva en cada guardado para no pisar lo que otra sección guardó después de
// cargar la página.
export async function guardarFamiliares(valores: ValoresFamiliares): Promise<void> {
  const actual = await obtenerPostulacion()
  await guardarPostulacion({
    ...actual,
    tiene_familiares: valores.tiene_familiares,
    familiares: valores.familiares,
  })
}

// Conecta la sección Familiares con la API: carga la postulación y entrega las props del
// componente. El contenedor del formulario lo monta así:
//
//	const { cargando, error, recargar, props } = useFamiliares()
//	{props && <Familiares {...props} />}
export function useFamiliares() {
  const { data, cargando, error, ejecutar } = useApi(cargar)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  const props = useMemo<FamiliaresProps | null>(
    () =>
      data && {
        valoresIniciales: { tiene_familiares: data.tiene_familiares, familiares: data.familiares },
        onGuardar: guardarFamiliares,
      },
    [data],
  )
  const recargar = useCallback(() => void ejecutar(), [ejecutar])

  return { cargando, error, recargar, props }
}
