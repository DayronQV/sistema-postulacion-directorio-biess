import { renderHook, waitFor } from '@testing-library/react'
import { vi, type Mock } from 'vitest'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type Postulacion,
  type ReferenciaPersonal,
} from '@/services/postulacion'
import { guardarReferencias, useReferencias } from './useReferencias'

vi.mock('@/services/postulacion', () => ({
  obtenerPostulacion: vi.fn(),
  guardarPostulacion: vi.fn(),
}))

const mockObtener = obtenerPostulacion as Mock
const mockGuardar = guardarPostulacion as Mock

const ana: ReferenciaPersonal = {
  nombres_apellidos: 'Ana Pérez',
  relacion: 'Amiga',
  telefono: '0999999999',
  correo: 'ana@example.com',
}

// Postulación con otras secciones llenas: guardar referencias no debe perderlas.
const postulacion = {
  id: 'p-1',
  estado: 'borrador',
  domicilio: { pais_id: 1, provincia_id: 17, canton_id: 170, calle_principal: 'Av. Amazonas' },
  tiene_familiares: true,
  familiares: [{ nombres_apellidos: 'Luis' }],
  referencias: [],
  cargo: 'afiliados',
} as unknown as Postulacion

beforeEach(() => {
  mockObtener.mockReset()
  mockGuardar.mockReset()
})

describe('guardarReferencias', () => {
  it('envía el formulario completo cambiando solo las referencias', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockResolvedValue(postulacion)

    await guardarReferencias([ana])

    expect(mockGuardar).toHaveBeenCalledWith({ ...postulacion, referencias: [ana] })
  })
})

describe('useReferencias', () => {
  it('carga la postulación y entrega las referencias al componente', async () => {
    mockObtener.mockResolvedValue({ ...postulacion, referencias: [ana] })

    const { result } = renderHook(() => useReferencias())

    await waitFor(() => expect(result.current.props).not.toBeNull())
    expect(result.current.props?.valor).toEqual([ana])
    expect(result.current.props?.onGuardar).toBe(guardarReferencias)
  })
})
