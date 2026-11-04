import { vi } from 'vitest'
import { ApiError } from '@/services/errores'
import * as servicio from '@/services/postulacion'
import {
  cargarPostulacion,
  cargarPostulacionSiFalta,
  enviarFormulario,
  guardarSeccion,
  marcarCambios,
  reiniciarFormulario,
} from './store'
import { useFormulario } from './store'
import { renderHook } from '@testing-library/react'
import { reiniciarToasts, useToasts } from '@/store/toasts'

vi.mock('@/services/postulacion')
const obtener = vi.mocked(servicio.obtenerPostulacion)
const guardar = vi.mocked(servicio.guardarPostulacion)
const enviar = vi.mocked(servicio.enviarPostulacion)

// Postulación mínima válida para los tipos; cada prueba cambia lo que le interesa.
function post(parcial: Partial<servicio.Postulacion> = {}): servicio.Postulacion {
  return { id: 'p1', codigo: 'A', estado: 'borrador', ...parcial } as servicio.Postulacion
}

function estado() {
  return renderHook(() => useFormulario()).result.current
}

beforeEach(() => {
  reiniciarToasts()
  reiniciarFormulario()
  obtener.mockReset()
  guardar.mockReset()
  enviar.mockReset()
})

describe('store de la postulación', () => {
  it('carga la postulación', async () => {
    obtener.mockResolvedValue(post())
    await cargarPostulacion()
    expect(estado().postulacion?.codigo).toBe('A')
    expect(estado().cargando).toBe(false)
  })

  it('cargarPostulacionSiFalta carga solo si no hay postulación ni una carga en curso', async () => {
    let resolver: (p: servicio.Postulacion) => void = () => {}
    obtener.mockReturnValue(new Promise((r) => (resolver = r)))

    const primera = cargarPostulacionSiFalta()
    await cargarPostulacionSiFalta() // ya hay una carga en curso
    expect(obtener).toHaveBeenCalledTimes(1)

    resolver(post())
    await primera
    await cargarPostulacionSiFalta() // ya está cargada
    expect(obtener).toHaveBeenCalledTimes(1)
    expect(estado().postulacion?.codigo).toBe('A')
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

  it('al guardar una sección muestra un toast de éxito', async () => {
    obtener.mockResolvedValue(post())
    await cargarPostulacion()
    guardar.mockResolvedValue(post({ cargo: 'jubilados' }))

    await guardarSeccion('cargo', { cargo: 'jubilados' })

    const avisos = renderHook(() => useToasts()).result.current
    expect(avisos).toEqual([
      expect.objectContaining({
        tipo: 'success',
        mensaje: 'La información se guardó correctamente.',
      }),
    ])
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
    // Sin éxito no hay toast.
    expect(renderHook(() => useToasts()).result.current).toHaveLength(0)
  })

  it('enviar deja la postulación en estado enviada y limpia los cambios pendientes', async () => {
    obtener.mockResolvedValue(post({ cargo: 'afiliados' }))
    await cargarPostulacion()
    enviar.mockResolvedValue(post({ cargo: 'afiliados', estado: 'enviada' }))
    marcarCambios('cierre', true)

    const error = await enviarFormulario()

    expect(error).toBeNull()
    expect(estado().postulacion).toMatchObject({ estado: 'enviada', cargo: 'afiliados' })
    expect(estado().seccionesConCambios).toEqual([])
    expect(estado().guardando).toBe(false)
  })

  it('si el envío falla devuelve el error y la postulación sigue en borrador', async () => {
    obtener.mockResolvedValue(post())
    await cargarPostulacion()
    enviar.mockRejectedValue(
      new ApiError(422, 'VALIDATION_ERROR', 'La postulación está incompleta', {
        cargo: 'Seleccione el cargo',
      }),
    )

    const error = await enviarFormulario()

    expect(error?.fields).toEqual({ cargo: 'Seleccione el cargo' })
    expect(estado().postulacion?.estado).toBe('borrador')
    expect(estado().guardando).toBe(false)
  })

  it('si ya estaba enviada (409) vuelve a leer la postulación para mostrarlo', async () => {
    obtener.mockResolvedValueOnce(post())
    await cargarPostulacion()
    enviar.mockRejectedValue(new ApiError(409, 'CONFLICT', 'La postulación ya fue enviada'))
    obtener.mockResolvedValueOnce(post({ estado: 'enviada' }))

    const error = await enviarFormulario()

    expect(error?.code).toBe('CONFLICT')
    expect(estado().postulacion?.estado).toBe('enviada')
  })
})
