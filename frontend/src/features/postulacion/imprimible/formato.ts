import type { Duracion } from '@/services/postulacion'

// Convierte AAAA-MM-DD en DD/MM/AAAA sin pasar por Date (evita el desfase de zona horaria).
export function formatearFecha(iso: string | null | undefined): string {
  if (!iso) return '—'
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(iso)
  return m ? `${m[3]}/${m[2]}/${m[1]}` : iso
}

export function formatearFechaHoy(fecha: Date): string {
  const dd = String(fecha.getDate()).padStart(2, '0')
  const mm = String(fecha.getMonth() + 1).padStart(2, '0')
  return `${dd}/${mm}/${fecha.getFullYear()}`
}

function plural(n: number, singular: string, pluralTexto: string) {
  return `${n} ${n === 1 ? singular : pluralTexto}`
}

export function formatearDuracion(d: Duracion | undefined): string {
  if (!d) return '—'
  return [
    plural(d.anios, 'año', 'años'),
    plural(d.meses, 'mes', 'meses'),
    plural(d.dias, 'día', 'días'),
  ].join(', ')
}

export function siNo(v: boolean | null | undefined): string {
  if (v === true) return 'Sí'
  if (v === false) return 'No'
  return '—'
}

export function texto(v: string | number | null | undefined): string {
  return v === null || v === undefined || v === '' ? '—' : String(v)
}
