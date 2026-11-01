import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FormProvider, useForm } from 'react-hook-form'
import { vi } from 'vitest'
import { CampoSiNo, CampoTexto } from './campos'

function Formulario({ alEnviar }: { alEnviar: (v: unknown) => void }) {
  const metodos = useForm()
  return (
    <FormProvider {...metodos}>
      <form noValidate onSubmit={metodos.handleSubmit(alEnviar)}>
        <CampoTexto
          name="cedula"
          etiqueta="Cédula"
          requerido
          reglas={{ required: 'La cédula es obligatoria' }}
        />
        <CampoSiNo name="familiares" etiqueta="¿Tiene familiares?" />
        <button type="submit">Enviar</button>
      </form>
    </FormProvider>
  )
}

describe('campos de formulario', () => {
  it('muestra la marca de obligatorio y el mensaje de error del campo', async () => {
    const alEnviar = vi.fn()
    render(<Formulario alEnviar={alEnviar} />)
    expect(screen.getByText('*')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Enviar' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('La cédula es obligatoria')
    expect(screen.getByLabelText(/Cédula/)).toHaveAttribute('aria-invalid', 'true')
    expect(alEnviar).not.toHaveBeenCalled()
  })

  it('envía los valores cuando el formulario es válido', async () => {
    const alEnviar = vi.fn()
    render(<Formulario alEnviar={alEnviar} />)
    await userEvent.type(screen.getByLabelText(/Cédula/), '1712345675')
    await userEvent.click(screen.getByRole('radio', { name: 'Sí' }))
    await userEvent.click(screen.getByRole('button', { name: 'Enviar' }))
    await waitFor(() => expect(alEnviar).toHaveBeenCalled())
    expect(alEnviar.mock.calls[0][0]).toEqual({ cedula: '1712345675', familiares: true })
  })
})

describe('CampoSiNo', () => {
  it('el grupo Sí/No toma el nombre de su etiqueta', () => {
    render(<Formulario alEnviar={() => undefined} />)
    expect(screen.getByRole('radiogroup', { name: '¿Tiene familiares?' })).toBeInTheDocument()
  })
})
