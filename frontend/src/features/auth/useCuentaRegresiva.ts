import { useCallback, useEffect, useState } from 'react'

// Duración del bloqueo si el servidor no informa el tiempo restante (docs/arquitectura.md).
export const BLOQUEO_POR_DEFECTO_S = 600

// Cuenta regresiva del bloqueo por intentos fallidos. `iniciar(segundos)` la arranca con
// el retry_after_seconds del servidor; `restante` llega a 0 al terminar. Se calcula
// contra la hora de fin, así no se atrasa si el navegador demora los intervalos.
export function useCuentaRegresiva() {
  const [fin, setFin] = useState<number | null>(null)
  const [restante, setRestante] = useState(0)

  const iniciar = useCallback((segundos: number) => {
    const s = Math.max(0, Math.ceil(segundos))
    setRestante(s)
    setFin(s > 0 ? Date.now() + s * 1000 : null)
  }, [])

  useEffect(() => {
    if (fin === null) return
    const id = setInterval(() => {
      const s = Math.max(0, Math.ceil((fin - Date.now()) / 1000))
      setRestante(s)
      if (s === 0) setFin(null)
    }, 250)
    return () => clearInterval(id)
  }, [fin])

  return { restante, activa: restante > 0, iniciar }
}

// 540 → "09:00"
export function formatoMinutos(segundos: number): string {
  const m = Math.floor(segundos / 60)
  const s = segundos % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
