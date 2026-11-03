import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { listarCantones, listarPaises, listarProvincias } from '@/services/catalogos'
import { ApiError } from '@/services/errores'
import { DOMICILIO_VACIO, type ValoresDomicilio } from '../domicilio'
import Domicilio from './Domicilio'
import ContenedorToasts from '@/components/ui/ContenedorToasts'
import { reiniciarToasts } from '@/store/toasts'

vi.mock('@/services/catalogos')

const PAISES = [
  { id: 52, nombre: 'Ecuador' },
  { id: 40, nombre: 'Colombia' },
]
const PROVINCIAS = [
  { id: 17, nombre: 'Pichincha' },
  { id: 9, nombre: 'Guayas' },
]
const CANTONES: Record<number, { id: number; nombre: string }[]> = {
  17: [
    { id: 170, nombre: 'Quito' },
    { id: 171, nombre: 'Cayambe' },
  ],
  9: [{ id: 90, nombre: 'Guayaquil' }],
}

function diferido<T>() {
  let resolver!: (valor: T) => void
  const promesa = new Promise<T>((r) => {
    resolver = r
  })
  return { promesa, resolver }
}

beforeEach(() => {
  // El aviso de guardado es un toast: se monta su contenedor y se limpia entre pruebas.
  reiniciarToasts()
  render(<ContenedorToasts />)
  vi.mocked(listarPaises).mockResolvedValue(PAISES)
  vi.mocked(listarProvincias).mockResolvedValue(PROVINCIAS)
  vi.mocked(listarCantones).mockImplementation(
    async (_signal, provinciaId) => CANTONES[provinciaId],
  )
})

afterEach(() => {
  vi.resetAllMocks()
})

function montar(props: Partial<React.ComponentProps<typeof Domicilio>> = {}) {
  const onGuardar = props.onGuardar ?? vi.fn()
  render(
    <Domicilio
      valoresIniciales={DOMICILIO_VACIO}
      correoPrincipal="ana@correo.com"
      onGuardar={onGuardar}
      {...props}
    />,
  )
  return { user: userEvent.setup(), onGuardar }
}

const select = (etiqueta: RegExp) => screen.getByLabelText(etiqueta) as HTMLSelectElement
const pais = () => select(/^País/)
const provincia = () => select(/^Provincia/)
const canton = () => select(/^Ciudad \/ Cantón/)

async function paisListo() {
  await waitFor(() => expect(pais()).toBeInTheDocument())
}

