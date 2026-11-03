import { useCallback, useEffect, useRef, useState } from 'react'

// "mm:ss" (o "h:mm:ss" si pasa de una hora) para el reloj del examen.
export function formatoReloj(segundos: number): string {
  const total = Math.max(0, Math.floor(segundos))
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const mm = String(m).padStart(2, '0')
  const ss = String(s).padStart(2, '0')
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`
}

// Reloj en cuenta regresiva. El servidor manda el tiempo restante (la autoridad del tiempo) y
// el reloj lo cuenta con un reloj monotónico (performance.now), que no cambia si el postulante
// mueve la hora de su equipo. Cada respuesta del servidor lo vuelve a sincronizar. Al llegar
// a 0 llama una sola vez a `alAgotarse`.
export function useReloj(alAgotarse: () => void) {
  const [restante, setRestante] = useState<number | null>(null)
  const limite = useRef<number | null>(null) // performance.now() en que se agota
  const avisado = useRef(false)
  const avisar = useRef(alAgotarse)
  useEffect(() => {
    avisar.current = alAgotarse
  })

  const sincronizar = useCallback((segundos: number) => {
    limite.current = performance.now() + segundos * 1000
    avisado.current = false
    setRestante(Math.max(0, Math.ceil(segundos)))
  }, [])

  const detener = useCallback(() => {
    limite.current = null
    setRestante(null)
  }, [])

  useEffect(() => {
    const id = setInterval(() => {
      if (limite.current === null) return
      const segundos = Math.max(0, Math.ceil((limite.current - performance.now()) / 1000))
      setRestante(segundos)
      if (segundos === 0 && !avisado.current) {
        avisado.current = true
        avisar.current()
      }
    }, 250)
    return () => clearInterval(id)
  }, [])

  return { restante, sincronizar, detener }
}
