import { useEffect, useRef } from 'react'
import { mostrarToast } from '@/store/toasts'

// Muestra un toast cuando `guardado` pasa de falso a verdadero, es decir, cada vez que un
// guardado termina bien. Solo reacciona al cambio: si el componente ya nace con `guardado`
// verdadero (o lo sigue estando), no repite el aviso.
export function useAvisoDeGuardado(guardado: boolean, mensaje: string) {
  const anterior = useRef(guardado)
  useEffect(() => {
    if (guardado && !anterior.current) mostrarToast(mensaje)
    anterior.current = guardado
  }, [guardado, mensaje])
}
