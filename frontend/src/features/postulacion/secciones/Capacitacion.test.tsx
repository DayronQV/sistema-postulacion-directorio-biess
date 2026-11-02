import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento } from '@/services/documentos'
import type { Capacitacion as CapacitacionDatos } from '@/services/postulacion'
import Capacitacion from './Capacitacion'

vi.mock('@/services/documentos', () => ({ eliminarDocumento: vi.fn() }))

// Cada clic entrega un documento nuevo (doc-1, doc-2...).
let siguiente = 0
vi.mock('@/components/form/DocumentoUpload', () => ({
  default: ({ onChange }: { onChange: (valor: unknown) => void }) => (
    <button
      type="button"
      onClick={() => {
        siguiente += 1
        onChange({ documento_id: `doc-${siguiente}`, nombre: 'respaldo.pdf', tamano_bytes: 100 })
      }}
    >
      Adjuntar PDF de prueba
    </button>
  ),
}))

const mockEliminar = eliminarDocumento as Mock

const curso: CapacitacionDatos = {
  tipo: 'curso',
  nombre: 'Finanzas',
  institucion: 'Universidad',
  fecha_inicio: '2024-01-01',
  fecha_fin: '2024-01-05',
  horas: 40,
  documento_id: '',
}

// Pega el texto de una vez: tipear carácter por carácter vuelve lentas estas pruebas.
async function escribir(etiqueta: RegExp, texto: string) {
  const campo = screen.getByLabelText(etiqueta)
  await userEvent.clear(campo)
  await userEvent.click(campo)
  await userEvent.paste(texto)
}

async function llenarCurso() {
  await userEvent.selectOptions(screen.getByLabelText(/^Tipo de evento/), 'curso')
  await escribir(/^Nombre del evento/, 'Finanzas')
  await escribir(/^Institución/, 'Universidad')
  await escribir(/^Número de horas/, '40')
  fireEvent.change(screen.getByLabelText(/^Fecha de inicio/), { target: { value: '2024-01-01' } })
  fireEvent.change(screen.getByLabelText(/^Fecha de finalización/), {
    target: { value: '2024-01-05' },
  })
}

beforeEach(() => {
  siguiente = 0
  mockEliminar.mockReset()
  mockEliminar.mockResolvedValue(undefined)
})

describe('Capacitacion', () => {
  it('agrega y guarda un evento sin exigir PDF', async () => {
    const onGuardar = vi.fn()
    render(<Capacitacion valor={[]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    await llenarCurso()
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))

    expect(onGuardar).toHaveBeenCalledWith([curso])
  })

  it('la tabla muestra la etiqueta del tipo, no su valor interno', () => {
    render(<Capacitacion valor={[{ ...curso, tipo: 'taller' }]} onGuardar={vi.fn()} />)
    expect(screen.getByText('Taller')).toBeInTheDocument()
    expect(screen.queryByText('taller')).not.toBeInTheDocument()
  })

  it('rechaza horas en cero y por encima del máximo', async () => {
    render(<Capacitacion valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Ingrese un número de horas mayor a cero')).toBeInTheDocument()

    await escribir(/^Número de horas/, '10001')
    await userEvent.click(screen.getByRole('button', { name: 'Agregar' }))
    expect(screen.getByText('Máximo 10000 horas')).toBeInTheDocument()
  })

  it('edita un evento y borra al guardar el PDF subido que no se usó', async () => {
    const onGuardar = vi.fn()
    render(<Capacitacion valor={[curso]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await escribir(/^Número de horas/, '60')
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Actualizar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))

    expect(onGuardar).toHaveBeenCalledWith([{ ...curso, horas: 60, documento_id: 'doc-2' }])
    expect(mockEliminar).toHaveBeenCalledTimes(1)
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
  })
})
