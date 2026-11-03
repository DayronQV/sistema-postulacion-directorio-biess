import { useCallback, useEffect, useState, type FormEvent } from 'react'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Input from '@/components/ui/Input'
import Loader from '@/components/ui/Loader'
import { ModalConfirmacion } from '@/components/ui/Modal'
import CargaMasiva from '@/features/evaluacion/admin/CargaMasiva'
import FormularioPregunta from '@/features/evaluacion/admin/FormularioPregunta'
import {
  actualizarPregunta,
  crearPregunta,
  eliminarPregunta,
  listarPreguntas,
  type PaginaPreguntas,
  type PreguntaBanco,
  type PreguntaNueva,
} from '@/services/adminEvaluacion'
import { ApiError } from '@/services/errores'
import EncabezadoPagina from '@/components/layout/EncabezadoPagina'

const POR_PAGINA = 20

// Un 400 del servidor trae las claves de `fields` como `opciones[1]`: son las mismas del formulario.
function mensaje(e: unknown, porDefecto: string) {
  return e instanceof ApiError ? e.message : porDefecto
}

// Banco de preguntas de la evaluación teórica (EF-14): listado con búsqueda, alta, edición,
// eliminación y carga masiva desde CSV. Solo el administrador (el servidor responde 403 al analista).
export default function EvaluacionBanco() {
  const [buscar, setBuscar] = useState('')
  const [consulta, setConsulta] = useState('')
  const [pagina, setPagina] = useState(1)
  const [resultado, setResultado] = useState<PaginaPreguntas>()
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')
  const [version, setVersion] = useState(0)

  const [formulario, setFormulario] = useState<{ pregunta: PreguntaBanco | null } | null>(null)
  const [guardando, setGuardando] = useState(false)
  const [erroresForm, setErroresForm] = useState<Record<string, string>>({})
  const [errorForm, setErrorForm] = useState('')
  const [porEliminar, setPorEliminar] = useState<PreguntaBanco | null>(null)
  const [eliminando, setEliminando] = useState(false)
  const [errorEliminar, setErrorEliminar] = useState('')
  const [aviso, setAviso] = useState('')
  const [cargaAbierta, setCargaAbierta] = useState(false)

  useEffect(() => {
    const control = new AbortController()
    listarPreguntas(control.signal, { buscar: consulta, pagina, por_pagina: POR_PAGINA })
      .then((data) => {
        if (control.signal.aborted) return
        // Tras eliminar la última pregunta de la última página, esa página ya no existe:
        // se vuelve a la última que tiene preguntas.
        if (data.items.length === 0 && data.total > 0 && pagina > 1) {
          setPagina(Math.ceil(data.total / data.por_pagina))
          return
        }
        setResultado(data)
        setError('')
      })
      .catch((e: unknown) => {
        if (control.signal.aborted) return
        setError(mensaje(e, 'No se pudo consultar el banco de preguntas.'))
      })
      .finally(() => {
        if (!control.signal.aborted) setCargando(false)
      })
    return () => control.abort()
  }, [consulta, pagina, version])

  const recargar = useCallback(() => {
    setCargando(true)
    setVersion((v) => v + 1)
  }, [])

  function buscarPreguntas(e: FormEvent) {
    e.preventDefault()
    setCargando(true)
    setPagina(1)
    setConsulta(buscar.trim())
  }

  function cerrarFormulario() {
    setFormulario(null)
    setErroresForm({})
    setErrorForm('')
  }

  async function guardar(pregunta: PreguntaNueva) {
    setGuardando(true)
    setErroresForm({})
    setErrorForm('')
    try {
      const control = new AbortController()
      if (formulario?.pregunta) {
        await actualizarPregunta(control.signal, formulario.pregunta.id, pregunta)
        setAviso('Pregunta actualizada.')
      } else {
        await crearPregunta(control.signal, pregunta)
        setAviso('Pregunta agregada al banco.')
      }
      cerrarFormulario()
      recargar()
    } catch (e) {
      setErroresForm(e instanceof ApiError ? (e.fields ?? {}) : {})
      setErrorForm(mensaje(e, 'No se pudo guardar la pregunta.'))
    } finally {
      setGuardando(false)
    }
  }

  async function eliminar() {
    if (!porEliminar) return
    setEliminando(true)
    setErrorEliminar('')
    try {
      await eliminarPregunta(new AbortController().signal, porEliminar.id)
      setPorEliminar(null)
      setAviso('Pregunta eliminada.')
      recargar()
    } catch (e) {
      setErrorEliminar(mensaje(e, 'No se pudo eliminar la pregunta.'))
    } finally {
      setEliminando(false)
    }
  }

  const totalPaginas = resultado
    ? Math.max(1, Math.ceil(resultado.total / resultado.por_pagina))
    : 1

  return (
    <main className="mx-auto flex w-full max-w-[1400px] flex-1 flex-col gap-space-md p-margin-mobile sm:p-margin">
      <EncabezadoPagina
        icono="quiz"
        titulo="Banco de preguntas"
        descripcion="Preguntas de opción múltiple de la evaluación teórica."
        acciones={
          <>
            <Button
              variante="secondary"
              icono="upload_file"
              onClick={() => setCargaAbierta((abierta) => !abierta)}
            >
              Carga masiva
            </Button>
            <Button icono="add" onClick={() => setFormulario({ pregunta: null })}>
              Nueva pregunta
            </Button>
          </>
        }
      />

      {aviso && (
        <Alert tipo="success" onCerrar={() => setAviso('')}>
          {aviso}
        </Alert>
      )}
      {cargaAbierta && <CargaMasiva onCargado={recargar} />}

      <Card>
        <form onSubmit={buscarPreguntas} className="flex flex-wrap items-end gap-3">
          <label className="flex min-w-[16rem] flex-1 flex-col gap-1 text-label-md text-on-surface">
            Buscar en el enunciado
            <Input
              value={buscar}
              placeholder="Texto de la pregunta"
              onChange={(e) => setBuscar(e.target.value)}
            />
          </label>
          <Button type="submit" variante="secondary" icono="search">
            Buscar
          </Button>
        </form>
      </Card>

      {error && (
        <Alert tipo="error" titulo="No se pudo cargar el banco">
          {error}
        </Alert>
      )}
      {cargando && !resultado ? (
        <Loader texto="Cargando preguntas..." />
      ) : (
        resultado && (
          <Card
            titulo={`${resultado.total} pregunta${resultado.total === 1 ? '' : 's'}`}
            descripcion={`Página ${resultado.pagina} de ${totalPaginas}`}
          >
            <div className="overflow-x-auto">
              <table className="w-full min-w-[720px] border-collapse text-left text-body-md">
                <thead>
                  <tr className="border-b border-outline-variant text-label-md text-on-surface-variant">
                    <th className="px-3 py-2">Enunciado</th>
                    <th className="px-3 py-2">Opciones</th>
                    <th className="px-3 py-2">Respuesta correcta</th>
                    <th className="px-3 py-2 text-right">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {resultado.items.map((p) => (
                    <tr key={p.id} className="border-b border-outline-variant/60 align-top">
                      <td className="px-3 py-3">{p.enunciado}</td>
                      <td className="px-3 py-3">{p.opciones.length}</td>
                      <td className="px-3 py-3">{p.opciones[p.respuesta_correcta]}</td>
                      <td className="px-3 py-3 text-right">
                        <div className="flex justify-end gap-2">
                          <Button
                            variante="ghost"
                            icono="edit"
                            aria-label={`Editar la pregunta ${p.id}`}
                            onClick={() => setFormulario({ pregunta: p })}
                            className="px-3"
                          />
                          <Button
                            variante="ghost"
                            icono="delete"
                            aria-label={`Eliminar la pregunta ${p.id}`}
                            onClick={() => {
                              setErrorEliminar('')
                              setPorEliminar(p)
                            }}
                            className="px-3"
                          />
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {resultado.items.length === 0 && (
                <p className="py-space-xl text-center text-body-md text-on-surface-variant">
                  {consulta
                    ? 'Ninguna pregunta coincide con la búsqueda.'
                    : 'El banco está vacío. Agregue preguntas o cárguelas desde un archivo.'}
                </p>
              )}
            </div>
            <div className="mt-space-lg flex items-center justify-between gap-3">
              <Button
                variante="secondary"
                icono="chevron_left"
                disabled={resultado.pagina <= 1 || cargando}
                onClick={() => {
                  setCargando(true)
                  setPagina(resultado.pagina - 1)
                }}
              >
                Anterior
              </Button>
              <span className="text-label-md text-on-surface-variant">
                {resultado.total} en total
              </span>
              <Button
                variante="secondary"
                disabled={resultado.pagina >= totalPaginas || cargando}
                onClick={() => {
                  setCargando(true)
                  setPagina(resultado.pagina + 1)
                }}
              >
                Siguiente
              </Button>
            </div>
          </Card>
        )
      )}

      {formulario && (
        <FormularioPregunta
          key={formulario.pregunta?.id ?? 'nueva'}
          pregunta={formulario.pregunta}
          guardando={guardando}
          erroresServidor={erroresForm}
          errorGeneral={errorForm}
          onGuardar={(p) => void guardar(p)}
          onCerrar={cerrarFormulario}
        />
      )}

      <ModalConfirmacion
        abierto={porEliminar !== null}
        titulo="Eliminar pregunta"
        mensaje={
          <div className="flex flex-col gap-space-sm">
            <p>¿Eliminar esta pregunta del banco? No se puede deshacer.</p>
            {porEliminar && <p className="font-semibold">{porEliminar.enunciado}</p>}
            {errorEliminar && <Alert tipo="error">{errorEliminar}</Alert>}
          </div>
        }
        textoConfirmar="Eliminar"
        peligroso
        cargando={eliminando}
        onConfirmar={() => void eliminar()}
        onCancelar={() => setPorEliminar(null)}
      />
    </main>
  )
}
