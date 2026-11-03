import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { ApiError } from '@/services/errores'
import { eliminarDocumento, subirDocumento, type DocumentoSubido } from '@/services/documentos'
import DocumentoUpload from './DocumentoUpload'

vi.mock('@/services/documentos', () => ({
  subirDocumento: vi.fn(),
  descargarDocumento: vi.fn(),
  eliminarDocumento: vi.fn(),
}))

const mockSubir = subirDocumento as Mock
const mockEliminar = eliminarDocumento as Mock

const anterior: DocumentoSubido = { documento_id: 'd-viejo', nombre: 'viejo.pdf', tamano_bytes: 10 }
const nuevo: DocumentoSubido = { documento_id: 'd-nuevo', nombre: 'nuevo.pdf', tamano_bytes: 20 }

function elegirPdf() {
  const entrada = document.querySelector('input[type="file"]') as HTMLInputElement
  return userEvent.upload(entrada, new File(['%PDF'], 'nuevo.pdf', { type: 'application/pdf' }))
}

beforeEach(() => {
  mockSubir.mockReset()
  mockEliminar.mockReset()
})

describe('DocumentoUpload', () => {
  it('al reemplazar borra el PDF anterior para que no quede huérfano', async () => {
    mockSubir.mockResolvedValue(nuevo)
    mockEliminar.mockResolvedValue(undefined)
    const alCambiar = vi.fn()
    render(<DocumentoUpload tipo="formacion" maxMb={2} valor={anterior} onChange={alCambiar} />)

    await elegirPdf()

    await waitFor(() => expect(alCambiar).toHaveBeenCalledWith(nuevo))
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'd-viejo')
  })

  it('si el anterior está en uso (409) conserva el nuevo sin mostrar error', async () => {
    mockSubir.mockResolvedValue(nuevo)
    mockEliminar.mockRejectedValue(new ApiError(409, 'CONFLICT', 'En uso'))
    const alCambiar = vi.fn()
    render(<DocumentoUpload tipo="formacion" maxMb={2} valor={anterior} onChange={alCambiar} />)

    await elegirPdf()

    await waitFor(() => expect(mockEliminar).toHaveBeenCalled())
    expect(alCambiar).toHaveBeenCalledWith(nuevo)
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('muestra el mensaje del servidor si la subida falla', async () => {
    mockSubir.mockRejectedValue(
      new ApiError(422, 'VALIDATION_ERROR', 'Se superaría el límite total de 50 MB'),
    )
    render(<DocumentoUpload tipo="formacion" maxMb={2} valor={null} onChange={() => undefined} />)

    await elegirPdf()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Se superaría el límite total de 50 MB',
    )
    expect(mockEliminar).not.toHaveBeenCalled()
  })
})

describe('DocumentoUpload con eliminarEnServidor={false}', () => {
  it('Quitar solo limpia el valor, sin borrar en el servidor', async () => {
    const alCambiar = vi.fn()
    render(
      <DocumentoUpload
        tipo="formacion"
        maxMb={2}
        valor={anterior}
        onChange={alCambiar}
        eliminarEnServidor={false}
      />,
    )

    await userEvent.click(screen.getByRole('button', { name: 'Quitar archivo' }))

    expect(alCambiar).toHaveBeenCalledWith(null)
    expect(mockEliminar).not.toHaveBeenCalled()
  })

  it('Reemplazar no borra el PDF anterior', async () => {
    mockSubir.mockResolvedValue(nuevo)
    const alCambiar = vi.fn()
    render(
      <DocumentoUpload
        tipo="formacion"
        maxMb={2}
        valor={anterior}
        onChange={alCambiar}
        eliminarEnServidor={false}
      />,
    )

    await elegirPdf()

    await waitFor(() => expect(alCambiar).toHaveBeenCalledWith(nuevo))
    expect(mockEliminar).not.toHaveBeenCalled()
  })
})
