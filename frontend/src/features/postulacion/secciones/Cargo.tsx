import { useId, useRef, useState } from 'react'
import Campo from '@/components/form/Campo'
import DocumentoUpload from '@/components/form/DocumentoUpload'
import Alert from '@/components/ui/Alert'
import Button from '@/components/ui/Button'
import Input from '@/components/ui/Input'
import RadioSiNo from '@/components/ui/RadioSiNo'
import Spinner from '@/components/ui/Spinner'
import { ApiError } from '@/services/errores'
import type { DocumentoSubido } from '@/services/documentos'
import type { ResultadoCargo, TipoCargo } from '@/services/postulacion'
import {
  CARGOS,
  ETIQUETA_CARGO,
  MAX_MB_RESPALDO,
  MAX_NOMBRE_ORGANIZACION,
  MSG_CARGO_VALIDO,
  PREGUNTA_RESPALDO,
  erroresDeCargo,
  type ErroresCargo,
} from '../reglasCargo'

export interface RespaldoOrganizacion {
  nombre: string
  documento_id: string | null
}

// Lo que ya hay guardado en la postulación: el cargo puede estar sin elegir.
export interface CargoInicial {
  cargo: TipoCargo | ''
  respaldo_organizacion: RespaldoOrganizacion
}

export interface ValoresCargo {
  cargo: TipoCargo
  respaldo_organizacion: RespaldoOrganizacion
}

export interface CargoProps {
  valoresIniciales: CargoInicial
  // Valida el cargo con el Historial Laboral. Debe rechazar la promesa si el postulante no
  // cumple (el servidor responde 422 y cierra la postulación); quien monte la sección decide
  // cómo mostrar ese bloqueo (ver useCargo).
  onValidar: (cargo: TipoCargo) => Promise<ResultadoCargo>
  // Guarda la sección. Debe rechazar la promesa si falla; un ApiError con `fields`
  // (cargo, respaldo_organizacion.nombre, ...) se muestra bajo el botón.
  onGuardar: (valores: ValoresCargo) => Promise<void>
}

type Validacion =
  | { estado: 'validando' }
  | { estado: 'valido'; mensaje: string }
  | { estado: 'error'; mensaje: string }

const MSG_ERROR_VALIDAR = 'No se pudo validar el cargo. Inténtelo nuevamente.'

