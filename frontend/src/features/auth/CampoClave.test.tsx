import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import CampoClave from './CampoClave'
import IndicadorFortaleza from './IndicadorFortaleza'

function Formulario({ children }: { children: ReactNode }) {
  const form = useForm({ defaultValues: { password: '', confirmacion: '' } })
  return <FormProvider {...form}>{children}</FormProvider>
}

function montar() {
  render(
    <Formulario>
      <CampoClave name="password" etiqueta="Contraseña" />
      <CampoClave name="confirmacion" etiqueta="Confirmar contraseña" />
    </Formulario>,
  )
  return userEvent.setup()
}

describe('CampoClave', () => {
  it('empieza oculta y el ojo la muestra y la vuelve a ocultar', async () => {
    const user = montar()
    const campo = screen.getByLabelText(/^Contraseña/)
    expect(campo).toHaveAttribute('type', 'password')

    await user.click(screen.getByRole('button', { name: 'Mostrar contraseña' }))
    expect(campo).toHaveAttribute('type', 'text')
    expect(screen.getByRole('button', { name: 'Ocultar contraseña' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )

    await user.click(screen.getByRole('button', { name: 'Ocultar contraseña' }))
    expect(campo).toHaveAttribute('type', 'password')
  })

  it('cada campo tiene su propio ojo', async () => {
    const user = montar()
    await user.click(screen.getByRole('button', { name: 'Mostrar confirmar contraseña' }))

    expect(screen.getByLabelText(/^Confirmar contraseña/)).toHaveAttribute('type', 'text')
    expect(screen.getByLabelText(/^Contraseña/)).toHaveAttribute('type', 'password')
  })

  it('el ojo no envía el formulario', async () => {
    const enviar = vi.fn((e: React.FormEvent) => e.preventDefault())
    render(
      <Formulario>
        <form onSubmit={enviar}>
          <CampoClave name="password" etiqueta="Contraseña" />
        </form>
      </Formulario>,
    )
    await userEvent.setup().click(screen.getByRole('button', { name: 'Mostrar contraseña' }))
    expect(enviar).not.toHaveBeenCalled()
  })
})

describe('IndicadorFortaleza', () => {
  it('con la contraseña vacía ocupa su lugar pero no se ve, para que la página no salte', () => {
    render(<IndicadorFortaleza password="" />)
    const indicador = screen.getByTestId('indicador-fortaleza')
    expect(indicador).toBeInTheDocument()
    expect(indicador).toHaveClass('invisible')
  })

  it('con la primera letra se ve con su nivel', () => {
    render(<IndicadorFortaleza password="a" />)
    const indicador = screen.getByTestId('indicador-fortaleza')
    expect(indicador).not.toHaveClass('invisible')
    expect(indicador).toHaveTextContent('Seguridad: Débil')
  })
})
