import { useCallback } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import { useApi } from '@/hooks/useApi'
import { registrarCuenta, type DatosCuenta } from '@/services/auth'
import { ApiError } from '@/services/errores'
import { aplicarErroresDeCampo } from './errores'
import { esquemaCuenta, type FormCuenta } from './esquemas'
import CampoClave from './CampoClave'
import IndicadorFortaleza from './IndicadorFortaleza'

interface PasoCuentaProps {
  ticket: string
  cedula: string
  onRegistrado: () => void
  // El ticket de identidad venció (10 min) o ya se usó: hay que volver al paso 1.
  onTicketVencido: (mensaje: string) => void
  onVolver: () => void
}

// Paso 2 del registro: correo y contraseña de la cuenta, con el ticket del paso 1.
export default function PasoCuenta({
  ticket,
  cedula,
  onRegistrado,
  onTicketVencido,
  onVolver,
}: PasoCuentaProps) {
  const form = useForm<FormCuenta>({
    resolver: zodResolver(esquemaCuenta),
    defaultValues: { email: '', password: '', confirmacion: '' },
  })
  const { setError } = form

  const llamada = useCallback(
    async (signal: AbortSignal, datos: DatosCuenta) => {
      try {
        return await registrarCuenta(signal, datos)
      } catch (e) {
        if (e instanceof ApiError) {
          if (e.code === 'UNAUTHORIZED') onTicketVencido(e.message)
          aplicarErroresDeCampo(e, setError, { email: 'email', password: 'password' })
        }
        throw e
      }
    },
    [onTicketVencido, setError],
  )
  const { ejecutar, cargando, error } = useApi(llamada)

  const password = useWatch({ control: form.control, name: 'password' })

  const enviar = form.handleSubmit(async (v) => {
    const id = await ejecutar({ ticket, email: v.email.trim(), password: v.password })
    if (id) onRegistrado()
  })

  return (
    <FormProvider {...form}>
      <form noValidate onSubmit={enviar} className="flex flex-col gap-space-md">
        <Alert tipo="success" titulo="Identidad validada">
          Cédula {cedula}. Ahora cree las credenciales de su cuenta.
        </Alert>
        <AvisoCuenta error={error} />

        <fieldset disabled={cargando} className="flex flex-col gap-space-md">
          <CampoTexto
            name="email"
            etiqueta="Correo electrónico"
            requerido
            type="email"
            autoComplete="email"
            placeholder="nombre@correo.com"
          />
          <div className="flex flex-col gap-space-sm">
            <CampoClave
              name="password"
              etiqueta="Contraseña"
              requerido
              autoComplete="new-password"
              ayuda="Al menos 8 caracteres, con letras y números."
            />
            <IndicadorFortaleza password={password} />
          </div>
          <CampoClave
            name="confirmacion"
            etiqueta="Confirmar contraseña"
            requerido
            autoComplete="new-password"
          />
        </fieldset>

        <div className="flex flex-col-reverse gap-space-sm sm:flex-row">
          <Button variante="ghost" onClick={onVolver} disabled={cargando}>
            Volver
          </Button>
          <Button
            type="submit"
            className="sm:flex-1"
            cargando={cargando}
            textoCargando="Creando cuenta..."
          >
            Crear cuenta
          </Button>
        </div>
      </form>
    </FormProvider>
  )
}

function AvisoCuenta({ error }: { error: ApiError | null }) {
  // El 401 lo maneja la página (vuelve al paso 1); los errores de correo y contraseña
  // se muestran junto a cada campo.
  if (!error || error.code === 'UNAUTHORIZED') return null
  if (error.fields?.email || error.fields?.password) return null
  return <Alert tipo="error">{error.message}</Alert>
}
