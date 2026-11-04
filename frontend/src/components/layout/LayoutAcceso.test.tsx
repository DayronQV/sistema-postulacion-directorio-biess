import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import LayoutAcceso from './LayoutAcceso'

function montar(ruta: string) {
  render(
    <MemoryRouter initialEntries={[ruta]}>
      <Routes>
        <Route element={<LayoutAcceso />}>
          <Route path="/" element={<p>Inicio</p>} />
          <Route path="/login" element={<p>Login del postulante</p>} />
          <Route path="/admin/login" element={<p>Login del panel</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
  return userEvent.setup()
}

describe('LayoutAcceso: enlace al acceso del personal', () => {
  it.each(['/', '/login'])(
    'en %s ofrece el acceso del personal y lleva al login del panel',
    async (ruta) => {
      const user = montar(ruta)
      await user.click(screen.getByRole('link', { name: /Acceso del personal del BIESS/ }))
      expect(await screen.findByText('Login del panel')).toBeInTheDocument()
    },
  )

  it('en el login del panel no repite el enlace (la tarjeta ya trae el de vuelta)', () => {
    montar('/admin/login')
    expect(
      screen.queryByRole('link', { name: /Acceso del personal del BIESS/ }),
    ).not.toBeInTheDocument()
  })
})

describe('LayoutAcceso: pie de página', () => {
  it('el escudo y la línea no se encogen: en el celular el texto largo los cortaba', () => {
    montar('/login')
    const pie = screen.getByRole('contentinfo')
    const escudo = within(pie).getByText('verified_user')
    expect(escudo).toHaveClass('shrink-0')
    expect(escudo.nextElementSibling).toHaveClass('shrink-0')
  })
})
