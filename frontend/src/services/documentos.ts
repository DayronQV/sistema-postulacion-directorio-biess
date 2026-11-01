import { http } from './http'

export type TipoDocumento =
  'organizacion' | 'formacion' | 'experiencia' | 'publicacion' | 'capacitacion'

export interface DocumentoSubido {
  documento_id: string
  nombre: string
  tamano_bytes: number
}

export async function subirDocumento(
  signal: AbortSignal,
  tipo: TipoDocumento,
  archivo: File,
  alProgresar?: (porcentaje: number) => void,
) {
  const formulario = new FormData()
  formulario.append('tipo', tipo)
  formulario.append('archivo', archivo)
  const { data } = await http.post<DocumentoSubido>('/postulacion/documentos', formulario, {
    signal,
    onUploadProgress: ({ loaded, total }) => {
      if (total && alProgresar) alProgresar(Math.min(100, Math.round((loaded * 100) / total)))
    },
  })
  return data
}

export async function descargarDocumento(signal: AbortSignal, documentoId: string) {
  const { data } = await http.get<Blob>(`/postulacion/documentos/${documentoId}`, {
    signal,
    responseType: 'blob',
  })
  return data
}

export async function eliminarDocumento(signal: AbortSignal, documentoId: string) {
  await http.delete(`/postulacion/documentos/${documentoId}`, { signal })
}
