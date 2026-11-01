import { useState } from 'react'
import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import FormularioIdentidad from '@/features/auth/FormularioIdentidad'
import { solicitarRecuperacion } from '@/services/auth'

// "¿Olvidaste tu contraseña?" (EF-11): valida la identidad contra el Registro Civil y
// el servidor envía el enlace al correo registrado. El mensaje de confirmación es el
// mismo tenga o no cuenta la cédula, para no revelar quién está registrado.
export default function Recuperar() {
  const [enviado, setEnviado] = useState(false)

  return (
    <main className="mx-auto flex w-full max-w-lg flex-1 flex-col gap-space-lg px-margin-mobile py-space-lg sm:px-margin">
      <div className="flex flex-col gap-space-xs">
        <h1 className="text-headline-lg-mobile text-primary sm:text-headline-lg">
          Recuperar contraseña
        </h1>
        {!enviado && (
          <p className="text-body-md text-on-surface-variant">
            Compruebe su identidad y le enviaremos un enlace para crear una nueva contraseña.
          </p>
        )}
      </div>

      <Card>
        {enviado ? (
          <div className="flex flex-col gap-space-md">
            <Alert tipo="success" titulo="Solicitud recibida">
              Si la cédula tiene una cuenta, enviamos al correo registrado un enlace para
              restablecer la contraseña. El enlace vence en 30 minutos y sirve una sola vez.
            </Alert>
            <Button variante="secondary" anchoCompleto onClick={() => setEnviado(false)}>
              Enviar otra solicitud
            </Button>
          </div>
        ) : (
          <FormularioIdentidad
            enviar={solicitarRecuperacion}
            onExito={() => setEnviado(true)}
            textoBoton="Enviar enlace"
            mensajeBloqueo="recuperación fallida, inténtelo nuevamente en 10 minutos"
          />
        )}
      </Card>

      <p className="text-center text-body-md text-on-surface-variant">
        ¿Recordó su contraseña?{' '}
        <Link to="/login" className="text-body-bold text-primary-container underline">
          Inicie sesión
        </Link>
      </p>
    </main>
  )
}
