import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import PasoCuenta from '@/features/auth/PasoCuenta'
import PasoIdentidad from '@/features/auth/PasoIdentidad'
import TarjetaAcceso from '@/components/layout/TarjetaAcceso'

// Tiempo que se muestra el mensaje de éxito antes de ir al login.
export const ESPERA_REDIRECCION_MS = 4000

type Estado =
  | { paso: 'identidad'; aviso?: string }
  | { paso: 'cuenta'; ticket: string; cedula: string }
  | { paso: 'listo' }

// Registro en dos pasos (EF-11): identidad contra el Registro Civil y luego correo y
// contraseña. Al terminar lleva al inicio de sesión.
export default function Registro() {
  const [estado, setEstado] = useState<Estado>({ paso: 'identidad' })
  const navigate = useNavigate()

  const onTicketVencido = useCallback(
    (mensaje: string) => setEstado({ paso: 'identidad', aviso: mensaje }),
    [],
  )

  useEffect(() => {
    if (estado.paso !== 'listo') return
    const id = setTimeout(
      () => navigate('/login', { state: { registrado: true } }),
      ESPERA_REDIRECCION_MS,
    )
    return () => clearTimeout(id)
  }, [estado.paso, navigate])

  return (
    <TarjetaAcceso icono="person_add" titulo="Registro">
      <div className="flex flex-col gap-space-xs text-center">
        {estado.paso !== 'listo' && (
          <p className="text-body-md text-on-surface-variant">
            Paso {estado.paso === 'identidad' ? 1 : 2} de 2:{' '}
            {estado.paso === 'identidad' ? 'validar su identidad' : 'crear su cuenta'}
          </p>
        )}
      </div>

      <div>
        {estado.paso === 'identidad' && (
          <div className="flex flex-col gap-space-md">
            {estado.aviso && <Alert tipo="warning">{estado.aviso}</Alert>}
            <PasoIdentidad
              onValidado={(ticket, cedula) => setEstado({ paso: 'cuenta', ticket, cedula })}
            />
          </div>
        )}
        {estado.paso === 'cuenta' && (
          <PasoCuenta
            ticket={estado.ticket}
            cedula={estado.cedula}
            onRegistrado={() => setEstado({ paso: 'listo' })}
            onTicketVencido={onTicketVencido}
            onVolver={() => setEstado({ paso: 'identidad' })}
          />
        )}
        {estado.paso === 'listo' && (
          <div className="flex flex-col gap-space-md">
            <Alert tipo="success" titulo="Registro exitoso">
              Ya puede iniciar sesión con su cédula y su contraseña. Lo llevaremos al inicio de
              sesión en unos segundos.
            </Alert>
            <Button
              anchoCompleto
              onClick={() => navigate('/login', { state: { registrado: true } })}
            >
              Ir a iniciar sesión
            </Button>
          </div>
        )}
      </div>

      {estado.paso !== 'listo' && (
        <p className="text-center text-body-md text-on-surface-variant">
          ¿Ya tiene una cuenta?{' '}
          <Link to="/login" className="text-body-bold text-primary-container underline">
            Inicie sesión
          </Link>
        </p>
      )}
    </TarjetaAcceso>
  )
}
