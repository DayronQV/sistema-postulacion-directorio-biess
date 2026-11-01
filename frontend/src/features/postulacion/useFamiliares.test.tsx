import { renderHook, waitFor } from '@testing-library/react'
import { vi, type Mock } from 'vitest'
import { ApiError } from '@/services/errores'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type Familiar,
  type Postulacion,
} from '@/services/postulacion'
import { guardarFamiliares, useFamiliares } from './useFamiliares'

vi.mock('@/services/postulacion', () => ({
  obtenerPostulacion: vi.fn(),
  guardarPostulacion: vi.fn(),
}))

const mockObtener = obtenerPostulacion as Mock
const mockGuardar = guardarPostulacion as Mock

const ana: Familiar = {
  nombres_apellidos: 'Ana Pérez',
  parentesco: 'hermano/a',
  institucion: 'IESS',
  area: 'Finanzas',
  cargo: 'Analista',
}

// Postulación con otras secciones llenas: guardar familiares no debe perderlas.
const postulacion = {
  id: 'p-1',
  codigo: 'BIESS-2026-000001',
  estado: 'borrador',
  informacion_personal: { nombres: 'Ana' },
  domicilio: { pais_id: 1, provincia_id: 17, canton_id: 170, calle_principal: 'Av. Amazonas' },
  tiene_familiares: false,
  familiares: [],
  cargo: 'afiliados',
  respaldo_organizacion: { nombre: 'Asociación', documento_id: null },
  tiene_publicaciones: true,
  consentimiento: true,
  updated_at: '2026-11-01T10:00:00Z',
} as unknown as Postulacion

afterEach(() => {
  vi.clearAllMocks()
})

describe('guardarFamiliares', () => {
  it('lee la postulación y envía el formulario ENTERO con solo los familiares cambiados', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockResolvedValue(postulacion)

    await guardarFamiliares({ tiene_familiares: true, familiares: [ana] })

    expect(mockObtener).toHaveBeenCalledTimes(1)
    expect(mockGuardar).toHaveBeenCalledTimes(1)
    expect(mockGuardar).toHaveBeenCalledWith({
      ...postulacion,
      tiene_familiares: true,
      familiares: [ana],
    })
    const enviado = mockGuardar.mock.calls[0][0]
    expect(enviado.domicilio).toEqual(postulacion.domicilio)
    expect(enviado.cargo).toBe('afiliados')
    expect(enviado.consentimiento).toBe(true)
  })

  it('lee la postulación en cada guardado para no pisar a otras secciones', async () => {
    mockObtener
      .mockResolvedValueOnce(postulacion)
      .mockResolvedValueOnce({ ...postulacion, cargo: 'jubilados' })
    mockGuardar.mockResolvedValue(postulacion)

    await guardarFamiliares({ tiene_familiares: false, familiares: [] })
    await guardarFamiliares({ tiene_familiares: false, familiares: [] })

    expect(mockGuardar.mock.calls[0][0].cargo).toBe('afiliados')
    expect(mockGuardar.mock.calls[1][0].cargo).toBe('jubilados')
  })

  it('no guarda nada si la lectura falla y propaga el error', async () => {
    const fallo = new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.')
    mockObtener.mockRejectedValue(fallo)

    await expect(guardarFamiliares({ tiene_familiares: false, familiares: [] })).rejects.toBe(fallo)
    expect(mockGuardar).not.toHaveBeenCalled()
  })

  it('propaga el error de validación del servidor', async () => {
    const fallo = new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
      'familiares[0].area': 'Campo obligatorio',
    })
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockRejectedValue(fallo)

    await expect(guardarFamiliares({ tiene_familiares: true, familiares: [ana] })).rejects.toBe(
      fallo,
    )
  })
})

describe('useFamiliares', () => {
  it('carga la postulación y entrega los valores iniciales', async () => {
    mockObtener.mockResolvedValue({ ...postulacion, tiene_familiares: true, familiares: [ana] })
    const { result } = renderHook(() => useFamiliares())

    expect(result.current.cargando).toBe(true)
    expect(result.current.props).toBeNull()
    await waitFor(() => expect(result.current.props).not.toBeNull())

    expect(result.current.cargando).toBe(false)
    expect(result.current.props?.valoresIniciales).toEqual({
      tiene_familiares: true,
      familiares: [ana],
    })
    expect(result.current.props?.onGuardar).toBe(guardarFamiliares)
  })

  it('expone el error de carga y permite reintentar', async () => {
    mockObtener
      .mockRejectedValueOnce(
        new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.'),
      )
      .mockResolvedValue(postulacion)
    const { result } = renderHook(() => useFamiliares())

    await waitFor(() =>
      expect(result.current.error?.message).toBe('No se pudo conectar con el servidor.'),
    )
    expect(result.current.props).toBeNull()

    result.current.recargar()
    await waitFor(() => expect(result.current.props).not.toBeNull())
    expect(result.current.error).toBeNull()
  })
})