describe('Domicilio: cascada país → provincia → cantón', () => {
  it('muestra un loader mientras carga los países y luego la lista', async () => {
    const { promesa, resolver } = diferido<typeof PAISES>()
    vi.mocked(listarPaises).mockReturnValue(promesa)
    montar()

    expect(screen.getByText('Cargando países...')).toBeInTheDocument()
    expect(screen.queryByLabelText(/^País/)).not.toBeInTheDocument()

    resolver(PAISES)
    await paisListo()
    expect(screen.queryByText('Cargando países...')).not.toBeInTheDocument()
    expect(Array.from(pais().options).map((o) => o.text)).toEqual([
      'Seleccione el país',
      'Ecuador',
      'Colombia',
    ])
  })

  it('sin país no muestra provincia ni ciudad', async () => {
    montar()
    await paisListo()

    expect(screen.queryByLabelText(/^Provincia/)).not.toBeInTheDocument()
    expect(screen.queryByLabelText(/^Ciudad/)).not.toBeInTheDocument()
    expect(listarProvincias).not.toHaveBeenCalled()
  })

  it('con Ecuador carga las provincias con su loader y deja el cantón deshabilitado', async () => {
    const { promesa, resolver } = diferido<typeof PROVINCIAS>()
    vi.mocked(listarProvincias).mockReturnValue(promesa)
    const { user } = montar()
    await paisListo()

    await user.selectOptions(pais(), 'Ecuador')

    expect(await screen.findByText('Cargando provincias...')).toBeInTheDocument()
    expect(listarProvincias).toHaveBeenCalledWith(expect.any(AbortSignal), 52)
    resolver(PROVINCIAS)
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    expect(canton()).toBeDisabled()
    expect(listarCantones).not.toHaveBeenCalled()
  })

  it('al elegir la provincia carga los cantones con su loader', async () => {
    const { promesa, resolver } = diferido<(typeof CANTONES)[17]>()
    vi.mocked(listarCantones).mockReturnValue(promesa)
    const { user } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Ecuador')
    await waitFor(() => expect(provincia()).toBeInTheDocument())

    await user.selectOptions(provincia(), 'Pichincha')

    expect(await screen.findByText('Cargando cantones...')).toBeInTheDocument()
    expect(listarCantones).toHaveBeenCalledWith(expect.any(AbortSignal), 17)
    resolver(CANTONES[17])
    await waitFor(() => expect(canton()).toBeEnabled())
    expect(Array.from(canton().options).map((o) => o.text)).toEqual([
      'Seleccione el cantón',
      'Quito',
      'Cayambe',
    ])
  })

  it('al cambiar la provincia se limpia el cantón y se cargan los de la nueva', async () => {
    const { user } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Ecuador')
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    await user.selectOptions(provincia(), 'Pichincha')
    await waitFor(() => expect(canton()).toBeEnabled())
    await user.selectOptions(canton(), 'Quito')
    expect(canton().value).toBe('170')

    await user.selectOptions(provincia(), 'Guayas')

    await waitFor(() =>
      expect(Array.from(canton().options).map((o) => o.text)).toContain('Guayaquil'),
    )
    expect(canton().value).toBe('')
    expect(Array.from(canton().options).map((o) => o.text)).not.toContain('Quito')
  })

  it('al cambiar el país se limpian provincia y cantón', async () => {
    const { user } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Ecuador')
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    await user.selectOptions(provincia(), 'Pichincha')
    await waitFor(() => expect(canton()).toBeEnabled())
    await user.selectOptions(canton(), 'Quito')

    // A otro país: texto libre, vacío.
    await user.selectOptions(pais(), 'Colombia')
    expect(screen.getByLabelText(/^Provincia \/ Estado/)).toHaveValue('')
    expect(screen.getByLabelText(/^Ciudad\*$/)).toHaveValue('')
    expect(screen.queryByLabelText(/^Ciudad \/ Cantón/)).not.toBeInTheDocument()

    // De vuelta a Ecuador: nada seleccionado.
    await user.selectOptions(pais(), 'Ecuador')
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    expect(provincia().value).toBe('')
    expect(canton()).toBeDisabled()
    expect(canton().value).toBe('')
  })

  it('al cambiar de país el texto libre anterior se borra', async () => {
    const { user } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Colombia')
    await user.type(screen.getByLabelText(/^Provincia \/ Estado/), 'Antioquia')
    await user.type(screen.getByLabelText(/^Ciudad\*$/), 'Medellín')

    await user.selectOptions(pais(), 'Ecuador')
    await user.selectOptions(pais(), 'Colombia')

    expect(screen.getByLabelText(/^Provincia \/ Estado/)).toHaveValue('')
    expect(screen.getByLabelText(/^Ciudad\*$/)).toHaveValue('')
  })

  it('abre con los valores guardados, sin borrarlos al cargar las listas', async () => {
    montar({
      valoresIniciales: {
        ...DOMICILIO_VACIO,
        paisId: '52',
        provinciaId: '17',
        cantonId: '171',
        telefonoCelular: '0999999999',
      },
    })

    await waitFor(() => expect(canton()).toBeEnabled())
    await waitFor(() => expect(canton().value).toBe('171'))
    expect(pais().value).toBe('52')
    expect(provincia().value).toBe('17')
    expect(listarCantones).toHaveBeenCalledWith(expect.any(AbortSignal), 17)
  })

  it('ofrece reintentar cuando falla una lista', async () => {
    vi.mocked(listarProvincias).mockRejectedValueOnce(
      new ApiError(0, 'NETWORK_ERROR', 'No se pudo conectar con el servidor.'),
    )
    const { user } = montar()
    await paisListo()

    await user.selectOptions(pais(), 'Ecuador')

    expect(await screen.findByText(/No se pudieron cargar las provincias/)).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Reintentar' }))
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    expect(screen.queryByText(/No se pudieron cargar/)).not.toBeInTheDocument()
  })

  it('muestra el error si falla la carga de los países', async () => {
    vi.mocked(listarPaises).mockRejectedValue(new ApiError(500, 'INTERNAL_ERROR', 'Error interno'))
    montar()

    expect(await screen.findByText(/No se pudieron cargar los países/)).toBeInTheDocument()
  })
})

