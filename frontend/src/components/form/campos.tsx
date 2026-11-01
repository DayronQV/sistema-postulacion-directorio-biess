import { useId, type ComponentProps } from 'react'
import { Controller, get, useFormContext, type RegisterOptions } from 'react-hook-form'
import Checkbox from '@/components/ui/Checkbox'
import DateInput from '@/components/ui/DateInput'
import Input from '@/components/ui/Input'
import RadioSiNo from '@/components/ui/RadioSiNo'
import Select, { type OpcionSelect } from '@/components/ui/Select'
import Textarea from '@/components/ui/Textarea'
import Campo from './Campo'

// Envoltorios para react-hook-form: leen register y errores del <FormProvider> más cercano.

interface Base {
  name: string
  etiqueta: string
  requerido?: boolean
  ayuda?: string
  reglas?: RegisterOptions
}

function useCampo(name: string) {
  const id = useId()
  const { formState } = useFormContext()
  const mensaje = get(formState.errors, `${name}.message`) as string | undefined
  return { id, error: mensaje }
}

function describedBy(id: string, error?: string, ayuda?: string) {
  if (error) return `${id}-error`
  return ayuda ? `${id}-ayuda` : undefined
}

export function CampoTexto({
  name,
  etiqueta,
  requerido,
  ayuda,
  reglas,
  ...resto
}: Base & Omit<ComponentProps<typeof Input>, 'name' | 'invalido'>) {
  const { register } = useFormContext()
  const { id, error } = useCampo(name)
  return (
    <Campo id={id} etiqueta={etiqueta} requerido={requerido} error={error} ayuda={ayuda}>
      <Input
        id={id}
        invalido={Boolean(error)}
        aria-describedby={describedBy(id, error, ayuda)}
        {...register(name, reglas)}
        {...resto}
      />
    </Campo>
  )
}

export function CampoFecha({
  name,
  etiqueta,
  requerido,
  ayuda,
  reglas,
  ...resto
}: Base & Omit<ComponentProps<typeof DateInput>, 'name' | 'invalido'>) {
  const { register } = useFormContext()
  const { id, error } = useCampo(name)
  return (
    <Campo id={id} etiqueta={etiqueta} requerido={requerido} error={error} ayuda={ayuda}>
      <DateInput
        id={id}
        invalido={Boolean(error)}
        aria-describedby={describedBy(id, error, ayuda)}
        {...register(name, reglas)}
        {...resto}
      />
    </Campo>
  )
}

export function CampoSelect({
  name,
  etiqueta,
  requerido,
  ayuda,
  reglas,
  opciones,
  placeholder = 'Seleccione...',
}: Base & { opciones: OpcionSelect[]; placeholder?: string }) {
  const { register } = useFormContext()
  const { id, error } = useCampo(name)
  return (
    <Campo id={id} etiqueta={etiqueta} requerido={requerido} error={error} ayuda={ayuda}>
      <Select
        id={id}
        opciones={opciones}
        placeholder={placeholder}
        invalido={Boolean(error)}
        aria-describedby={describedBy(id, error, ayuda)}
        {...register(name, reglas)}
      />
    </Campo>
  )
}

export function CampoTextarea({
  name,
  etiqueta,
  requerido,
  ayuda,
  reglas,
  ...resto
}: Base & Omit<ComponentProps<typeof Textarea>, 'name' | 'invalido'>) {
  const { register } = useFormContext()
  const { id, error } = useCampo(name)
  return (
    <Campo id={id} etiqueta={etiqueta} requerido={requerido} error={error} ayuda={ayuda}>
      <Textarea
        id={id}
        invalido={Boolean(error)}
        aria-describedby={describedBy(id, error, ayuda)}
        {...register(name, reglas)}
        {...resto}
      />
    </Campo>
  )
}

export function CampoSiNo({ name, etiqueta, requerido, ayuda, reglas }: Base) {
  const { control } = useFormContext()
  const { id, error } = useCampo(name)
  return (
    <Campo id={id} etiqueta={etiqueta} requerido={requerido} error={error} ayuda={ayuda}>
      <Controller
        name={name}
        control={control}
        rules={reglas}
        render={({ field }) => (
          <RadioSiNo
            id={id}
            aria-labelledby={`${id}-etiqueta`}
            aria-describedby={describedBy(id, error, ayuda)}
            aria-invalid={Boolean(error) || undefined}
            valor={field.value}
            onChange={field.onChange}
          />
        )}
      />
    </Campo>
  )
}

export function CampoCheckbox({ name, etiqueta, reglas }: Omit<Base, 'requerido' | 'ayuda'>) {
  const { register } = useFormContext()
  const { id, error } = useCampo(name)
  return (
    <div className="flex flex-col gap-1.5">
      <Checkbox
        id={id}
        etiqueta={etiqueta}
        invalido={Boolean(error)}
        aria-describedby={describedBy(id, error)}
        {...register(name, reglas)}
      />
      {error && (
        <p id={`${id}-error`} role="alert" className="text-label-sm text-error">
          {error}
        </p>
      )}
    </div>
  )
}
