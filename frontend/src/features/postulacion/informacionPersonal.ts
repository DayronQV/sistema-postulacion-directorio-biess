import type { Postulacion } from '@/services/postulacion'

// Reglas de la sección "Información personal". El servidor es quien las hace cumplir
// (PUT /postulacion responde 422 y cierra la postulación); aquí solo se muestran.

// Texto comparable: sin tildes, en minúsculas y con los espacios normalizados.
const normalizar = (texto: string) =>
  texto.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase().split(/\s+/).filter(Boolean).join(' ')

export function esNacionalidadEcuatoriana(nacionalidad: string): boolean {
  const n = normalizar(nacionalidad)
  return n === 'ecuatoriana' || n === 'ecuatoriano'
}

// Casado/a y unión de hecho son los estados civiles que exigen los datos de la pareja.
export function estadoCivilConPareja(estadoCivil: string): boolean {
  const e = normalizar(estadoCivil)
  return e === 'casado/a' || e === 'casado' || e === 'casada' || e === 'union de hecho'
}

// AAAA-MM-DD → DD/MM/AAAA, sin pasar por Date para que la zona horaria no cambie el día.
export function formatearFecha(iso: string): string {
  const m = /^(\d{4})-(\d{2})-(\d{2})/.exec(iso)
  return m ? `${m[3]}/${m[2]}/${m[1]}` : iso
}

export const MSG_NACIONALIDAD = 'Solo pueden postularse personas de nacionalidad ecuatoriana.'

// Motivo por el que la persona no puede continuar con el formulario, o null si puede.
// Un extranjero que todavía no guardó nada sigue en borrador en el servidor, pero no puede
// continuar: se le muestra el bloqueo sin esperar al primer guardado.
export function motivoDeBloqueo(
  p: Pick<Postulacion, 'estado' | 'informacion_personal'>,
): string | null {
  const { nacionalidad } = p.informacion_personal
  if (!esNacionalidadEcuatoriana(nacionalidad)) {
    return `${MSG_NACIONALIDAD} El Registro Civil registra su nacionalidad como «${nacionalidad}», por lo que su postulación está cerrada.`
  }
  if (p.estado === 'cerrada') return 'Su postulación está cerrada y ya no se puede modificar.'
  return null
}

// Nombre corto para la barra superior: primer nombre y primer apellido
// ("Dayron Adrian" + "Quiñonez Valencia" → "Dayron Quiñonez").
export function nombreCorto(nombres: string, apellidos: string): string {
  return [nombres, apellidos]
    .map((parte) => parte.trim().split(/\s+/)[0] ?? '')
    .filter(Boolean)
    .join(' ')
}