// Sección "Selección del cargo" del formulario de postulación (EF-12): un solo cargo, validado
// con el Historial Laboral, y la organización que respalda al postulante.
export default function Cargo({ valoresIniciales, onValidar, onGuardar }: CargoProps) {
  const grupoId = useId()
  const preguntaId = useId()
  const peticion = useRef(0)
  const inicial = valoresIniciales

  const [cargo, setCargo] = useState<TipoCargo | ''>(inicial.cargo)
  // Un cargo ya guardado se validó en el último guardado; el servidor vuelve a validarlo al
  // guardar. No se consulta al abrir la sección para no cerrar la postulación sin que el
  // postulante haya hecho nada.
  const [validacion, setValidacion] = useState<Validacion | null>(
    inicial.cargo ? { estado: 'valido', mensaje: MSG_CARGO_VALIDO[inicial.cargo] } : null,
  )
  const [tieneRespaldo, setTieneRespaldo] = useState<boolean | null>(
    inicial.respaldo_organizacion.nombre || inicial.respaldo_organizacion.documento_id
      ? true
      : inicial.cargo
        ? false
        : null,
  )
  const [nombre, setNombre] = useState(inicial.respaldo_organizacion.nombre)
  const [documento, setDocumento] = useState<DocumentoSubido | null>(
    inicial.respaldo_organizacion.documento_id
      ? {
          documento_id: inicial.respaldo_organizacion.documento_id,
          nombre: 'Documento de respaldo guardado',
          tamano_bytes: 0,
        }
      : null,
  )
  const [errores, setErrores] = useState<ErroresCargo>({})
  const [guardando, setGuardando] = useState(false)
  const [guardado, setGuardado] = useState(false)
  const [errorGuardar, setErrorGuardar] = useState<ApiError | null>(null)

  const validando = validacion?.estado === 'validando'
  const cargoValido = validacion?.estado === 'valido' && cargo !== '' ? cargo : null

  async function validar(elegido: TipoCargo) {
    const numero = ++peticion.current
    setCargo(elegido)
    setValidacion({ estado: 'validando' })
    setErrores({})
    setGuardado(false)
    setErrorGuardar(null)
    try {
      const resultado = await onValidar(elegido)
      // Si el postulante eligió otro cargo mientras tanto, esta respuesta ya no cuenta.
      if (numero !== peticion.current) return
      setValidacion(
        resultado.elegible
          ? { estado: 'valido', mensaje: resultado.mensaje }
          : { estado: 'error', mensaje: resultado.mensaje },
      )
    } catch (e) {
      if (numero !== peticion.current) return
      setValidacion({
        estado: 'error',
        mensaje: e instanceof ApiError ? e.message : MSG_ERROR_VALIDAR,
      })
    }
  }

  async function guardar() {
    const encontrados = erroresDeCargo({
      cargoValidado: cargoValido,
      tieneRespaldo,
      nombre,
      tieneDocumento: documento !== null,
    })
    setErrores(encontrados)
    setGuardado(false)
    setErrorGuardar(null)
    if (Object.keys(encontrados).length > 0 || cargoValido === null) return

    setGuardando(true)
    try {
      await onGuardar({
        cargo: cargoValido,
        respaldo_organizacion:
          tieneRespaldo && documento
            ? { nombre: nombre.trim(), documento_id: documento.documento_id }
            : { nombre: '', documento_id: null },
      })
      setGuardado(true)
    } catch (e) {
      setErrorGuardar(
        e instanceof ApiError
          ? e
          : new ApiError(0, 'UNKNOWN', 'No se pudo guardar la información.'),
      )
    } finally {
      setGuardando(false)
    }
  }

  const camposServidor = errorGuardar?.fields ? Object.entries(errorGuardar.fields) : []

  return (
    <section className="flex flex-col gap-5" aria-labelledby={`${grupoId}-titulo`}>
      <div>
        <h2 id={`${grupoId}-titulo`} className="text-headline-sm text-primary">
          Selección del cargo
        </h2>
        <p className="text-body-md text-on-surface-variant">
          Elija el cargo al que postula. Se verificará su calidad con el Historial Laboral del IESS.
        </p>
      </div>

      <fieldset className="flex flex-col gap-3" disabled={guardando}>
        <legend className="mb-1 text-label-md text-on-surface">
          Cargo
          <span aria-hidden="true" className="ml-0.5 font-bold text-error">
            *
          </span>
        </legend>
        {CARGOS.map((opcion) => (
          <label
            key={opcion}
            htmlFor={`${grupoId}-${opcion}`}
            className="flex cursor-pointer items-start gap-3 rounded-xl border border-outline-variant p-4 text-body-md text-on-surface has-[:checked]:border-primary-container has-[:checked]:bg-surface-container-low"
          >
            <input
              id={`${grupoId}-${opcion}`}
              type="radio"
              name={`${grupoId}-cargo`}
              value={opcion}
              checked={cargo === opcion}
              onChange={() => void validar(opcion)}
              className="mt-0.5 h-5 w-5 shrink-0 accent-primary-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
            />
            <span>{ETIQUETA_CARGO[opcion]}</span>
          </label>
        ))}
        {errores.cargo && (
          <p role="alert" className="text-label-sm text-error">
            {errores.cargo}
          </p>
        )}
      </fieldset>

      {validando && (
        <p role="status" className="flex items-center gap-2 text-body-md text-primary-container">
          <Spinner className="h-4 w-4" /> Validando con el Historial Laboral...
        </p>
      )}
      {validacion?.estado === 'valido' && <Alert tipo="success">{validacion.mensaje}</Alert>}
      {validacion?.estado === 'error' && (
        <div className="flex flex-col gap-3">
          <Alert tipo="error" titulo="No se pudo validar el cargo">
            {validacion.mensaje}
          </Alert>
          {cargo !== '' && (
            <div>
              <Button variante="secondary" onClick={() => void validar(cargo)}>
                Reintentar validación
              </Button>
            </div>
          )}
        </div>
      )}

      {cargoValido && (
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <p id={preguntaId} className="text-label-md text-on-surface">
              {PREGUNTA_RESPALDO[cargoValido]}
            </p>
            <RadioSiNo
              valor={tieneRespaldo}
              onChange={(valor) => {
                setTieneRespaldo(valor)
                setErrores((actual) => ({ ...actual, respaldo: undefined }))
                setGuardado(false)
              }}
              deshabilitado={guardando}
              aria-labelledby={preguntaId}
              aria-invalid={Boolean(errores.respaldo)}
            />
            {errores.respaldo && (
              <p role="alert" className="text-label-sm text-error">
                {errores.respaldo}
              </p>
            )}
          </div>

          {tieneRespaldo && (
            <div className="grid grid-cols-1 gap-4 rounded-xl bg-surface-container-low p-4 md:grid-cols-2">
              <Campo
                id={`${grupoId}-nombre`}
                etiqueta="Nombre de Organización:"
                requerido
                error={errores.nombre}
              >
                <Input
                  id={`${grupoId}-nombre`}
                  value={nombre}
                  maxLength={MAX_NOMBRE_ORGANIZACION}
                  onChange={(evento) => {
                    setNombre(evento.target.value)
                    setGuardado(false)
                  }}
                  disabled={guardando}
                  invalido={Boolean(errores.nombre)}
                  aria-describedby={errores.nombre ? `${grupoId}-nombre-error` : undefined}
                />
              </Campo>
              <div className="flex flex-col gap-1.5">
                <DocumentoUpload
                  tipo="organizacion"
                  maxMb={MAX_MB_RESPALDO}
                  etiqueta="Documento de respaldo *"
                  valor={documento}
                  onChange={(nuevo) => {
                    setDocumento(nuevo)
                    setErrores((actual) => ({ ...actual, documento: undefined }))
                    setGuardado(false)
                  }}
                />
                {errores.documento && (
                  <p role="alert" className="text-label-sm text-error">
                    {errores.documento}
                  </p>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      {guardado && <Alert tipo="success">La postulación se guardó correctamente.</Alert>}
      {errorGuardar && (
        <Alert tipo="error" titulo={errorGuardar.message}>
          {camposServidor.length > 0 && (
            <ul className="list-disc pl-4">
              {camposServidor.map(([campo, mensaje]) => (
                <li key={campo}>
                  {campo}: {mensaje}
                </li>
              ))}
            </ul>
          )}
        </Alert>
      )}
      <div className="flex justify-end">
        <Button
          onClick={() => void guardar()}
          cargando={guardando}
          textoCargando="Guardando..."
          disabled={validando}
        >
          Guardar Postulación
        </Button>
      </div>
    </section>
  )
}
