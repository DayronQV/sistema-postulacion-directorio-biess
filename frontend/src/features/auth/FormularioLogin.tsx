import { useCallback } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'
import { CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import { useApi } from '@/hooks/useApi'
import { iniciarSesionPostulante, type Credenciales, type RespuestaLogin } from '@/services/auth'
import { ApiError } from '@/services/errores'
import { aplicarErroresDeCampo } from './errores'
import { esquemaLogin, type FormLogin } from './esquemas'
import { BLOQUEO_POR_DEFECTO_S, formatoMinutos, useCuentaRegresiva } from './useCuentaRegresiva'

const MSG_BLOQUEO = 'Cuenta bloqueada por intentos fallidos, inténtelo nuevamente en 10 minutos'

interface FormularioLoginProps {
  onIngreso: (respuesta: RespuestaLogin) => void
}

// Inicio de sesión del postulante con cédula y contraseña. Al tercer intento fallido
// el servidor bloquea la cuenta 10 minutos: aquí se muestra la cuenta regresiva con el
// formulario deshabilitado.
export default function FormularioLogin({ onIngreso }: FormularioLoginProps) {
  const form = useForm<FormLogin>({
    resolver: zodResolver(esquemaLogin),
    defaultValues: { cedula: '', password: '' },
  })
  const { setError, resetField } = form
  const { iniciar, restante, activa: bloqueado } = useCuentaRegresiva()

  const llamada = useCallback(
    async (signal: AbortSignal, credenciales: Credenciales) => {
      try {
        return await iniciarSesionPostulante(signal, credenciales)
      } catch (e) {
        if (e instanceof ApiError) {
          if (e.code === 'LOCKED') iniciar(e.retryAfterSeconds ?? BLOQUEO_POR_DEFECTO_S)
          // Tras un intento fallido se limpia la contraseña para el siguiente.
          if (e.code === 'UNAUTHORIZED' || e.code === 'LOCKED') resetField('password')
          aplicarErroresDeCampo(e, setError, { cedula: 'cedula', password: 'password' })
        }
        throw e
      }
    },
    [iniciar, resetField, setError],
  )
  const { ejecutar, cargando, error } = useApi(llamada)

  const enviar = form.handleSubmit(async (v) => {
    const respuesta = await ejecutar({ cedula: v.cedula.trim(), password: v.password })
    if (respuesta) onIngreso(respuesta)
  })

  return (
    <FormProvider {...form}>
      <form noValidate onSubmit={enviar} className="flex flex-col gap-space-md">
        <AvisoLogin error={error} bloqueado={bloqueado} restante={restante} />

        <fieldset disabled={bloqueado || cargando} className="flex flex-col gap-space-md">
          <CampoTexto
            name="cedula"
            etiqueta="Cédula de identidad"
            requerido
            inputMode="numeric"
            autoComplete="username"
            maxLength={10}
            placeholder="10 dígitos"
          />
          <div className="flex flex-col gap-space-xs">
            <CampoTexto
              name="password"
              etiqueta="Contraseña"
              requerido
              type="password"
              autoComplete="current-password"
            />
            <Link
              to="/recuperar"
              className="self-end text-label-md text-primary-container underline"
            >
              ¿Olvidaste tu contraseña?
            </Link>
          </div>
        </fieldset>

        <Button
          type="submit"
          anchoCompleto
          cargando={cargando}
          textoCargando="Ingresando..."
          disabled={bloqueado}
        >
          Iniciar sesión
        </Button>
      </form>
    </FormProvider>
  )
}

interface AvisoLoginProps {
  error: ApiError | null
  bloqueado: boolean
  restante: number
}

function AvisoLogin({ error, bloqueado, restante }: AvisoLoginProps) {
  if (bloqueado) {
    return (
      <Alert tipo="warning" titulo={error?.message || MSG_BLOQUEO}>
        Podrá intentarlo nuevamente en <strong role="timer">{formatoMinutos(restante)}</strong>.
      </Alert>
    )
  }
  if (!error) return null
  if (error.code === 'LOCKED') return <Alert tipo="info">Ya puede intentarlo nuevamente.</Alert>
  // Los errores por campo ya se muestran junto a cada campo.
  if (error.code === 'VALIDATION_ERROR' && error.fields) return null
  return <Alert tipo="error">{error.message}</Alert>
}
