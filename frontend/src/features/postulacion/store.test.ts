import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
import * as servicio from '@/services/postulacion'
import {
  cargarPostulacion,
  guardarSeccion,
  marcarCambios,
  reiniciarFormulario,
} from './store'
import { useFormulario } from './store'
import { renderHook } from '@testing-library/react'

vi.mock('@/services/postulacion')
const obtener = vi.mocked(servicio.obtenerPostulacion)
const guardar = vi.mocked(servicio.guardarPostulacion)

// Postulación mínima válida para los tipos; cada prueba cambia lo que le interesa.
function post(parcial: Partial<servicio.Postulacion> = {}): servicio.Postulacion {
  return { id: 'p1', codigo: 'A', estado: 'borrador', ...parcial } as servicio.Postulacion
}

function estado() {
  return renderHook(() => useFormulario()).result.current
}

beforeEach(() => {
  reiniciarFormulario()
  obtener.mockReset()
  guardar.mockReset()
})

describe('store de la postulación', () => {
  it('carga la postulación', async () => {
    obtener.mockResolvedValue(post())
    await cargarPostulacion()
    expect(estado().postulacion?.codigo).toBe('A')
    expect(estado().cargando).toBe(false)
  })

  it('guarda una sección enviando el estado completo y conserva lo ya cargado', async () => {
    obtener.mockResolvedValue(post({ tiene_familiares: true }))
    await cargarPostulacion()
    guardar.mockResolvedValue({} as servicio.Postulacion)

    marcarCambios('adicional', true)
    const ok = await guardarSeccion('adicional', { cargo: 'afiliados' })

    expect(ok).toBe(true)
    expect(guardar).toHaveBeenCalledWith(
      expect.objectContaining({ codigo: 'A', tiene_familiares: true, cargo: 'afiliados' }),
    )
    expect(estado().postulacion).toMatchObject({ codigo: 'A', cargo: 'afiliados' })
    expect(estado().seccionesConCambios).toEqual([])
  })

  it('si el servidor rechaza, deja el error y conserva los cambios sin guardar', async () => {
    obtener.mockResolvedValue(post())
    await cargarPostulacion()
    guardar.mockRejectedValue(new ApiError(422, 'VALIDATION_ERROR', 'Cargo no elegible'))

    marcarCambios('cargo', true)
    const ok = await guardarSeccion('cargo', { cargo: 'jubilados' })

    expect(ok).toBe(false)
    expect(estado().errorGuardado?.message).toBe('Cargo no elegible')
    expect(estado().seccionesConCambios).toEqual(['cargo'])
    expect(estado().postulacion).not.toHaveProperty('cargo')
  })
})
