import { Link, useLocation, useNavigate } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import FormularioLoginAdmin from '@/features/admin/FormularioLoginAdmin'
import type { EstadoRedireccion } from '@/routes/RutaProtegida'
import { RUTA_ADMIN } from '@/routes/rutas'
import type { RespuestaLoginAdmin } from '@/services/adminAuth'
import { iniciarSesion } from '@/store/sesion'

// Ingreso al panel de administración (EF-13) para analistas y administradores.
export default function LoginAdmin() {
  const navigate = useNavigate()
  const estado = (useLocation().state ?? {}) as Partial<EstadoRedireccion>

  const onIngreso = (r: RespuestaLoginAdmin) => {
    iniciarSesion(r.token, r.expires_in)
    // Vuelve a la pantalla del panel que pidió el login, si la hay.
    const destino = estado.desde?.startsWith(RUTA_ADMIN) ? estado.desde : RUTA_ADMIN
    navigate(destino, { replace: true })
  }

  return (
    <main className="mx-auto flex w-full max-w-lg flex-1 flex-col gap-space-lg px-margin-mobile py-space-lg sm:px-margin">
      <div className="flex flex-col gap-space-xs">
        <h1 className="text-headline-lg-mobile text-primary sm:text-headline-lg">
          Panel de administración
        </h1>
        <p className="text-body-md text-on-surface-variant">
          Ingreso para el personal de Talento Humano y de la Comisión de Selección.
        </p>
      </div>

      {estado.expirada && (
        <Alert tipo="warning" titulo="Su sesión expiró">
          Inicie sesión nuevamente para continuar donde estaba.
        </Alert>
      )}

      <Card>
        <FormularioLoginAdmin onIngreso={onIngreso} />
      </Card>

      <p className="text-center text-body-md text-on-surface-variant">
        ¿Es postulante?{' '}
        <Link to="/login" className="text-body-bold text-primary-container underline">
          Ingrese aquí
        </Link>
      </p>
    </main>
  )
}
