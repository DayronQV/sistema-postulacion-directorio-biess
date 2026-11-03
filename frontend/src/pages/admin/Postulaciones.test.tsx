import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { vi, type Mock } from 'vitest'
import { descargarMatrizAdmin, listarPostulacionesAdmin } from '@/services/adminPostulaciones'
import Postulaciones from './Postulaciones'

vi.mock('@/services/adminPostulaciones', () => ({
  listarPostulacionesAdmin: vi.fn(),
  descargarMatrizAdmin: vi.fn(),
}))
const mockListar = listarPostulacionesAdmin as Mock
const mockDescargarMatriz = descargarMatrizAdmin as Mock

beforeEach(() => {
  vi.restoreAllMocks()
  mockListar.mockReset()
  mockDescargarMatriz.mockReset()
  mockDescargarMatriz.mockResolvedValue({
    contenido: new Blob(['xlsx']),
    nombre: 'matriz.xlsx',
  })
  Object.defineProperty(URL, 'createObjectURL', {
    configurable: true,
    value: vi.fn(() => 'blob:matriz'),
  })
  Object.defineProperty(URL, 'revokeObjectURL', { configurable: true, value: vi.fn() })
  vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined)
  mockListar.mockResolvedValue({
    items: [
      {
        id: 'p1',
        codigo: 'BIESS-2026-000001',
        cedula: '0912345678',
        nombres: 'Ana',
        apellidos: 'Pérez',
        cargo: 'afiliados',
        estado: 'enviada',
        elegible: true,
      },
    ],
    total: 21,
    pagina: 1,
    por_pagina: 20,
  })
})

function renderPagina() {
  return render(
    <MemoryRouter>
      <Postulaciones />
    </MemoryRouter>,
  )
}

it('muestra la nómina y el enlace al detalle', async () => {
  renderPagina()

  expect(await screen.findByText('BIESS-2026-000001')).toBeInTheDocument()
  expect(screen.getByText('Ana Pérez')).toBeInTheDocument()
  expect(screen.getByRole('link', { name: 'Ver detalle' })).toHaveAttribute(
    'href',
    '/admin/postulaciones/p1',
  )
  expect(screen.getAllByText('Página 1 de 2')).toHaveLength(2)
})

it('aplica filtros y vuelve a la primera página', async () => {
  const user = userEvent.setup()
  renderPagina()
  await screen.findByText('BIESS-2026-000001')

  await user.selectOptions(screen.getByLabelText('Cargo'), 'afiliados')

  expect(mockListar).toHaveBeenLastCalledWith(
    expect.any(AbortSignal),
    expect.objectContaining({ cargo: 'afiliados', pagina: 1 }),
  )
})

it('muestra una lista vacía sin bloquear la página', async () => {
  mockListar.mockResolvedValue({ items: [], total: 0, pagina: 1, por_pagina: 20 })
  renderPagina()

  expect(await screen.findByText('No hay postulantes con estos filtros.')).toBeInTheDocument()
})

it('exporta Excel con los filtros activos y muestra el loader', async () => {
  const user = userEvent.setup()
  renderPagina()
  await screen.findByText('BIESS-2026-000001')
  await user.selectOptions(screen.getByLabelText('Cargo'), 'afiliados')
  await user.selectOptions(screen.getByLabelText('Estado'), 'enviada')

  await user.click(screen.getByRole('button', { name: 'Exportar Excel' }))

  await waitFor(() =>
    expect(mockDescargarMatriz).toHaveBeenCalledWith(
      expect.any(AbortSignal),
      'xlsx',
      expect.objectContaining({ cargo: 'afiliados', estado: 'enviada', pagina: 1 }),
    ),
  )
  expect(URL.createObjectURL).toHaveBeenCalled()
  expect(HTMLAnchorElement.prototype.click).toHaveBeenCalled()
})
