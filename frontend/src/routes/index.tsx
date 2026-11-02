import { Navigate, Route, Routes } from 'react-router-dom'
import LayoutPrivado from '@/components/layout/LayoutPrivado'
import LayoutPublico from '@/components/layout/LayoutPublico'
import Componentes from '@/pages/Componentes'
import FormularioPostulacion from '@/pages/FormularioPostulacion'
import Inicio from '@/pages/Inicio'
import Login from '@/pages/Login'
import Recuperar from '@/pages/Recuperar'
import Registro from '@/pages/Registro'
import Restablecer from '@/pages/Restablecer'
import RutaInvitado from './RutaInvitado'
import RutaProtegida from './RutaProtegida'
import { RUTA_FORMULARIO, RUTA_LOGIN } from './rutas'

export default function AppRoutes() {
  return (
    <Routes>
      {/* Rutas públicas */}
      <Route element={<LayoutPublico />}>
        <Route path="/" element={<Inicio />} />
        {/* Solo sin sesión: con sesión abierta se va a la página principal */}
        <Route element={<RutaInvitado />}>
          <Route path="/registro" element={<Registro />} />
          <Route path={RUTA_LOGIN} element={<Login />} />
        </Route>
        <Route path="/recuperar" element={<Recuperar />} />
        <Route path="/restablecer" element={<Restablecer />} />
        {/* Muestra interna de componentes: solo en desarrollo */}
        {import.meta.env.DEV && <Route path="/componentes" element={<Componentes />} />}
      </Route>

      {/* Rutas privadas del postulante: todas dentro de RutaProtegida */}
      <Route element={<RutaProtegida />}>
        <Route element={<LayoutPrivado />}>
          <Route path={RUTA_FORMULARIO} element={<FormularioPostulacion />} />
        </Route>
      </Route>

      {/* Una URL desconocida no muestra una página en blanco */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
