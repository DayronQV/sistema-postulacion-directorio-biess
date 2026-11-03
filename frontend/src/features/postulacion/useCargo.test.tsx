import { act, renderHook, waitFor } from '@testing-library/react'
import { vi, type Mock } from 'vitest'
import { ApiError } from '@/services/errores'
import {
  guardarPostulacion,
  obtenerPostulacion,
  validarCargo,
  type Postulacion,
} from '@/services/postulacion'
import { guardarCargo, useCargo } from './useCargo'

vi.mock('@/services/postulacion', () => ({
  obtenerPostulacion: vi.fn(),
  guardarPostulacion: vi.fn(),
  validarCargo: vi.fn(),
}))

const mockObtener = obtenerPostulacion as Mock
const mockGuardar = guardarPostulacion as Mock
const mockValidar = validarCargo as Mock

// Postulación con otras secciones llenas: guardar el cargo no debe perderlas.
const postulacion = {
  id: 'p-1',
  codigo: 'BIESS-2026-000001',
  estado: 'borrador',
  informacion_personal: { nacionalidad: 'Ecuatoriana' },
  domicilio: { pais_id: 1, provincia_id: 17, canton_id: 170, calle_principal: 'Av. Amazonas' },
  tiene_familiares: true,
  familiares: [{ nombres_apellidos: 'Luis' }],
  referencias: [{ nombres_apellidos: 'Rosa' }],
  cargo: '',
  respaldo_organizacion: { nombre: '', documento_id: null },
  consentimiento: true,
} as unknown as Postulacion

beforeEach(() => {
  mockObtener.mockReset()
  mockGuardar.mockReset()
  mockValidar.mockReset()
})

describe('guardarCargo', () => {
  it('envía el formulario completo cambiando solo el cargo y el respaldo', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockResolvedValue(postulacion)
    const respaldo = { nombre: 'Federación', documento_id: 'doc-1' }

    await guardarCargo({ cargo: 'empleador', respaldo_organizacion: respaldo })

    expect(mockGuardar).toHaveBeenCalledWith({
      ...postulacion,
      cargo: 'empleador',
      respaldo_organizacion: respaldo,
    })
  })

  it('lee la postulación justo antes de guardar, no la que cargó la página', async () => {
    mockObtener.mockResolvedValue({ ...postulacion, referencias: [{ nombres_apellidos: 'Nueva' }] })
    mockGuardar.mockResolvedValue(postulacion)

    await guardarCargo({
      cargo: 'afiliados',
      respaldo_organizacion: { nombre: '', documento_id: null },
    })

    expect(mockGuardar.mock.calls[0][0].referencias).toEqual([{ nombres_apellidos: 'Nueva' }])
  })
})

describe('useCargo', () => {
  it('carga la postulación y entrega el cargo guardado y el respaldo al componente', async () => {
    mockObtener.mockResolvedValue({
      ...postulacion,
      cargo: 'jubilados',
      respaldo_organizacion: { nombre: 'Asociación', documento_id: 'doc-2' },
    })

    const { result } = renderHook(() => useCargo())

    expect(result.current.cargando).toBe(true)
    await waitFor(() => expect(result.current.props).not.toBeNull())
    expect(result.current.cargando).toBe(false)
    expect(result.current.bloqueo).toBeNull()
    expect(result.current.props?.valoresIniciales).toEqual({
      cargo: 'jubilados',
      respaldo_organizacion: { nombre: 'Asociación', documento_id: 'doc-2' },
    })
  })

  it('trata un cargo desconocido como sin elegir', async () => {
    mockObtener.mockResolvedValue({ ...postulacion, cargo: 'gerente' })

    const { result } = renderHook(() => useCargo())

    await waitFor(() => expect(result.current.props).not.toBeNull())
    expect(result.current.props?.valoresIniciales.cargo).toBe('')
  })

  it('expone el error de carga', async () => {
    mockObtener.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))

    const { result } = renderHook(() => useCargo())

    await waitFor(() => expect(result.current.error?.message).toBe('Error interno'))
    expect(result.current.props).toBeNull()
  })

  it('un 422 al validar activa el bloqueo con el mensaje del servidor y relanza el error', async () => {
    mockObtener.mockResolvedValue(postulacion)
    const rechazo = new ApiError(
      422,
      'VALIDATION_ERROR',
      'NO puede continuar, por cuanto no es empleador',
    )
    mockValidar.mockRejectedValue(rechazo)
    const { result } = renderHook(() => useCargo())
    await waitFor(() => expect(result.current.props).not.toBeNull())

    await act(async () => {
      await expect(result.current.props?.onValidar('empleador')).rejects.toBe(rechazo)
    })

    expect(result.current.bloqueo).toBe('NO puede continuar, por cuanto no es empleador')
    expect(result.current.codigo).toBe('BIESS-2026-000001')
  })

  it('un 422 al guardar también activa el bloqueo', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockRejectedValue(new ApiError(422, 'VALIDATION_ERROR', 'No acredita jubilación'))
    const { result } = renderHook(() => useCargo())
    await waitFor(() => expect(result.current.props).not.toBeNull())

    await act(async () => {
      await expect(
        result.current.props?.onGuardar({
          cargo: 'jubilados',
          respaldo_organizacion: { nombre: '', documento_id: null },
        }),
      ).rejects.toBeInstanceOf(ApiError)
    })

    expect(result.current.bloqueo).toBe('No acredita jubilación')
  })

  it('un error que no es 422 no bloquea el formulario', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockValidar.mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))
    const { result } = renderHook(() => useCargo())
    await waitFor(() => expect(result.current.props).not.toBeNull())

    await act(async () => {
      await expect(result.current.props?.onValidar('afiliados')).rejects.toBeInstanceOf(ApiError)
    })

    expect(result.current.bloqueo).toBeNull()
  })

  it('un 409 vuelve a leer la postulación', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockRejectedValue(new ApiError(409, 'CONFLICT', 'No editable'))
    const { result } = renderHook(() => useCargo())
    await waitFor(() => expect(result.current.props).not.toBeNull())
    mockObtener.mockClear()
    mockObtener.mockResolvedValue({ ...postulacion, estado: 'cerrada' })

    await act(async () => {
      await expect(
        result.current.props?.onGuardar({
          cargo: 'afiliados',
          respaldo_organizacion: { nombre: '', documento_id: null },
        }),
      ).rejects.toBeInstanceOf(ApiError)
    })

    await waitFor(() => expect(result.current.bloqueo).not.toBeNull())
    expect(mockObtener).toHaveBeenCalled()
  })

  it('bloquea a una persona no ecuatoriana sin esperar al primer guardado', async () => {
    mockObtener.mockResolvedValue({
      ...postulacion,
      informacion_personal: { nacionalidad: 'Colombiana' },
    })

    const { result } = renderHook(() => useCargo())

    await waitFor(() => expect(result.current.bloqueo).toMatch(/nacionalidad ecuatoriana/))
  })
})
