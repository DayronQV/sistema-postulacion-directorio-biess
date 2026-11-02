import {
  CARGOS,
  ETIQUETA_CARGO,
  MAX_NOMBRE_ORGANIZACION,
  MSG_CARGO_VALIDO,
  PREGUNTA_RESPALDO,
  erroresDeCargo,
  etiquetaCampoCargo,
  esCargo,
} from './reglasCargo'

describe('textos del cargo', () => {
  it('ofrece los tres cargos con su texto completo', () => {
    expect(CARGOS).toEqual(['afiliados', 'jubilados', 'empleador'])
    expect(ETIQUETA_CARGO.afiliados).toBe(
      'Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS',
    )
    expect(ETIQUETA_CARGO.jubilados).toBe(
      'Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS',
    )
    expect(ETIQUETA_CARGO.empleador).toBe(
      'Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS',
    )
  })

  it('usa el resultado válido del requerimiento en cada cargo', () => {
    expect(MSG_CARGO_VALIDO).toEqual({
      afiliados: 'Postulante afiliado al IESS',
      jubilados: 'Postulante acredita jubilación',
      empleador: 'Postulante es empleador',
    })
  })

  it('cambia la organización de la pregunta de respaldo según el cargo', () => {
    expect(PREGUNTA_RESPALDO.afiliados).toContain('organización de trabajadores')
    expect(PREGUNTA_RESPALDO.jubilados).toContain('organización de jubilados')
    expect(PREGUNTA_RESPALDO.empleador).toContain('federación de cámara')
  })
})

describe('esCargo', () => {
  it('reconoce solo los tres cargos', () => {
    expect(['afiliados', 'jubilados', 'empleador'].every(esCargo)).toBe(true)
    expect(esCargo('')).toBe(false)
    expect(esCargo('gerente')).toBe(false)
    expect(esCargo('Afiliados')).toBe(false)
  })
})

describe('erroresDeCargo', () => {
  const base = {
    cargoValidado: 'afiliados' as const,
    tieneRespaldo: false,
    nombre: '',
    tieneDocumento: false,
  }

  it('exige un cargo validado', () => {
    expect(erroresDeCargo({ ...base, cargoValidado: null }).cargo).toBeTruthy()
  })

  it('exige responder la pregunta de respaldo', () => {
    expect(erroresDeCargo({ ...base, tieneRespaldo: null }).respaldo).toBeTruthy()
  })

  it('con respaldo No no pide nada más', () => {
    expect(erroresDeCargo({ ...base, nombre: '', tieneDocumento: false })).toEqual({})
  })

  it('con respaldo Sí exige el nombre y el documento', () => {
    expect(erroresDeCargo({ ...base, tieneRespaldo: true })).toEqual({
      nombre: 'Campo obligatorio',
      documento: 'Cargue el documento de respaldo',
    })
    expect(
      erroresDeCargo({ ...base, tieneRespaldo: true, nombre: '   ', tieneDocumento: true }),
    ).toEqual({ nombre: 'Campo obligatorio' })
    expect(
      erroresDeCargo({ ...base, tieneRespaldo: true, nombre: 'Federación', tieneDocumento: true }),
    ).toEqual({})
  })

  it('limita la longitud del nombre', () => {
    const largo = 'a'.repeat(MAX_NOMBRE_ORGANIZACION + 1)
    expect(
      erroresDeCargo({ ...base, tieneRespaldo: true, nombre: largo, tieneDocumento: true }).nombre,
    ).toMatch(/Máximo/)
  })
})

describe('etiquetaCampoCargo', () => {
  it('traduce las claves del servidor a un texto legible', () => {
    expect(etiquetaCampoCargo('respaldo_organizacion.nombre')).toBe('Nombre de la organización')
    expect(etiquetaCampoCargo('respaldo_organizacion.documento_id')).toBe('Documento de respaldo')
    expect(etiquetaCampoCargo('cargo')).toBe('Cargo')
    expect(etiquetaCampoCargo('otro.campo')).toBe('otro.campo')
  })
})
