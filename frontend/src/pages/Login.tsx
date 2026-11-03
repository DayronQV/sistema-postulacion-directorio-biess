import { Link, useLocation, useNavigate } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import TarjetaAcceso from '@/components/layout/TarjetaAcceso'
import FormularioLogin from '@/features/auth/FormularioLogin'
import type { EstadoRedireccion } from '@/routes/RutaProtegida'
import { RUTA_FORMULARIO } from '@/routes/rutas'
import type { RespuestaLogin } from '@/services/auth'
import { iniciarSesion } from '@/store/sesion'

interface EstadoNavegacion extends Partial<EstadoRedireccion> {
  registrado?: boolean // viene del registro recién terminado
}

// Inicio de sesión del postulante (EF-11). Guarda la sesión y redirige: en el primer
// ingreso, obligatoriamente al formulario de postulación; si no, a la ruta privada
// que pidió el login o al formulario.
export default function Login() {
  const navigate = useNavigate()
  const estado = (useLocation().state ?? {}) as EstadoNavegacion

  const onIngreso = (r: RespuestaLogin) => {
    iniciarSesion(r.token, r.expires_in, { primerIngreso: r.primer_ingreso })
    const destino = r.primer_ingreso ? RUTA_FORMULARIO : (estado.desde ?? RUTA_FORMULARIO)
    navigate(destino, { replace: true })
  }

  return (
    <TarjetaAcceso
      icono="person"
      titulo="Iniciar sesión"
      descripcion="Ingrese con su cédula y la contraseña que registró."
      pie={
        <>
          ¿Aún no tiene cuenta?{' '}
          <Link to="/registro" className="text-body-bold text-primary-container underline">
            Regístrese
          </Link>
        </>
      }
    >
      {estado.expirada && (
        <Alert tipo="warning" titulo="Su sesión expiró">
          Inicie sesión nuevamente para continuar donde estaba.
        </Alert>
      )}

      {estado.registrado && (
        <Alert tipo="success" titulo="Registro exitoso">
          Ya puede iniciar sesión con su cédula y su contraseña.
        </Alert>
      )}

      <FormularioLogin onIngreso={onIngreso} />
    </TarjetaAcceso>
  )
}
