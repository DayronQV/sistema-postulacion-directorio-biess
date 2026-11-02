import { useEffect } from 'react'
import { FormProvider, useForm, useWatch } from 'react-hook-form'
import { CampoFecha, CampoSiNo, CampoTexto } from '@/components/form/campos'
import Alert from '@/components/ui/Alert'
import Card from '@/components/ui/Card'
import type {
  Inhabilidades as InhabilidadesGuardadas,
  RemovidoDetalle,
} from '@/services/postulacion'
import {
  PREGUNTAS_INHABILIDADES,
  PREGUNTA_REMOVIDO,
  preguntasQueInhabilitan,
} from '../inhabilidades'
import PieSeccion from './PieSeccion'
import type { PropsSeccion } from './tipos'

// Los nombres del formulario son p1..p13 porque react-hook-form interpreta "1" como un
// índice de arreglo; al guardar se convierten a las claves "1".."13" de docs/api.md.
interface Valores {
  respuestas: Record<string, boolean | null>
  removido: RemovidoDetalle
}

const detalleVacio: RemovidoDetalle = { institucion: '', cargo: '', fecha_fin: '', motivo: '' }

function aValores(inhab: InhabilidadesGuardadas | undefined): Valores {
  const respuestas: Record<string, boolean | null> = {}
  PREGUNTAS_INHABILIDADES.forEach((_, i) => {
    const v = inhab?.[String(i + 1)]
    respuestas[`p${i + 1}`] = typeof v === 'boolean' ? v : null
  })
  return { respuestas, removido: { ...detalleVacio, ...inhab?.removido_detalle } }
}

function aInhabilidades(valores: Valores): InhabilidadesGuardadas {
  const salida: InhabilidadesGuardadas = {}
  PREGUNTAS_INHABILIDADES.forEach((_, i) => {
    salida[String(i + 1)] = valores.respuestas[`p${i + 1}`] === true
  })
  // Sin la pregunta 7 en Sí no hay detalle que guardar.
  salida.removido_detalle =
    valores.respuestas[`p${PREGUNTA_REMOVIDO}`] === true ? valores.removido : { ...detalleVacio }
  return salida
}

const responder = { validate: (v: unknown) => typeof v === 'boolean' || 'Responda esta pregunta' }

export default function Inhabilidades({ titulo, datos, onGuardar, onCambios }: PropsSeccion) {
  const metodos = useForm<Valores>({
    mode: 'onTouched',
    defaultValues: aValores(datos.inhabilidades),
  })
  const { handleSubmit, control, reset, formState } = metodos

  const respuestas = useWatch({ control, name: 'respuestas' })
  const removido = respuestas[`p${PREGUNTA_REMOVIDO}`] === true
  // Solo cuentan las preguntas ya respondidas, para no avisar por las que siguen vacías.
  const contestadas: InhabilidadesGuardadas = {}
  Object.entries(respuestas).forEach(([clave, valor]) => {
    if (typeof valor === 'boolean') contestadas[clave.slice(1)] = valor
  })
  const inhabilitan = preguntasQueInhabilitan(contestadas)

  useEffect(() => {
    onCambios(formState.isDirty)
  }, [formState.isDirty, onCambios])

  const guardar = handleSubmit(async (valores) => {
    const ok = await onGuardar({ inhabilidades: aInhabilidades(valores) })
    if (ok) reset(valores)
  })

  return (
    <Card titulo={titulo} descripcion="Responda las 13 preguntas. Todas son obligatorias.">
      <FormProvider {...metodos}>
        <form noValidate onSubmit={guardar} className="flex flex-col gap-space-lg">
          {inhabilitan.length > 0 && (
            <Alert tipo="warning" titulo="Respuestas que pueden inhabilitar su postulación">
              Revise las preguntas {inhabilitan.join(', ')}. Puede guardar igualmente; la revisión
              final la realiza el analista.
            </Alert>
          )}

          <ol className="flex flex-col gap-space-md">
            {PREGUNTAS_INHABILIDADES.map((texto, i) => {
              const n = i + 1
              return (
                <li key={n} className="flex flex-col gap-space-md">
                  <CampoSiNo
                    name={`respuestas.p${n}`}
                    etiqueta={`${n}. ${texto}`}
                    requerido
                    reglas={responder}
                  />
                  {n === PREGUNTA_REMOVIDO && removido && (
                    <fieldset className="grid grid-cols-1 gap-space-md rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
                      <legend className="px-1 text-label-md text-on-surface">
                        Detalle de la remoción
                      </legend>
                      <CampoTexto
                        name="removido.institucion"
                        etiqueta="Institución"
                        requerido
                        reglas={{
                          validate: (v: string) => v.trim() !== '' || 'Indique la institución',
                        }}
                      />
                      <CampoTexto name="removido.cargo" etiqueta="Cargo" />
                      <CampoFecha name="removido.fecha_fin" etiqueta="Fecha de finalización" />
                      <CampoTexto
                        name="removido.motivo"
                        etiqueta="Motivo de finalización"
                        requerido
                        reglas={{
                          validate: (v: string) => v.trim() !== '' || 'Indique el motivo',
                        }}
                      />
                    </fieldset>
                  )}
                </li>
              )
            })}
          </ol>

          <PieSeccion />
        </form>
      </FormProvider>
    </Card>
  )
}
