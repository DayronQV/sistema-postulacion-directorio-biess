import { useCallback, useEffect, useMemo, useState } from 'react'
import { useApi } from '@/hooks/useApi'
import { ApiError } from '@/services/errores'
import {
  guardarPostulacion,
  obtenerPostulacion,
  validarCargo,
  type Postulacion,
  type ResultadoCargo,
  type TipoCargo,
} from '@/services/postulacion'
import { motivoDeBloqueo } from './informacionPersonal'
import { esCargo } from './reglasCargo'
import type { CargoProps, ValoresCargo } from './secciones/Cargo'

// useApi pasa la señal de cancelación como primer argumento; esta llamada la ignora.
const cargar = () => obtenerPostulacion()

// Guarda la sección. PUT /postulacion reemplaza el estado completo: se lee la postulación
// justo antes, se cambian solo el cargo y el respaldo, y se envía el formulario entero para
// no borrar las demás secciones. El servidor vuelve a validar el cargo con el Historial
// Laboral en cada guardado (422 y cierre si ya no cumple).
export async function guardarCargo(valores: ValoresCargo): Promise<void> {
  const actual = await obtenerPostulacion()
  await guardarPostulacion({
    ...actual,
    cargo: valores.cargo,
    respaldo_organizacion: valores.respaldo_organizacion,
  })
}

function cargoGuardado(p: Postulacion): TipoCargo | '' {
  return esCargo(p.cargo) ? p.cargo : ''
}

// Conecta la sección Selección del cargo con la API. El contenedor del formulario la monta
// así, y con `bloqueo` decide si muestra el resto de las secciones:
//
//	const { cargando, error, recargar, bloqueo, codigo, props } = useCargo()
//	{bloqueo ? <PostulacionBloqueada motivo={bloqueo} codigo={codigo} /> : props && <Cargo {...props} />}
export function useCargo() {
  const { data, cargando, error, ejecutar } = useApi(cargar)
  // Mensaje del servidor cuando rechaza el cargo (422): la postulación queda cerrada.
  const [bloqueoServidor, setBloqueoServidor] = useState<string | null>(null)

  useEffect(() => {
    void ejecutar()
  }, [ejecutar])

  // 422: el postulante no cumple la calidad del cargo (o su nacionalidad) y el servidor ya
  // cerró la postulación. 409: ya no es un borrador; se vuelve a leer para mostrar su estado.
  // El error se relanza para que la sección no siga como si hubiera funcionado.
  const alFallar = useCallback(
    (fallo: unknown): never => {
      if (fallo instanceof ApiError) {
        if (fallo.status === 422) setBloqueoServidor(fallo.message)
        if (fallo.status === 409) void ejecutar()
      }
      throw fallo
    },
    [ejecutar],
  )

  const props = useMemo<CargoProps | null>(
    () =>
      data && {
        valoresIniciales: {
          cargo: cargoGuardado(data),
          respaldo_organizacion: {
            nombre: data.respaldo_organizacion?.nombre ?? '',
            documento_id: data.respaldo_organizacion?.documento_id ?? null,
          },
        },
        onValidar: (cargo): Promise<ResultadoCargo> => validarCargo(cargo).catch(alFallar),
        onGuardar: (valores) => guardarCargo(valores).catch(alFallar),
      },
    [data, alFallar],
  )
  const recargar = useCallback(() => void ejecutar(), [ejecutar])

  return {
    // Antes del primer efecto todavía no hay datos ni error: también es "cargando".
    cargando: cargando || (data === null && error === null),
    error,
    recargar,
    bloqueo: bloqueoServidor ?? (data ? motivoDeBloqueo(data) : null),
    codigo: data?.codigo,
    props,
  }
}
