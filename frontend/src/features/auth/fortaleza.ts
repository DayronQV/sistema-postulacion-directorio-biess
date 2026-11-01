import { MIN_PASSWORD } from './esquemas'

export interface Fortaleza {
  nivel: 0 | 1 | 2 | 3 // 0 vacía, 1 débil, 2 aceptable, 3 fuerte
  etiqueta: string
}

// Estimación simple para el indicador visual: no reemplaza la regla del servidor.
export function fortalezaPassword(p: string): Fortaleza {
  if (!p) return { nivel: 0, etiqueta: '' }
  let puntos = 0
  if ([...p].length >= MIN_PASSWORD) puntos++
  if ([...p].length >= 12) puntos++
  if (/\p{Ll}/u.test(p) && /\p{Lu}/u.test(p)) puntos++
  if (/\p{N}/u.test(p)) puntos++
  if (/[^\p{L}\p{N}]/u.test(p)) puntos++

  const cumpleMinimo = [...p].length >= MIN_PASSWORD && /\p{L}/u.test(p) && /\p{N}/u.test(p)
  if (!cumpleMinimo || puntos <= 2) return { nivel: 1, etiqueta: 'Débil' }
  if (puntos <= 3) return { nivel: 2, etiqueta: 'Aceptable' }
  return { nivel: 3, etiqueta: 'Fuerte' }
}
