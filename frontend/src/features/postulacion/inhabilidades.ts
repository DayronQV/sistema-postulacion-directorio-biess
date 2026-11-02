import type { Inhabilidades } from '@/services/postulacion'

// Texto exacto del requerimiento (docs/especificacion.md, EF-12, sección 9).
export const PREGUNTAS_INHABILIDADES: string[] = [
  '¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política?',
  '¿Se encuentra inhabilitado para ejercer el comercio?',
  '¿Está en mora, directa o indirectamente, con instituciones del Estado o bajo control de la Superintendencia de Bancos, incluidas las off-shore?',
  '¿Mantiene vínculo, por propiedad o administración, con instituciones del sistema financiero y de seguros privados, bolsas de valores, casas de valores y administradoras de fondos y fideicomisos?',
  '¿Es funcionario o empleado del IESS - BIESS?',
  '¿Mantiene interés propio o representa a terceros en la propiedad, dirección o gestión de aseguradoras u otras personas del sistema nacional de seguridad social?',
  '¿En los últimos 5 años ha sido removido por algún organismo de control?',
  '¿Mantiene obligaciones pendientes con el SRI?',
  '¿Ha incurrido en los últimos 5 años en castigo de sus obligaciones por parte de una institución financiera?',
  '¿Actualmente litiga contra el IESS - BIESS?',
  '¿Está procesado, enjuiciado o condenado por delito de corrupción o crimen organizado?',
  '¿Mantiene en firme responsabilidades de la Contraloría General del Estado?',
  '¿Está registrado en la base de datos de la Unidad de Análisis Financiero y Económico?',
]

export const PREGUNTA_REMOVIDO = 7

function respuesta(inhab: Inhabilidades | undefined, n: number): boolean | undefined {
  const v = inhab?.[String(n)]
  return typeof v === 'boolean' ? v : undefined
}

// Números de las preguntas cuya respuesta inhabilita: la 1 en "No" y de la 2 a la 13 en "Sí".
// Solo es un aviso para el postulante; la elegibilidad la calcula y guarda el servidor.
export function preguntasQueInhabilitan(inhab: Inhabilidades | undefined): number[] {
  const numeros: number[] = []
  for (let n = 1; n <= PREGUNTAS_INHABILIDADES.length; n++) {
    const r = respuesta(inhab, n)
    if (r === undefined) continue
    if (n === 1 ? r === false : r === true) numeros.push(n)
  }
  return numeros
}

// Las 13 preguntas respondidas y, si la 7 es Sí, su institución y su motivo.
export function inhabilidadesCompletas(inhab: Inhabilidades | undefined): boolean {
  for (let n = 1; n <= PREGUNTAS_INHABILIDADES.length; n++) {
    if (respuesta(inhab, n) === undefined) return false
  }
  if (respuesta(inhab, PREGUNTA_REMOVIDO)) {
    const d = inhab?.removido_detalle
    return Boolean(d?.institucion?.trim() && d.motivo?.trim())
  }
  return true
}
