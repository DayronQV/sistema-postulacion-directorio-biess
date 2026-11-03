import {
  MAX_AREA,
  MAX_CARGO,
  MAX_NOMBRES,
  PARENTESCOS,
  esquemaFamiliar,
  etiquetaParentesco,
  repartirErrores,
} from './reglasFamiliares'

const valido = {
  nombres_apellidos: 'Ana Pérez',
  parentesco: 'hermano/a',
  institucion: 'IESS',
  area: 'Finanzas',
  cargo: 'Analista',
}

describe('esquemaFamiliar', () => {
  it('acepta una fila completa y recorta los espacios', () => {
    const r = esquemaFamiliar.safeParse({ ...valido, nombres_apellidos: '  Ana Pérez ' })
    expect(r.success && r.data.nombres_apellidos).toBe('Ana Pérez')
  })

  it('acepta los 15 parentescos y rechaza otros', () => {
    expect(PARENTESCOS).toHaveLength(15)
    for (const p of PARENTESCOS)
      expect(esquemaFamiliar.safeParse({ ...valido, parentesco: p }).success).toBe(true)
    expect(esquemaFamiliar.safeParse({ ...valido, parentesco: 'primo' }).success).toBe(false)
    expect(esquemaFamiliar.safeParse({ ...valido, parentesco: '' }).success).toBe(false)
  })

  it('exige la institución IESS o BIESS', () => {
    expect(esquemaFamiliar.safeParse({ ...valido, institucion: 'BIESS' }).success).toBe(true)
    for (const i of ['', 'ISSFA', 'iess']) {
      expect(esquemaFamiliar.safeParse({ ...valido, institucion: i }).success).toBe(false)
    }
  })

  it('exige nombres, área y cargo no vacíos y acotados', () => {
    const casos = [
      ['nombres_apellidos', ' ', MAX_NOMBRES],
      ['area', '', MAX_AREA],
      ['cargo', '\t', MAX_CARGO],
    ] as const
    for (const [campo, vacio, max] of casos) {
      expect(esquemaFamiliar.safeParse({ ...valido, [campo]: vacio }).success).toBe(false)
      expect(esquemaFamiliar.safeParse({ ...valido, [campo]: 'a'.repeat(max) }).success).toBe(true)
      expect(esquemaFamiliar.safeParse({ ...valido, [campo]: 'a'.repeat(max + 1) }).success).toBe(
        false,
      )
    }
  })
})

describe('etiquetaParentesco', () => {
  it('pone la primera letra en mayúscula', () => {
    expect(etiquetaParentesco('cónyuge')).toBe('Cónyuge')
    expect(etiquetaParentesco('')).toBe('')
  })
})

describe('repartirErrores', () => {
  it('separa el error de la sección, los de cada fila y los de otros campos', () => {
    const r = repartirErrores({
      familiares: 'Agregue al menos un familiar',
      'familiares[0].parentesco': 'Valor no permitido',
      'familiares[2].area': 'Campo obligatorio',
      'familiares[2].cargo': 'Campo obligatorio',
      cargo: 'Cargo no válido',
    })
    expect(r.general).toBe('Agregue al menos un familiar')
    expect(r.filas[0]).toEqual(['Parentesco: Valor no permitido'])
    expect(r.filas[2]).toEqual(['Área: Campo obligatorio', 'Cargo: Campo obligatorio'])
    expect(r.otros).toEqual(['cargo: Cargo no válido'])
  })

  it('sin errores devuelve todo vacío', () => {
    expect(repartirErrores(undefined)).toEqual({ filas: {}, otros: [] })
  })
})
