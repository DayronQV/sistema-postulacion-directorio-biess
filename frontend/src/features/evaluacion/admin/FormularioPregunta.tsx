import { useState, type FormEvent } from 'react'
import Campo from '@/components/form/Campo'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import Modal from '@/components/ui/Modal'
import Textarea from '@/components/ui/Textarea'
import type { PreguntaBanco, PreguntaNueva } from '@/services/adminEvaluacion'
import { MAX_ENUNCIADO, MAX_OPCION, MAX_OPCIONES, MIN_OPCIONES, validarPregunta } from './preguntas'

interface FormularioPreguntaProps {
  // Pregunta que se edita; sin ella, el formulario crea una nueva.
  pregunta?: PreguntaBanco | null
  guardando: boolean
  // Errores que devolvió el servidor, con las mismas claves que validarPregunta.
  erroresServidor?: Record<string, string>
  // Mensaje general del servidor (por ejemplo, 409 si ya hay evaluaciones iniciadas).
  errorGeneral?: string
  onGuardar: (pregunta: PreguntaNueva) => void
  onCerrar: () => void
}

// Formulario de una pregunta: enunciado, opciones dinámicas (agregar y quitar) y la marca de
// la correcta. Se monta solo mientras está abierto, así cada apertura parte de cero.
export default function FormularioPregunta({
  pregunta,
  guardando,
  erroresServidor,
  errorGeneral,
  onGuardar,
  onCerrar,
}: FormularioPreguntaProps) {
  const [enunciado, setEnunciado] = useState(pregunta?.enunciado ?? '')
  const [opciones, setOpciones] = useState<string[]>(pregunta?.opciones ?? ['', ''])
  const [correcta, setCorrecta] = useState<number | null>(pregunta?.respuesta_correcta ?? null)
  const [errores, setErrores] = useState<Record<string, string>>({})
  const visibles = { ...erroresServidor, ...errores }

  function cambiarOpcion(i: number, valor: string) {
    setOpciones((actual) => actual.map((o, j) => (j === i ? valor : o)))
  }

  function quitarOpcion(i: number) {
    setOpciones((actual) => actual.filter((_, j) => j !== i))
    // La marca sigue a su opción: si se quita la correcta se pierde; si se quita una anterior se corre.
    setCorrecta((actual) => {
      if (actual === null || actual === i) return null
      return actual > i ? actual - 1 : actual
    })
  }

  function enviar(e: FormEvent) {
    e.preventDefault()
    const encontrados = validarPregunta({ enunciado, opciones, respuesta_correcta: correcta })
    setErrores(encontrados)
    if (Object.keys(encontrados).length > 0 || correcta === null) return
    onGuardar({
      enunciado: enunciado.trim(),
      opciones: opciones.map((o) => o.trim()),
      respuesta_correcta: correcta,
    })
  }

  return (
    <Modal
      abierto
      titulo={pregunta ? 'Editar pregunta' : 'Nueva pregunta'}
      onCerrar={onCerrar}
      bloqueado={guardando}
      pie={
        <>
          <Button variante="secondary" onClick={onCerrar} disabled={guardando}>
            Cancelar
          </Button>
          <Button type="submit" form="formulario-pregunta" cargando={guardando}>
            Guardar pregunta
          </Button>
        </>
      }
    >
      <form
        id="formulario-pregunta"
        noValidate
        onSubmit={enviar}
        className="flex flex-col gap-space-md"
      >
        {errorGeneral && (
          <Alert tipo="error" titulo="No se pudo guardar">
            {errorGeneral}
          </Alert>
        )}
        <Campo id="pregunta-enunciado" etiqueta="Enunciado" requerido error={visibles.enunciado}>
          <Textarea
            id="pregunta-enunciado"
            value={enunciado}
            maxLength={MAX_ENUNCIADO}
            invalido={Boolean(visibles.enunciado)}
            onChange={(e) => setEnunciado(e.target.value)}
          />
        </Campo>

        <fieldset className="flex flex-col gap-space-sm">
          <legend className="text-label-md text-on-surface">
            Opciones <span className="font-bold text-error">*</span>
            <span className="ml-2 text-legal-code text-on-surface-variant">
              El botón redondo marca la correcta
            </span>
          </legend>
          {opciones.map((opcion, i) => (
            <div key={i} className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <input
                  type="radio"
                  name="opcion-correcta"
                  aria-label={`Opción ${i + 1} es la correcta`}
                  checked={correcta === i}
                  onChange={() => setCorrecta(i)}
                  className="h-5 w-5 shrink-0 accent-primary-container"
                />
                <Input
                  aria-label={`Opción ${i + 1}`}
                  value={opcion}
                  maxLength={MAX_OPCION}
                  invalido={Boolean(visibles[`opciones[${i}]`])}
                  onChange={(e) => cambiarOpcion(i, e.target.value)}
                />
                <Button
                  variante="ghost"
                  icono="delete"
                  aria-label={`Quitar la opción ${i + 1}`}
                  disabled={opciones.length <= MIN_OPCIONES}
                  onClick={() => quitarOpcion(i)}
                  className="px-3"
                />
              </div>
              {visibles[`opciones[${i}]`] && (
                <p role="alert" className="text-label-sm text-error">
                  {visibles[`opciones[${i}]`]}
                </p>
              )}
            </div>
          ))}
          {visibles.opciones && (
            <p role="alert" className="text-label-sm text-error">
              {visibles.opciones}
            </p>
          )}
          {visibles.respuesta_correcta && (
            <p role="alert" className="text-label-sm text-error">
              {visibles.respuesta_correcta}
            </p>
          )}
          <div>
            <Button
              variante="secondary"
              icono="add"
              disabled={opciones.length >= MAX_OPCIONES}
              onClick={() => setOpciones((actual) => [...actual, ''])}
            >
              Agregar opción
            </Button>
          </div>
        </fieldset>
      </form>
    </Modal>
  )
}
