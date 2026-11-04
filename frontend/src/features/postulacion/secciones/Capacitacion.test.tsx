import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, type Mock } from 'vitest'
import { eliminarDocumento } from '@/services/documentos'
import { ApiError } from '@/services/errores'
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
  it('Agregar lleva la pantalla al formulario y pone el cursor en su primer campo', async () => {
    const desplazar = vi.fn()
    Element.prototype.scrollIntoView = desplazar
    render(<Capacitacion valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    expect(desplazar).toHaveBeenCalledTimes(1)
    expect(screen.getByLabelText(/^Tipo de evento/)).toHaveFocus()
  })

  it('agrega y guarda un evento sin exigir PDF', async () => {
    const onGuardar = vi.fn()
    render(<Capacitacion valor={[]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    await llenarCurso()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith([curso])
    // Al guardar, la fila aparece en la tabla y el formulario se cierra.
    expect(await screen.findByText('Universidad')).toBeInTheDocument()
    expect(screen.queryByLabelText(/^Número de horas/)).not.toBeInTheDocument()
  })

  it('la tabla muestra la etiqueta del tipo, no su valor interno', () => {
    render(<Capacitacion valor={[{ ...curso, tipo: 'taller' }]} onGuardar={vi.fn()} />)
    expect(screen.getByText('Taller')).toBeInTheDocument()
    expect(screen.queryByText('taller')).not.toBeInTheDocument()
  })

  it('rechaza horas en cero y por encima del máximo', async () => {
    render(<Capacitacion valor={[]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Ingrese un número de horas mayor a cero')).toBeInTheDocument()

    await escribir(/^Número de horas/, '10001')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))
    expect(screen.getByText('Máximo 10000 horas')).toBeInTheDocument()
  })

  it('edita un evento y borra al guardar el PDF subido que no se usó', async () => {
    const onGuardar = vi.fn()
    render(<Capacitacion valor={[curso]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await escribir(/^Número de horas/, '60')
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Adjuntar PDF de prueba' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith([{ ...curso, horas: 60, documento_id: 'doc-2' }])
    await waitFor(() => expect(mockEliminar).toHaveBeenCalledTimes(1))
    expect(mockEliminar).toHaveBeenCalledWith(expect.any(AbortSignal), 'doc-1')
  })

  it('con el formulario abierto solo hay Cancelar y Guardar', async () => {
    render(<Capacitacion valor={[curso]} onGuardar={vi.fn()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    expect(screen.getByRole('button', { name: 'Cancelar' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Guardar' })).toBeInTheDocument()
    for (const sobrante of ['Agregar', 'Actualizar', 'Guardar información']) {
      expect(screen.queryByRole('button', { name: sobrante })).not.toBeInTheDocument()
    }
  })

  it('Guardar información solo aparece con algo pendiente (tras eliminar)', async () => {
    const onGuardar = vi.fn()
    render(<Capacitacion valor={[curso]} onGuardar={onGuardar} />)
    const boton = () => screen.queryByRole('button', { name: 'Guardar información' })
    expect(boton()).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Eliminar' }))
    expect(onGuardar).not.toHaveBeenCalled()
    // Con el formulario abierto se oculta y al cancelar vuelve.
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    expect(boton()).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancelar' }))
    await userEvent.click(screen.getByRole('button', { name: 'Guardar información' }))
    expect(onGuardar).toHaveBeenCalledWith([])
    await waitFor(() => expect(boton()).not.toBeInTheDocument())
  })

  it('si el servidor rechaza el guardado, el formulario sigue abierto y la lista no cambia', async () => {
    const onGuardar = vi.fn().mockRejectedValue(new ApiError(500, 'INTERNAL', 'Error interno'))
    render(<Capacitacion valor={[curso]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Editar' }))
    await escribir(/^Número de horas/, '60')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(await screen.findByText('Error interno')).toBeInTheDocument()
    expect(screen.getByLabelText(/^Número de horas/)).toHaveValue(60)
    expect(screen.getByText('40')).toBeInTheDocument()
  })

  it('agregar conserva los eventos que ya estaban guardados', async () => {
    const onGuardar = vi.fn()
    render(<Capacitacion valor={[curso]} onGuardar={onGuardar} />)
    await userEvent.click(screen.getByRole('button', { name: 'Agregar capacitación' }))
    await llenarCurso()
    await escribir(/^Nombre del evento/, 'Riesgos')
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(onGuardar).toHaveBeenCalledWith([curso, { ...curso, nombre: 'Riesgos' }])
  })
})
