import { cedulaValida } from './cedula'
import { esquemaCuenta, esquemaIdentidad, esquemaLogin, mensajePassword } from './esquemas'
import { fortalezaPassword } from './fortaleza'

// Primer mensaje de error del campo, o undefined si el campo es válido.
function errorDe(
  resultado: { success: boolean; error?: { issues: { path: PropertyKey[]; message: string }[] } },
  campo: string,
) {
  return resultado.error?.issues.find((i) => i.path[0] === campo)?.message
}

describe('cedulaValida (módulo 10, igual que el servidor)', () => {
  it.each(['1712345675', '0912345675', '0924681356', '1802468130'])('acepta %s', (c) => {
    expect(cedulaValida(c)).toBe(true)
  })

  it.each([
    ['dígito verificador incorrecto', '1712345676'],
    ['provincia 25', '2512345675'],
    ['provincia 00', '0012345675'],
    ['tercer dígito 6', '1762345675'],
    ['9 dígitos', '171234567'],
    ['letras', '17123456a5'],
  ])('rechaza %s', (_, c) => {
    expect(cedulaValida(c)).toBe(false)
  })
})

describe('esquemaIdentidad', () => {
  const base = {
    cedula: '1712345675',
    dato: 'fecha' as const,
    fechaExpedicion: '2019-05-20',
    codigoDactilar: '',
  }

  it('acepta cédula con fecha o con código dactilar', () => {
    expect(esquemaIdentidad.safeParse(base).success).toBe(true)
    expect(
      esquemaIdentidad.safeParse({
        ...base,
        dato: 'codigo',
        fechaExpedicion: '',
        codigoDactilar: 'v1234v5678',
      }).success,
    ).toBe(true)
  })

  it('pide una cédula válida', () => {
    expect(errorDe(esquemaIdentidad.safeParse({ ...base, cedula: '' }), 'cedula')).toBe(
      'Ingrese su número de cédula',
    )
    expect(errorDe(esquemaIdentidad.safeParse({ ...base, cedula: '1234567890' }), 'cedula')).toBe(
      'Cédula no válida',
    )
  })

  it('exige el dato elegido y su formato', () => {
    expect(
      errorDe(esquemaIdentidad.safeParse({ ...base, fechaExpedicion: '' }), 'fechaExpedicion'),
    ).toBeDefined()
    expect(
      errorDe(
        esquemaIdentidad.safeParse({ ...base, fechaExpedicion: '2999-01-01' }),
        'fechaExpedicion',
      ),
    ).toBe('La fecha no puede ser futura')
    const codigo = { ...base, dato: 'codigo' as const, fechaExpedicion: '' }
    expect(
      errorDe(esquemaIdentidad.safeParse({ ...codigo, codigoDactilar: 'V1234' }), 'codigoDactilar'),
    ).toBeDefined()
    expect(
      errorDe(
        esquemaIdentidad.safeParse({ ...codigo, codigoDactilar: 'V1234-5678' }),
        'codigoDactilar',
      ),
    ).toBeDefined()
  })
})

describe('esquemaCuenta', () => {
  const base = { email: 'ana@correo.com', password: 'clave1234', confirmacion: 'clave1234' }

  it('acepta datos válidos', () => {
    expect(esquemaCuenta.safeParse(base).success).toBe(true)
  })

  it('valida el correo', () => {
    expect(errorDe(esquemaCuenta.safeParse({ ...base, email: '' }), 'email')).toBe(
      'Ingrese su correo electrónico',
    )
    expect(errorDe(esquemaCuenta.safeParse({ ...base, email: 'ana.correo.com' }), 'email')).toBe(
      'Correo electrónico no válido',
    )
  })

  it('exige que la confirmación coincida', () => {
    expect(
      errorDe(esquemaCuenta.safeParse({ ...base, confirmacion: 'otra1234' }), 'confirmacion'),
    ).toBe('Las contraseñas no coinciden')
  })
})

describe('mensajePassword (misma regla que el servidor)', () => {
  it.each([
    ['corta', 'abc123'],
    ['corta con tildes: 5 caracteres y 9 bytes', 'áááá1'],
    ['sin números', 'clavesegura'],
    ['sin letras', '12345678'],
    ['más de 72 bytes', 'a1'.repeat(37)],
  ])('rechaza la contraseña %s', (_, p) => {
    expect(mensajePassword(p)).toBeDefined()
  })

  it('acepta 8 caracteres con tildes aunque ocupen más bytes', () => {
    expect(mensajePassword('ñandú123')).toBeUndefined()
  })
})

describe('fortalezaPassword', () => {
  it('va de débil a fuerte', () => {
    expect(fortalezaPassword('').nivel).toBe(0)
    expect(fortalezaPassword('abc').etiqueta).toBe('Débil')
    expect(fortalezaPassword('clave1234').etiqueta).toBe('Débil')
    expect(fortalezaPassword('Clave12345').etiqueta).toBe('Aceptable')
    expect(fortalezaPassword('Clave-Segura-2026').etiqueta).toBe('Fuerte')
  })
})

describe('esquemaLogin', () => {
  it('pide cédula válida y contraseña, sin la regla de formato del registro', () => {
    expect(esquemaLogin.safeParse({ cedula: '1712345675', password: 'x' }).success).toBe(true)
    expect(
      errorDe(esquemaLogin.safeParse({ cedula: '1712345675', password: '' }), 'password'),
    ).toBe('Ingrese su contraseña')
    expect(errorDe(esquemaLogin.safeParse({ cedula: '123', password: 'x' }), 'cedula')).toBe(
      'Cédula no válida',
    )
  })
})
