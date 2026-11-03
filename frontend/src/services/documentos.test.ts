import { beforeEach, describe, expect, it, vi } from 'vitest'
import { descargarDocumento, eliminarDocumento, subirDocumento } from './documentos'
import { http } from './http'

vi.mock('./http', () => ({
  http: { post: vi.fn(), get: vi.fn(), delete: vi.fn() },
}))

describe('servicio de documentos', () => {
  beforeEach(() => vi.clearAllMocks())

  it('envía el PDF y su tipo como multipart', async () => {
    vi.mocked(http.post).mockResolvedValue({
      data: { documento_id: 'd1', nombre: 'x.pdf', tamano_bytes: 4 },
    })
    const archivo = new File(['%PDF'], 'x.pdf', { type: 'application/pdf' })
    const resultado = await subirDocumento(new AbortController().signal, 'formacion', archivo)
    const formulario = vi.mocked(http.post).mock.calls[0][1] as FormData
    expect(formulario.get('tipo')).toBe('formacion')
    expect(formulario.get('archivo')).toBe(archivo)
    expect(resultado.documento_id).toBe('d1')
  })

  it('descarga y elimina por identificador', async () => {
    const pdf = new Blob(['%PDF'], { type: 'application/pdf' })
    vi.mocked(http.get).mockResolvedValue({ data: pdf })
    vi.mocked(http.delete).mockResolvedValue({ data: undefined })
    const signal = new AbortController().signal
    expect(await descargarDocumento(signal, 'd1')).toBe(pdf)
    await eliminarDocumento(signal, 'd1')
    expect(http.get).toHaveBeenCalledWith(
      '/postulacion/documentos/d1',
      expect.objectContaining({ responseType: 'blob' }),
    )
    expect(http.delete).toHaveBeenCalledWith('/postulacion/documentos/d1', { signal })
  })
})
