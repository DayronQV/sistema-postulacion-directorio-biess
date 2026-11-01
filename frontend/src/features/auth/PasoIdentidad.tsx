import { useCallback } from 'react'
import { Controller, FormProvider, useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'
import { CampoFecha, CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import { useApi } from '@/hooks/useApi'
import { validarIdentidad, type DatosIdentidad } from '@/services/auth'
import { ApiError } from '@/services/errores'
import { aplicarErroresDeCampo } from './errores'
import { esquemaIdentidad, type FormIdentidad } from './esquemas'
import SelectorDato from './SelectorDato'
import { formatoMinutos, useCuentaRegresiva } from './useCuentaRegresiva'

// Duración del bloqueo si el servidor no informa el tiempo restante (docs/arquitectura.md).
const BLOQUEO_POR_DEFECTO_S = 600
const MSG_BLOQUEO = 'registro fallido, inténtelo nuevamente en 10 minutos'

interface PasoIdentidadProps {
  onValidado: (ticket: string, cedula: string) => void
}

// Paso 1 del registro: cédula y fecha de expedición o código dactilar, contra el
// Registro Civil (simulado). Al tercer fallo el servidor bloquea 10 minutos y aquí se
// muestra la cuenta regresiva con el formulario deshabilitado.
export default function PasoIdentidad({ onValidado }: PasoIdentidadProps) {
  const form = useForm<FormIdentidad>({
    resolver: zodResolver(esquemaIdentidad),
    defaultValues: { cedula: '', dato: 'fecha', fechaExpedicion: '', codigoDactilar: '' },
  })
  const { setError } = form
  const { iniciar, restante, activa: bloqueado } = useCuentaRegresiva()

  // Los efectos del error (bloqueo y errores por campo) se aplican al fallar la llamada.
  const llamada = useCallback(
    async (signal: AbortSignal, datos: DatosIdentidad) => {
      try {
        return await validarIdentidad(signal, datos)
      } catch (e) {
        if (e instanceof ApiError) {
          if (e.code === 'LOCKED') iniciar(e.retryAfterSeconds ?? BLOQUEO_POR_DEFECTO_S)
          aplicarErroresDeCampo(e, setError, {
            cedula: 'cedula',
            fecha_expedicion: 'fechaExpedicion',
            codigo_dactilar: 'codigoDactilar',
          })
        }
        throw e
      }
    },
    [iniciar, setError],
  )
  const { ejecutar, cargando, error } = useApi(llamada)

  const dato = useWatch({ control: form.control, name: 'dato' })

  const enviar = form.handleSubmit(async (v) => {
    const cedula = v.cedula.trim()
    const ticket = await ejecutar(
      v.dato === 'fecha'
        ? { cedula, fecha_expedicion: v.fechaExpedicion }
        : { cedula, codigo_dactilar: v.codigoDactilar.trim().toUpperCase() },
    )
    if (ticket) onValidado(ticket, cedula)
  })

  return (
    <FormProvider {...form}>
      <form noValidate onSubmit={enviar} className="flex flex-col gap-space-md">
        <AvisoIdentidad error={error} bloqueado={bloqueado} restante={restante} />

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

          <div className="flex flex-col gap-1.5">
            <span className="text-label-md text-on-surface">Comprobar identidad con</span>
            <Controller
              name="dato"
              control={form.control}
              render={({ field }) => (
                <SelectorDato
                  valor={field.value}
                  deshabilitado={bloqueado || cargando}
                  onChange={(v) => {
                    field.onChange(v)
                    form.clearErrors(['fechaExpedicion', 'codigoDactilar'])
                  }}
                />
              )}
            />
          </div>

          {dato === 'fecha' ? (
            <CampoFecha
              name="fechaExpedicion"
              etiqueta="Fecha de expedición de la cédula"
              requerido
              ayuda="Día, mes y año (DD/MM/AAAA), como aparece en su cédula."
            />
          ) : (
            <CampoTexto
              name="codigoDactilar"
              etiqueta="Código dactilar"
              requerido
              maxLength={10}
              autoCapitalize="characters"
              placeholder="Ej.: V1234V5678"
              ayuda="10 letras y números, al reverso de su cédula."
            />
          )}
        </fieldset>

        <Button
          type="submit"
          anchoCompleto
          cargando={cargando}
          textoCargando="Validando..."
          disabled={bloqueado}
        >
          Validar identidad
        </Button>
      </form>
    </FormProvider>
  )
}

interface AvisoIdentidadProps {
  error: ApiError | null
  bloqueado: boolean
  restante: number
}

function AvisoIdentidad({ error, bloqueado, restante }: AvisoIdentidadProps) {
  if (bloqueado) {
    return (
      <Alert tipo="warning" titulo={error?.message || MSG_BLOQUEO}>
        Podrá intentarlo nuevamente en <strong role="timer">{formatoMinutos(restante)}</strong>.
      </Alert>
    )
  }
  if (!error) return null
  if (error.code === 'LOCKED') return <Alert tipo="info">Ya puede intentarlo nuevamente.</Alert>
  if (error.code === 'CONFLICT') {
    return (
      <Alert tipo="info" titulo={error.message}>
        Si ya tiene una cuenta,{' '}
        <Link to="/login" className="font-bold underline">
          inicie sesión
        </Link>
        .
      </Alert>
    )
  }
  // Los errores por campo ya se muestran junto a cada campo.
  if (error.code === 'VALIDATION_ERROR' && error.fields) return null
  return <Alert tipo="error">{error.message}</Alert>
}
