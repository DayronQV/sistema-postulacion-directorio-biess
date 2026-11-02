import { useCallback } from 'react'
import { FormProvider, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import { aplicarErroresDeCampo } from '@/features/auth/errores'
import {
  BLOQUEO_POR_DEFECTO_S,
  formatoMinutos,
  useCuentaRegresiva,
} from '@/features/auth/useCuentaRegresiva'
import { useApi } from '@/hooks/useApi'
import {
  iniciarSesionAdmin,
  type CredencialesAdmin,
  type RespuestaLoginAdmin,
} from '@/services/adminAuth'
import { ApiError } from '@/services/errores'

const MSG_BLOQUEO = 'Cuenta bloqueada por intentos fallidos, inténtelo nuevamente en 10 minutos'

const esquema = z.object({
  usuario: z.string().trim().min(1, 'Ingrese su usuario').max(50, 'Máximo 50 caracteres'),
  password: z.string().min(1, 'Ingrese su contraseña'),
})
type Valores = z.infer<typeof esquema>

interface FormularioLoginAdminProps {
  onIngreso: (respuesta: RespuestaLoginAdmin) => void
}

// Ingreso de analistas y administradores con usuario y contraseña. Tiene el mismo bloqueo
// que el login del postulante: al tercer fallo, cuenta regresiva y formulario deshabilitado.
export default function FormularioLoginAdmin({ onIngreso }: FormularioLoginAdminProps) {
  const form = useForm<Valores>({
    resolver: zodResolver(esquema),
    defaultValues: { usuario: '', password: '' },
  })
  const { setError, resetField } = form
  const { iniciar, restante, activa: bloqueado } = useCuentaRegresiva()

  const llamada = useCallback(
    async (signal: AbortSignal, credenciales: CredencialesAdmin) => {
      try {
        return await iniciarSesionAdmin(signal, credenciales)
      } catch (e) {
        if (e instanceof ApiError) {
          if (e.code === 'LOCKED') iniciar(e.retryAfterSeconds ?? BLOQUEO_POR_DEFECTO_S)
          if (e.code === 'UNAUTHORIZED' || e.code === 'LOCKED') resetField('password')
          aplicarErroresDeCampo(e, setError, { usuario: 'usuario', password: 'password' })
        }
        throw e
      }
    },
    [iniciar, resetField, setError],
  )
  const { ejecutar, cargando, error } = useApi(llamada)

  const enviar = form.handleSubmit(async (v) => {
    const respuesta = await ejecutar({ usuario: v.usuario.trim(), password: v.password })
    if (respuesta) onIngreso(respuesta)
  })

  return (
    <FormProvider {...form}>
      <form noValidate onSubmit={enviar} className="flex flex-col gap-space-md">
        {bloqueado ? (
          <Alert tipo="warning" titulo={error?.message || MSG_BLOQUEO}>
            Podrá intentarlo nuevamente en <strong role="timer">{formatoMinutos(restante)}</strong>.
          </Alert>
        ) : error?.code === 'LOCKED' ? (
          <Alert tipo="info">Ya puede intentarlo nuevamente.</Alert>
        ) : error && !(error.code === 'VALIDATION_ERROR' && error.fields) ? (
          <Alert tipo="error">{error.message}</Alert>
        ) : null}

        <fieldset disabled={bloqueado || cargando} className="flex flex-col gap-space-md">
          <CampoTexto name="usuario" etiqueta="Usuario" requerido autoComplete="username" />
          <CampoTexto
            name="password"
            etiqueta="Contraseña"
            requerido
            type="password"
            autoComplete="current-password"
          />
        </fieldset>

        <Button
          type="submit"
          anchoCompleto
          cargando={cargando}
          textoCargando="Ingresando..."
          disabled={bloqueado}
        >
          Ingresar al panel
        </Button>
      </form>
    </FormProvider>
  )
}
