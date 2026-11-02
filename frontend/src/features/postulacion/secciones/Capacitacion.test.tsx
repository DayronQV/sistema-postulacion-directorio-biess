import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'
import Capacitacion from './Capacitacion'

vi.mock('@/components/form/DocumentoUpload', () => ({ default: () => <span>Documento opcional</span> }))

describe('Capacitacion', () => {
  it('agrega y guarda un evento sin exigir PDF', async () => {
    const usuario = userEvent.setup(); const onGuardar = vi.fn(); render(<Capacitacion valor={[]} onGuardar={onGuardar} />)
    await usuario.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    await usuario.selectOptions(screen.getByLabelText(/^Tipo de evento/), 'curso')
    await usuario.type(screen.getByLabelText(/^Nombre del evento/), 'Finanzas')
    await usuario.type(screen.getByLabelText(/^Institución/), 'Universidad')
    await usuario.type(screen.getByLabelText(/^Número de horas/), '40')
    await usuario.type(screen.getByLabelText(/^Fecha de inicio/), '2024-01-01')
    await usuario.type(screen.getByLabelText(/^Fecha de finalización/), '2024-01-05')
    await usuario.click(screen.getByRole('button', { name: 'Agregar' }))
    await usuario.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith([expect.objectContaining({ tipo: 'curso', nombre: 'Finanzas', horas: 40, documento_id: '' })])
  })

  it('rechaza horas en cero', async () => {
    const usuario = userEvent.setup(); render(<Capacitacion valor={[]} onGuardar={vi.fn()} />)
    await usuario.click(screen.getByRole('button', { name: 'Agregar capacitación' })); await usuario.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Ingrese un número de horas mayor a cero')).toBeInTheDocument()
  })
})
