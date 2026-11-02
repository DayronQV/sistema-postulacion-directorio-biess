import type { TipoCargo } from '@/services/postulacion'

// Textos y reglas de la sección "Selección del cargo" (docs/requerimiento-biess.md, 1.4.1 a
// 1.4.3). El servidor es quien valida el cargo con el Historial Laboral y hace cumplir las
// reglas; aquí se muestran y se avisa antes de enviar.

export const CARGOS: readonly TipoCargo[] = ['afiliados', 'jubilados', 'empleador']

export const ETIQUETA_CARGO: Record<TipoCargo, string> = {
  afiliados:
    'Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS',
  jubilados:
    'Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS',
  empleador:
    'Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS',
}

// Resultado cuando el postulante cumple la calidad que exige el cargo.
export const MSG_CARGO_VALIDO: Record<TipoCargo, string> = {
  afiliados: 'Postulante afiliado al IESS',
  jubilados: 'Postulante acredita jubilación',
  empleador: 'Postulante es empleador',
}

// Pregunta de respaldo: cambia la organización según el cargo.
export const PREGUNTA_RESPALDO: Record<TipoCargo, string> = {
  afiliados:
    '¿Se encuentra respaldadas por una organización de trabajadores de representación nacional legalmente constituida ante autoridad competente?',
  jubilados:
    '¿Se encuentra respaldadas por una organización de jubilados de representación nacional legalmente constituida ante autoridad competente?',
  empleador:
    '¿Se encuentra respaldadas por una federación de cámara de representación nacional legalmente constituida ante autoridad competente?',
}

export const MAX_NOMBRE_ORGANIZACION = 200

// Límite del documento de respaldo (PDF), en MB.
export const MAX_MB_RESPALDO = 4

export function esCargo(valor: string): valor is TipoCargo {
  return (CARGOS as readonly string[]).includes(valor)
}

export interface ErroresCargo {
  cargo?: string
  respaldo?: string
  nombre?: string
  documento?: string
}

interface EntradaCargo {
  // Cargo elegido y ya validado con el Historial Laboral; null si no lo hay.
  cargoValidado: TipoCargo | null
  // Respuesta a la pregunta de respaldo; null si todavía no responde.
  tieneRespaldo: boolean | null
  nombre: string
  tieneDocumento: boolean
}

// Las mismas reglas que valida el servidor (cargo_service.go): con respaldo "Sí" el nombre y
// el documento son obligatorios.
export function erroresDeCargo(e: EntradaCargo): ErroresCargo {
  const errores: ErroresCargo = {}
  if (!e.cargoValidado) {
    errores.cargo = 'Seleccione un cargo y espere el resultado de la validación'
    return errores
  }
  if (e.tieneRespaldo === null) errores.respaldo = 'Seleccione Sí o No'
  if (e.tieneRespaldo) {
    const nombre = e.nombre.trim()
    if (!nombre) errores.nombre = 'Campo obligatorio'
    else if ([...nombre].length > MAX_NOMBRE_ORGANIZACION)
      errores.nombre = `Máximo ${MAX_NOMBRE_ORGANIZACION} caracteres`
    if (!e.tieneDocumento) errores.documento = 'Cargue el documento de respaldo'
  }
  return errores
}

// Nombre legible de un campo del error 400 del servidor (docs/api.md, `fields`).
const ETIQUETA_CAMPO: Record<string, string> = {
  cargo: 'Cargo',
  'respaldo_organizacion.nombre': 'Nombre de la organización',
  'respaldo_organizacion.documento_id': 'Documento de respaldo',
}

export function etiquetaCampoCargo(campo: string): string {
  return ETIQUETA_CAMPO[campo] ?? campo
}
