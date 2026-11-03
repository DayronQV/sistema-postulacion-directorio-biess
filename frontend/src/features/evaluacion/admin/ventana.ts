import type { Fase } from '@/services/adminFases'

export const FASE_EVALUACION = 'Evaluación'

// Minutos que dura la ventana de la fase Evaluación (de su fecha de inicio a su fecha de fin).
// Es null si falta alguna de las dos fechas (sin límite por ese lado) o no son válidas.
export function minutosDeVentana(fase: Fase | undefined): number | null {
  if (!fase?.fecha_inicio || !fase.fecha_fin) return null
  const inicio = new Date(fase.fecha_inicio).getTime()
  const fin = new Date(fase.fecha_fin).getTime()
  if (Number.isNaN(inicio) || Number.isNaN(fin) || fin < inicio) return null
  return Math.floor((fin - inicio) / 60_000)
}

// Si la ventana de la fase es más corta que la duración del examen, devuelve los minutos de la
// ventana: quien entre cerca del cierre tendrá menos tiempo (la hora de cierre de cada examen es
// la menor entre su inicio más la duración y la fecha de fin de la fase).
export function ventanaMasCortaQueElExamen(
  fases: Fase[] | null | undefined,
  duracionMinutos: number | null,
): number | null {
  if (duracionMinutos === null) return null
  const ventana = minutosDeVentana(fases?.find((f) => f.nombre === FASE_EVALUACION))
  return ventana !== null && ventana < duracionMinutos ? ventana : null
}
