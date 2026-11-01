import type { Domicilio as DomicilioApi } from '@/services/postulacion'
import {
  DOMICILIO_VACIO,
  crearEsquemaDomicilio,
  domicilioParaApi,
  valoresDesdeApi,
  type ValoresDomicilio,
} from './domicilio'

const ecuador: ValoresDomicilio = {
  ...DOMICILIO_VACIO,
  paisId: '52',
  provinciaId: '17',
  cantonId: '170',
  telefonoCelular: '0999999999',
}
const colombia: ValoresDomicilio = {
  ...DOMICILIO_VACIO,
  paisId: '40',
  provinciaTexto: 'Antioquia',
  ciudadTexto: 'Medellín',
  telefonoCelular: '0999999999',
}

function errores(esEcuador: boolean, valores: ValoresDomicilio) {
  const r = crearEsquemaDomicilio(esEcuador).safeParse(valores)
  if (r.success) return {}
  // Como react-hook-form: de cada campo se muestra el primer error.
  const porCampo: Record<string, string> = {}
  for (const i of r.error.issues) porCampo[String(i.path[0])] ??= i.message
  return porCampo
}

describe('esquema del domicilio', () => {
  it('acepta Ecuador con provincia y cantón, y otro país con texto', () => {
    expect(errores(true, ecuador)).toEqual({})
    expect(errores(false, colombia)).toEqual({})
  })

  it('exige país y celular', () => {
    expect(errores(false, DOMICILIO_VACIO)).toEqual({
      paisId: 'Seleccione el país',
      telefonoCelular: 'Ingrese su teléfono celular',
    })
  })

  it('con Ecuador exige provincia y cantón por id', () => {
    const e = errores(true, { ...ecuador, provinciaId: '', cantonId: '' })
    expect(e).toEqual({ provinciaId: 'Campo obligatorio', cantonId: 'Campo obligatorio' })
  })

  it('con otro país exige provincia y ciudad en texto', () => {
    const e = errores(false, { ...colombia, provinciaTexto: ' ', ciudadTexto: '' })
    expect(e).toEqual({ provinciaTexto: 'Campo obligatorio', ciudadTexto: 'Campo obligatorio' })
  })

  it.each(['0899999999', '099999999', '09999999999', '09abcdefgh', '+593999999999', '09 9999999'])(
    'rechaza el celular %s',
    (celular) => {
      expect(errores(true, { ...ecuador, telefonoCelular: celular })).toEqual({
        telefonoCelular: 'El celular debe tener 10 dígitos y empezar con 09',
      })
    },
  )

  it('ignora los espacios alrededor del celular', () => {
    expect(errores(true, { ...ecuador, telefonoCelular: ' 0999999999 ' })).toEqual({})
  })

  it('el correo 2 es opcional pero debe ser válido', () => {
    expect(errores(true, { ...ecuador, emailSecundario: '' })).toEqual({})
    expect(errores(true, { ...ecuador, emailSecundario: 'ana@correo.com' })).toEqual({})
    for (const malo of ['ana', 'ana@', 'ana@correo', 'Ana <ana@correo.com>']) {
      expect(errores(true, { ...ecuador, emailSecundario: malo })).toEqual({
        emailSecundario: 'Correo electrónico no válido',
      })
    }
  })

  it('valida el teléfono de domicilio y las longitudes máximas', () => {
    expect(errores(true, { ...ecuador, telefonoDomicilio: '022345678' })).toEqual({})
    expect(errores(true, { ...ecuador, telefonoDomicilio: '02-23' })).toHaveProperty(
      'telefonoDomicilio',
    )
    expect(errores(true, { ...ecuador, callePrincipal: 'a'.repeat(101) })).toEqual({
      callePrincipal: 'Máximo 100 caracteres',
    })
    expect(errores(true, { ...ecuador, numero: '1'.repeat(21) })).toEqual({
      numero: 'Máximo 20 caracteres',
    })
    expect(errores(true, { ...ecuador, referencia: 'a'.repeat(201) })).toEqual({
      referencia: 'Máximo 200 caracteres',
    })
  })
})

describe('conversión entre la API y el formulario', () => {
  const api: DomicilioApi = {
    pais_id: 52,
    provincia_id: 17,
    canton_id: 170,
    provincia_texto: '',
    ciudad_texto: '',
    telefono_celular: '0999999999',
    email_secundario: 'otro@correo.com',
    calle_principal: 'Av. Amazonas',
    numero: 'N12',
    calle_secundaria: 'Colón',
    sector: 'La Mariscal',
    referencia: 'Frente al parque',
    telefono_domicilio: '022345678',
  }

  it('los ids pasan a texto y los nulos a cadena vacía', () => {
    expect(valoresDesdeApi(api)).toMatchObject({ paisId: '52', provinciaId: '17', cantonId: '170' })
    expect(
      valoresDesdeApi({ ...api, pais_id: null, provincia_id: null, canton_id: null }),
    ).toMatchObject({
      paisId: '',
      provinciaId: '',
      cantonId: '',
    })
  })

  it('ida y vuelta conserva todos los campos', () => {
    expect(domicilioParaApi(valoresDesdeApi(api))).toEqual(api)
  })

  it('los ids sin elegir se envían como null', () => {
    expect(domicilioParaApi(colombia)).toMatchObject({
      pais_id: 40,
      provincia_id: null,
      canton_id: null,
      provincia_texto: 'Antioquia',
      ciudad_texto: 'Medellín',
    })
  })
})
