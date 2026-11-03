import { Navigate, Route, Routes } from 'react-router-dom'
import LayoutPrivado from '@/components/layout/LayoutPrivado'
import LayoutPublico from '@/components/layout/LayoutPublico'
import Componentes from '@/pages/Componentes'
import FormularioPostulacion from '@/pages/FormularioPostulacion'
import ImprimirPostulacion from '@/pages/ImprimirPostulacion'
import Inicio from '@/pages/Inicio'
import Login from '@/pages/Login'
import Recuperar from '@/pages/Recuperar'
import Registro from '@/pages/Registro'
import Restablecer from '@/pages/Restablecer'
import LayoutAdmin from '@/features/admin/LayoutAdmin'
import LoginAdmin from '@/pages/admin/LoginAdmin'
import Fases from '@/pages/admin/Fases'
import PanelAdmin from '@/pages/admin/PanelAdmin'
import PostulacionDetalle from '@/pages/admin/PostulacionDetalle'
import Postulaciones from '@/pages/admin/Postulaciones'
import RutaInvitado from './RutaInvitado'
import RutaProtegida from './RutaProtegida'
import { RUTA_ADMIN, RUTA_FORMULARIO, RUTA_LOGIN, RUTA_LOGIN_ADMIN } from './rutas'

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
          <Route path={RUTA_LOGIN_ADMIN} element={<LoginAdmin />} />
        </Route>
        <Route path="/recuperar" element={<Recuperar />} />
        <Route path="/restablecer" element={<Restablecer />} />
        {/* Muestra interna de componentes: solo en desarrollo */}
        {import.meta.env.DEV && <Route path="/componentes" element={<Componentes />} />}
      </Route>

      {/* Rutas privadas del postulante: todas dentro de RutaProtegida. Un analista o un
          administrador no entra a ellas (el servidor también responde 403). */}
      <Route element={<RutaProtegida roles={['postulante']} />}>
        {/* Vista imprimible: sin menús, solo el documento */}
        <Route path="/postulacion/imprimir" element={<ImprimirPostulacion />} />
        <Route element={<LayoutPrivado />}>
          <Route path={RUTA_FORMULARIO} element={<FormularioPostulacion />} />
        </Route>
      </Route>

      {/* Panel de administración (EF-13): solo analistas y administradores; sin sesión,
          al login del panel. Cada módulo agrega aquí su ruta y su ítem en features/admin/menu.ts */}
      <Route
        element={
          <RutaProtegida roles={['analista', 'administrador']} redirigirA={RUTA_LOGIN_ADMIN} />
        }
      >
        <Route element={<LayoutAdmin />}>
          <Route path={RUTA_ADMIN} element={<PanelAdmin />} />
          <Route path={`${RUTA_ADMIN}/fases`} element={<Fases />} />
          <Route path={`${RUTA_ADMIN}/postulaciones`} element={<Postulaciones />} />
          <Route path={`${RUTA_ADMIN}/postulaciones/:id`} element={<PostulacionDetalle />} />
        </Route>
      </Route>

      {/* Una URL desconocida no muestra una página en blanco */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
