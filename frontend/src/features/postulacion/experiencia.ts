import type { Duracion } from '@/services/postulacion'

function diasDelMes(anio: number, mes: number) {
  return new Date(Date.UTC(anio, mes, 0)).getUTCDate()
}

export function calcularDuracion(inicio: string, fin: string): Duracion {
  const [ai, mi, di] = inicio.split('-').map(Number)
  const [af, mf, df] = fin.split('-').map(Number)
  if (!ai || !mi || !di || !af || !mf || !df || fin < inicio) return { anios: 0, meses: 0, dias: 0 }
  let anios = af - ai
  let meses = mf - mi
  let dias = df - di
  if (dias < 0) {
    dias += diasDelMes(af, mf - 1 || 12)
    meses--
  }
  if (meses < 0) {
    meses += 12
    anios--
  }
  return { anios, meses, dias }
}

export function sumarDuraciones(duraciones: Duracion[]): Duracion {
  const total = duraciones.reduce(
    (suma, actual) => ({
      anios: suma.anios + actual.anios,
      meses: suma.meses + actual.meses,
      dias: suma.dias + actual.dias,
    }),
    { anios: 0, meses: 0, dias: 0 },
  )
  total.meses += Math.floor(total.dias / 30)
  total.dias %= 30
  total.anios += Math.floor(total.meses / 12)
  total.meses %= 12
  return total
}

// "1 año" pero "0 años" y "2 años".
const plural = (n: number, uno: string, varios: string) => `${n} ${n === 1 ? uno : varios}`

export function textoDuracion({ anios, meses, dias }: Duracion) {
  return `${plural(anios, 'año', 'años')}, ${plural(meses, 'mes', 'meses')} y ${plural(dias, 'día', 'días')}`
}