describe('Domicilio: contacto y validaciones', () => {
  it('muestra el correo 1 precargado y de solo lectura', async () => {
    montar()
    await paisListo()

    const correo1 = screen.getByLabelText(/^Correo electrónico 1/)
    expect(correo1).toHaveValue('ana@correo.com')
    expect(correo1).toHaveAttribute('readonly')
  })

  it('no guarda y avisa qué falta cuando el formulario está vacío', async () => {
    const { user, onGuardar } = montar()
    await paisListo()

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Seleccione el país', { selector: 'p' })).toBeInTheDocument()
    expect(screen.getByText('Ingrese su teléfono celular')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it.each(['0899999999', '099999999', '09abcdefgh'])('rechaza el celular %s', async (celular) => {
    const { user, onGuardar } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Colombia')
    await user.type(screen.getByLabelText(/^Provincia \/ Estado/), 'Antioquia')
    await user.type(screen.getByLabelText(/^Ciudad\*$/), 'Medellín')
    await user.type(screen.getByLabelText(/^Teléfono celular/), celular)

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(
      await screen.findByText('El celular debe tener 10 dígitos y empezar con 09'),
    ).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('rechaza un correo 2 mal escrito', async () => {
    const { user, onGuardar } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Colombia')
    await user.type(screen.getByLabelText(/^Provincia \/ Estado/), 'Antioquia')
    await user.type(screen.getByLabelText(/^Ciudad\*$/), 'Medellín')
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')
    await user.type(screen.getByLabelText(/^Correo electrónico 2/), 'ana@correo')

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Correo electrónico no válido')).toBeInTheDocument()
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('con Ecuador exige provincia y cantón', async () => {
    const { user, onGuardar } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Ecuador')
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findAllByText('Campo obligatorio')).toHaveLength(2)
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('con otro país exige provincia y ciudad en texto', async () => {
    const { user, onGuardar } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Colombia')
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findAllByText('Campo obligatorio')).toHaveLength(2)
    expect(onGuardar).not.toHaveBeenCalled()
  })

  it('guarda un domicilio de Ecuador con los ids elegidos', async () => {
    const { user, onGuardar } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Ecuador')
    await waitFor(() => expect(provincia()).toBeInTheDocument())
    await user.selectOptions(provincia(), 'Pichincha')
    await waitFor(() => expect(canton()).toBeEnabled())
    await user.selectOptions(canton(), 'Quito')
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')
    await user.type(screen.getByLabelText(/^Correo electrónico 2/), ' otro@correo.com ')
    await user.type(screen.getByLabelText(/^Calle principal/), 'Av. Amazonas')

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    await waitFor(() => expect(onGuardar).toHaveBeenCalledTimes(1))
    expect(onGuardar).toHaveBeenCalledWith({
      ...DOMICILIO_VACIO,
      paisId: '52',
      provinciaId: '17',
      cantonId: '170',
      telefonoCelular: '0999999999',
      emailSecundario: 'otro@correo.com',
      callePrincipal: 'Av. Amazonas',
    } satisfies ValoresDomicilio)
  })

  it('guarda un domicilio de otro país con provincia y ciudad en texto', async () => {
    const { user, onGuardar } = montar()
    await paisListo()
    await user.selectOptions(pais(), 'Colombia')
    await user.type(screen.getByLabelText(/^Provincia \/ Estado/), 'Antioquia')
    await user.type(screen.getByLabelText(/^Ciudad\*$/), 'Medellín')
    await user.type(screen.getByLabelText(/^Teléfono celular/), '0999999999')

    await user.click(screen.getByRole('button', { name: 'Guardar' }))

    await waitFor(() => expect(onGuardar).toHaveBeenCalledTimes(1))
    expect(onGuardar).toHaveBeenCalledWith({
      ...DOMICILIO_VACIO,
      paisId: '40',
      provinciaTexto: 'Antioquia',
      ciudadTexto: 'Medellín',
      telefonoCelular: '0999999999',
    })
  })
})

describe('Domicilio: estados de guardado', () => {
  it('muestra el estado de carga del botón mientras guarda', async () => {
    montar({ guardando: true })
    await paisListo()

    const boton = screen.getByRole('button', { name: 'Guardando...' })
    expect(boton).toBeDisabled()
    expect(boton).toHaveAttribute('aria-busy', 'true')
  })

  it('confirma el guardado con un toast, al pasar de sin guardar a guardado', async () => {
    const props = {
      valoresIniciales: DOMICILIO_VACIO,
      correoPrincipal: 'ana@correo.com',
      onGuardar: vi.fn(),
    }
    const { rerender } = render(<Domicilio {...props} />)
    await paisListo()
    expect(screen.queryByText('Domicilio guardado.')).not.toBeInTheDocument()

    rerender(<Domicilio {...props} guardado />)

    expect(screen.getByText('Domicilio guardado.')).toBeInTheDocument()
  })

  it('muestra el error del servidor y lo asocia al campo', async () => {
    const error = new ApiError(400, 'VALIDATION_ERROR', 'Datos inválidos', {
      'domicilio.canton_id': 'El cantón no pertenece a la provincia',
      'domicilio.telefono_celular': 'El celular debe tener 10 dígitos y empezar con 09',
    })
    montar({
      error,
      valoresIniciales: { ...DOMICILIO_VACIO, paisId: '52', provinciaId: '9', cantonId: '170' },
    })

    expect(await screen.findByText('Datos inválidos')).toBeInTheDocument()
    expect(await screen.findByText('El cantón no pertenece a la provincia')).toBeInTheDocument()
    expect(
      screen.getByText('El celular debe tener 10 dígitos y empezar con 09'),
    ).toBeInTheDocument()
    expect(screen.queryByText('Domicilio guardado.')).not.toBeInTheDocument()
  })
})
