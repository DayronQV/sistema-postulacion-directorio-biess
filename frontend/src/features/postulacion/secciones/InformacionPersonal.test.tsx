import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ApiError } from '@/services/errores'
import type { InformacionPersonal as InformacionApi } from '@/services/postulacion'
import InformacionPersonal from './InformacionPersonal'
import PostulacionBloqueada from './PostulacionBloqueada'
import ContenedorToasts from '@/components/ui/ContenedorToasts'
import { reiniciarToasts } from '@/store/toasts'

const soltera: InformacionApi = {
  cedula: '1712345675',
  email: 'ana@correo.com',
  nombres: 'Ana Lucía',
  apellidos: 'Pérez Mora',
  genero: 'Femenino',
  fecha_nacimiento: '1990-03-04',
  nacionalidad: 'Ecuatoriana',
  estado_civil: 'Soltero/a',
}

const casada: InformacionApi = {
  ...soltera,
  estado_civil: 'Casado/a',
  pareja: { nombres_apellidos: 'Luis Alberto Párraga Mera', cedula: '0913579249' },
}

function dato(etiqueta: string) {
  return screen.getByText(etiqueta).nextElementSibling
}

beforeEach(() => {
  // El aviso de guardado es un toast: se monta su contenedor y se limpia entre pruebas.
  reiniciarToasts()
  render(<ContenedorToasts />)
})

describe('<InformacionPersonal />', () => {
  it('muestra los datos del Registro Civil con la fecha en DD/MM/AAAA', () => {
    render(<InformacionPersonal informacion={soltera} onGuardar={vi.fn()} />)

    expect(dato('Nombres')).toHaveTextContent('Ana Lucía')
    expect(dato('Apellidos')).toHaveTextContent('Pérez Mora')
    expect(dato('Cédula')).toHaveTextContent('1712345675')
    expect(dato('Género')).toHaveTextContent('Femenino')
    expect(dato('Fecha de nacimiento')).toHaveTextContent('04/03/1990')
    expect(dato('Nacionalidad')).toHaveTextContent('Ecuatoriana')
    expect(dato('Estado civil')).toHaveTextContent('Soltero/a')
  })

  it('es de solo lectura: no ofrece campos editables', () => {
    render(<InformacionPersonal informacion={casada} onGuardar={vi.fn()} />)

    expect(screen.queryAllByRole('textbox')).toHaveLength(0)
    expect(screen.queryAllByRole('combobox')).toHaveLength(0)
  })

  it.each(['Casado/a', 'Unión de hecho'])('muestra los datos de la pareja si es %s', (estado) => {
    render(
      <InformacionPersonal informacion={{ ...casada, estado_civil: estado }} onGuardar={vi.fn()} />,
    )

    expect(screen.getByRole('heading', { name: 'Datos de su pareja' })).toBeInTheDocument()
    expect(dato('Nombres y apellidos de la pareja')).toHaveTextContent('Luis Alberto Párraga Mera')
    expect(dato('Cédula de la pareja')).toHaveTextContent('0913579249')
  })

  it.each(['Soltero/a', 'Divorciado/a', 'Viudo/a'])('oculta la pareja si es %s', (estado) => {
    render(
      <InformacionPersonal
        informacion={{ ...soltera, estado_civil: estado }}
        onGuardar={vi.fn()}
      />,
    )

    expect(screen.queryByText('Datos de su pareja')).not.toBeInTheDocument()
    expect(screen.queryByText('Cédula de la pareja')).not.toBeInTheDocument()
  })

  it('no muestra la pareja de un estado civil sin pareja aunque el servidor la envíe', () => {
    render(
      <InformacionPersonal
        informacion={{ ...casada, estado_civil: 'Divorciado/a' }}
        onGuardar={vi.fn()}
      />,
    )

    expect(screen.queryByText('Luis Alberto Párraga Mera')).not.toBeInTheDocument()
  })

  it('el botón Guardar avisa a quien monta la sección', async () => {
    const onGuardar = vi.fn()
    render(<InformacionPersonal informacion={soltera} onGuardar={onGuardar} />)

    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledTimes(1)
  })

  it('deshabilita Guardar mientras guarda', () => {
    render(<InformacionPersonal informacion={soltera} onGuardar={vi.fn()} guardando />)

    expect(screen.getByRole('button', { name: /Guardando/ })).toBeDisabled()
  })

  it('confirma el guardado con un toast, al pasar de sin guardar a guardado', () => {
    const { rerender } = render(<InformacionPersonal informacion={soltera} onGuardar={vi.fn()} />)
    expect(screen.queryByText('La información personal se guardó.')).not.toBeInTheDocument()

    rerender(<InformacionPersonal informacion={soltera} onGuardar={vi.fn()} guardado />)

    expect(screen.getByText('La información personal se guardó.')).toBeInTheDocument()
  })

  it('muestra el error del servidor', () => {
    const error = new ApiError(500, 'INTERNAL_ERROR', 'Error interno del servidor')
    render(<InformacionPersonal informacion={soltera} onGuardar={vi.fn()} error={error} />)

    expect(screen.getByRole('alert')).toHaveTextContent('Error interno del servidor')
  })
})

describe('<PostulacionBloqueada />', () => {
  it('muestra el motivo y el código, sin ningún botón para continuar', () => {
    render(
      <PostulacionBloqueada
        motivo="Solo pueden postularse ecuatorianos."
        codigo="BIESS-2026-000001"
      />,
    )

    expect(screen.getByRole('heading', { name: 'Postulación cerrada' })).toBeInTheDocument()
    expect(screen.getByText('Código BIESS-2026-000001')).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent('Solo pueden postularse ecuatorianos.')
    expect(screen.queryAllByRole('button')).toHaveLength(0)
  })
})
