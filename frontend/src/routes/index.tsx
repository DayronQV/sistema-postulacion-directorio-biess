import { Route, Routes } from 'react-router-dom'
import FormularioPostulacion from '@/pages/FormularioPostulacion'
import Inicio from '@/pages/Inicio'
import Login from '@/pages/Login'
import Recuperar from '@/pages/Recuperar'
import Registro from '@/pages/Registro'

export default function AppRoutes() {
  return (
    <Routes>
      {/* Rutas públicas */}
      <Route path="/" element={<Inicio />} />
      <Route path="/registro" element={<Registro />} />
      <Route path="/login" element={<Login />} />
      <Route path="/recuperar" element={<Recuperar />} />

      {/* Rutas privadas: la guardia de sesión se agrega con la issue #9 */}
      <Route path="/postulacion" element={<FormularioPostulacion />} />
    </Routes>
  )
}
