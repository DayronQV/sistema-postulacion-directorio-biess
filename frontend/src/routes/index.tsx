import { Route, Routes } from 'react-router-dom'
import LayoutPrivado from '@/components/layout/LayoutPrivado'
import LayoutPublico from '@/components/layout/LayoutPublico'
import Componentes from '@/pages/Componentes'
import FormularioPostulacion from '@/pages/FormularioPostulacion'
import Inicio from '@/pages/Inicio'
import Login from '@/pages/Login'
import Recuperar from '@/pages/Recuperar'
import Registro from '@/pages/Registro'

export default function AppRoutes() {
  return (
    <Routes>
      {/* Rutas públicas */}
      <Route element={<LayoutPublico />}>
        <Route path="/" element={<Inicio />} />
        <Route path="/registro" element={<Registro />} />
        <Route path="/login" element={<Login />} />
        <Route path="/recuperar" element={<Recuperar />} />
        {/* Muestra interna de componentes: solo en desarrollo */}
        {import.meta.env.DEV && <Route path="/componentes" element={<Componentes />} />}
      </Route>

      {/* Rutas privadas: la guardia de sesión se agrega con la issue #9 */}
      <Route element={<LayoutPrivado />}>
        <Route path="/postulacion" element={<FormularioPostulacion />} />
      </Route>
    </Routes>
  )
}
