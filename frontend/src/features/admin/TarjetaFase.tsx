import { useState, type FormEvent } from 'react'
import Campo from '@/components/form/Campo'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Checkbox from '@/components/ui/Checkbox'
import Input from '@/components/ui/Input'
import { useApi } from '@/hooks/useApi'
import {
  actualizarFase,
  FASE_POSTULACION,
  type CambiosFase,
  type Fase,
} from '@/services/adminFases'

interface TarjetaFaseProps {
  fase: Fase
  // Solo el administrador cambia la fase; el analista la ve sin controles activos.
  editable: boolean
  onGuardada?: (fase: Fase) => void
}

const llamada = (signal: AbortSignal, id: number, cambios: CambiosFase) =>
  actualizarFase(signal, id, cambios)

// RFC 3339 → valor de <input type="datetime-local"> en la hora local ("2026-11-01T08:00").
function aLocal(iso: string | null) {
  if (!iso) return ''
  const d = new Date(iso)
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`
}

// Valor de datetime-local (hora local) → RFC 3339 en UTC; vacío es sin fecha.
function aIso(local: string) {
  return local ? new Date(local).toISOString() : null
}

function rango(fase: Fase) {
  const f = (iso: string) =>
    new Date(iso).toLocaleString('es-EC', { dateStyle: 'medium', timeStyle: 'short' })
  if (fase.fecha_inicio && fase.fecha_fin)
    return `Del ${f(fase.fecha_inicio)} al ${f(fase.fecha_fin)}`
  if (fase.fecha_inicio) return `Desde el ${f(fase.fecha_inicio)}, sin fecha de cierre`
  if (fase.fecha_fin) return `Hasta el ${f(fase.fecha_fin)}`
  return 'Sin fechas definidas'
}

// Una fase del concurso (EF-13): publicarla, fijar su periodo y, en la fase Postulación,
// habilitar el formulario EF-12. El servidor valida las mismas reglas.
export default function TarjetaFase({ fase, editable, onGuardada }: TarjetaFaseProps) {
  const esPostulacion = fase.nombre === FASE_POSTULACION
  const [publicada, setPublicada] = useState(fase.publicada)
  const [habilita, setHabilita] = useState(fase.habilita_formulario)
  const [inicio, setInicio] = useState(aLocal(fase.fecha_inicio))
  const [fin, setFin] = useState(aLocal(fase.fecha_fin))
  const [guardada, setGuardada] = useState(false)
  // Lo último que respondió el servidor: el encabezado muestra lo guardado, no lo editado.
  const [vigente, setVigente] = useState(fase)
  const { ejecutar, cargando, error } = useApi(llamada)

  const errorFin =
    inicio && fin && fin < inicio
      ? 'La fecha de fin no puede ser anterior a la de inicio'
      : undefined
  const id = (campo: string) => `fase-${fase.id}-${campo}`

  const guardar = async (e: FormEvent) => {
    e.preventDefault()
    if (errorFin) return
    setGuardada(false)
    const r = await ejecutar(fase.id, {
      publicada,
      habilita_formulario: esPostulacion && publicada && habilita,
      fecha_inicio: aIso(inicio),
      fecha_fin: aIso(fin),
    })
    if (r) {
      setVigente(r)
      setGuardada(true)
      onGuardada?.(r)
    }
  }

  return (
    <Card
      titulo={`${fase.orden}. ${fase.nombre}`}
      descripcion={rango(vigente)}
      acciones={
        <Badge tono={vigente.publicada ? 'success' : 'neutral'}>
          {vigente.publicada ? 'Publicada' : 'Sin publicar'}
        </Badge>
      }
    >
      <form className="flex flex-col gap-space-md" onSubmit={(e) => void guardar(e)} noValidate>
        <fieldset disabled={!editable || cargando} className="flex flex-col gap-space-md">
          <legend className="sr-only">Configuración de la fase {fase.nombre}</legend>
          <Checkbox
            id={id('publicada')}
            etiqueta="Fase publicada"
            checked={publicada}
            onChange={(e) => {
              setPublicada(e.target.checked)
              if (!e.target.checked) setHabilita(false)
            }}
          />
          {esPostulacion && (
            <Checkbox
              id={id('habilita')}
              etiqueta="Habilitar el formulario de postulación"
              checked={habilita}
              disabled={!publicada}
              invalido={!!error?.fields?.habilita_formulario}
              onChange={(e) => setHabilita(e.target.checked)}
            />
          )}
          <div className="grid gap-space-md sm:grid-cols-2">
            <Campo id={id('inicio')} etiqueta="Fecha de inicio" error={error?.fields?.fecha_inicio}>
              <Input
                id={id('inicio')}
                type="datetime-local"
                value={inicio}
                onChange={(e) => setInicio(e.target.value)}
              />
            </Campo>
            <Campo
              id={id('fin')}
              etiqueta="Fecha de fin"
              error={errorFin ?? error?.fields?.fecha_fin}
              ayuda="Sin fecha, la fase no tiene límite por ese lado."
            >
              <Input
                id={id('fin')}
                type="datetime-local"
                value={fin}
                invalido={!!errorFin}
                onChange={(e) => setFin(e.target.value)}
              />
            </Campo>
          </div>
        </fieldset>

        {error && <Alert tipo="error">{error.fields?.habilita_formulario ?? error.message}</Alert>}
        {guardada && !error && <Alert tipo="success">Fase guardada.</Alert>}

        {editable && (
          <div className="flex justify-end">
            <Button type="submit" icono="save" cargando={cargando} textoCargando="Guardando...">
              Guardar
            </Button>
          </div>
        )}
      </form>
    </Card>
  )
}
