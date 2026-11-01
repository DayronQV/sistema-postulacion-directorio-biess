import { useCallback, useEffect, useState } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import { aplicarErroresDeCampo } from '@/features/auth/errores'
import { esquemaRestablecer, type FormRestablecer } from '@/features/auth/esquemas'
import IndicadorFortaleza from '@/features/auth/IndicadorFortaleza'
import { useApi } from '@/hooks/useApi'
import { restablecerPassword, type DatosRestablecer } from '@/services/auth'
import { ApiError } from '@/services/errores'

// Tiempo que se muestra el mensaje de éxito antes de ir al login.
export const ESPERA_REDIRECCION_MS = 4000

// Página del enlace del correo (/restablecer?token=...): crea la nueva contraseña.
export default function Restablecer() {
  const token = useSearchParams()[0].get('token')?.trim() ?? ''
  const [listo, setListo] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    if (!listo) return
    const id = setTimeout(() => navigate('/login', { replace: true }), ESPERA_REDIRECCION_MS)
    return () => clearTimeout(id)
  }, [listo, navigate])

  return (
    <main className="mx-auto flex w-full max-w-lg flex-1 flex-col gap-space-lg px-margin-mobile py-space-lg sm:px-margin">
      <h1 className="text-headline-lg-mobile text-primary sm:text-headline-lg">
        Restablecer contraseña
      </h1>
      <Card>
        {!token ? (
          <EnlaceInvalido mensaje="El enlace para restablecer la contraseña no es válido." />
        ) : listo ? (
          <div className="flex flex-col gap-space-md">
            <Alert tipo="success" titulo="Contraseña actualizada">
              Ya puede iniciar sesión con su nueva contraseña. Lo llevaremos al inicio de sesión en
              unos segundos.
            </Alert>
            <Button anchoCompleto onClick={() => navigate('/login', { replace: true })}>
              Ir a iniciar sesión
            </Button>
          </div>
        ) : (
          <FormularioNuevaPassword token={token} onListo={() => setListo(true)} />
        )}
      </Card>
    </main>
  )
}

interface FormularioNuevaPasswordProps {
  token: string
  onListo: () => void
}

function FormularioNuevaPassword({ token, onListo }: FormularioNuevaPasswordProps) {
  const form = useForm<FormRestablecer>({
    resolver: zodResolver(esquemaRestablecer),
    defaultValues: { password: '', confirmacion: '' },
  })
  const { setError } = form

  const llamada = useCallback(
    async (signal: AbortSignal, datos: DatosRestablecer) => {
      try {
        await restablecerPassword(signal, datos)
        return true // 204 sin cuerpo: se devuelve un valor para distinguir el éxito
      } catch (e) {
        if (e instanceof ApiError) aplicarErroresDeCampo(e, setError, { password: 'password' })
        throw e
      }
    },
    [setError],
  )
  const { ejecutar, cargando, error } = useApi(llamada)
  const password = useWatch({ control: form.control, name: 'password' })

  const enviar = form.handleSubmit(async (v) => {
    if (await ejecutar({ token, password: v.password })) onListo()
  })

  // Token vencido, ya usado o inexistente: el formulario ya no sirve.
  if (error?.code === 'UNAUTHORIZED') return <EnlaceInvalido mensaje={error.message} />

  return (
    <FormProvider {...form}>
      <form noValidate onSubmit={enviar} className="flex flex-col gap-space-md">
        {error && !error.fields?.password && <Alert tipo="error">{error.message}</Alert>}

        <fieldset disabled={cargando} className="flex flex-col gap-space-md">
          <div className="flex flex-col gap-space-sm">
            <CampoTexto
              name="password"
              etiqueta="Nueva contraseña"
              requerido
              type="password"
              autoComplete="new-password"
              ayuda="Al menos 8 caracteres, con letras y números."
            />
            <IndicadorFortaleza password={password} />
          </div>
          <CampoTexto
            name="confirmacion"
            etiqueta="Confirmar contraseña"
            requerido
            type="password"
            autoComplete="new-password"
          />
        </fieldset>

        <Button type="submit" anchoCompleto cargando={cargando} textoCargando="Guardando...">
          Guardar contraseña
        </Button>
      </form>
    </FormProvider>
  )
}

function EnlaceInvalido({ mensaje }: { mensaje: string }) {
  return (
    <div className="flex flex-col gap-space-md">
      <Alert tipo="error" titulo="Enlace no válido o vencido">
        {mensaje}
      </Alert>
      <Link to="/recuperar" className="text-center text-body-bold text-primary-container underline">
        Solicitar un nuevo enlace
      </Link>
    </div>
  )
}
