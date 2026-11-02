import { renderHook, waitFor } from '@testing-library/react'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento } from '@/services/documentos'
import {
  guardarPostulacion,
  obtenerPostulacion,
  type FormacionAcademica,
  type Postulacion,
} from '@/services/postulacion'
import { guardarFormacion, useFormacion } from './useFormacion'

vi.mock('@/services/postulacion', () => ({
  obtenerPostulacion: vi.fn(),
  guardarPostulacion: vi.fn(),
}))
vi.mock('@/services/documentos', () => ({ eliminarDocumento: vi.fn() }))

const mockObtener = obtenerPostulacion as Mock
const mockGuardar = guardarPostulacion as Mock
const mockEliminar = eliminarDocumento as Mock

const formacion: FormacionAcademica = {
  nivel: 'tercer_nivel',
  institucion: 'UPSE',
  tipo: 'nacional',
  titulo: 'Economista',
  area: 'economia',
  senescyt_numero: '1001',
  senescyt_fecha: '2020-01-31',
  documento_id: 'doc-1',
}

const postulacion = {
  id: 'p-1',
  estado: 'borrador',
  familiares: [{ nombres_apellidos: 'Luis' }],
  referencias: [{ nombres_apellidos: 'Ana' }],
  formacion: [],
  capacitaciones: [],
  cargo: 'afiliados',
} as unknown as Postulacion

beforeEach(() => {
  mockObtener.mockReset()
  mockGuardar.mockReset()
  mockEliminar.mockReset()
})

describe('guardarFormacion', () => {
  it('envía el formulario completo cambiando solo la formación', async () => {
    mockObtener.mockResolvedValue(postulacion)
    mockGuardar.mockResolvedValue(postulacion)
    await guardarFormacion([formacion])
    expect(mockGuardar).toHaveBeenCalledWith({ ...postulacion, formacion: [formacion] })
  })

  it('elimina después del PUT los respaldos que quedaron sin referencia', async () => {
    mockObtener.mockResolvedValue({
      ...postulacion,
      formacion: [
        { ...formacion, documento_id: 'doc-viejo' },
        { ...formacion, documento_id: 'doc-conservado' },
      ],
    })
    mockGuardar.mockResolvedValue(postulacion)
    mockEliminar.mockResolvedValue(undefined)

    await guardarFormacion([{ ...formacion, documento_id: 'doc-conservado' }])

    expect(mockGuardar).toHaveBeenCalled()
    expect(mockEliminar).toHaveBeenCalledTimes(1)
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-viejo')
  })
})

describe('useFormacion', () => {
  it('carga la postulación y entrega la formación al componente', async () => {
    mockObtener.mockResolvedValue({ ...postulacion, formacion: [formacion] })
    const { result } = renderHook(() => useFormacion())
    await waitFor(() => expect(result.current.props).not.toBeNull())
    expect(result.current.props?.valor).toEqual([formacion])
    expect(result.current.props?.onGuardar).toBe(guardarFormacion)
  })
})
