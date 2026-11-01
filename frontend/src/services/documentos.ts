import { http } from './http'

export type TipoDocumento =
  'organizacion' | 'formacion' | 'experiencia' | 'publicacion' | 'capacitacion'

export interface DocumentoSubido {
  documento_id: string
  nombre: string
  tamano_bytes: number
}

// El cliente general corta a los 15 s; un PDF de 4 MB en una conexión lenta tarda más,
// así que la subida tiene su propio plazo. Se puede cancelar antes con la señal.
const PLAZO_SUBIDA_MS = 120_000

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
    timeout: PLAZO_SUBIDA_MS,
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
