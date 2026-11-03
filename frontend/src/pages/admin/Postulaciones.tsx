import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Alert from '@/components/ui/Alert'
import Badge from '@/components/ui/Badge'
import Button from '@/components/ui/Button'
import Card from '@/components/ui/Card'
import Loader from '@/components/ui/Loader'
import Select from '@/components/ui/Select'
import {
  listarPostulacionesAdmin,
  type EstadoPostulacionAdmin,
  type FiltrosPostulacionesAdmin,
  type PaginaPostulacionesAdmin,
} from '@/services/adminPostulaciones'
import { ApiError } from '@/services/errores'

const etiquetasCargo = {
  afiliados: 'Afiliados',
  jubilados: 'Jubilados',
  empleador: 'Empleador',
}

const etiquetasEstado: Record<EstadoPostulacionAdmin, string> = {
  borrador: 'Borrador',
  enviada: 'Enviada',
  cerrada: 'Cerrada',
}

const opcionesCargo = Object.entries(etiquetasCargo).map(([valor, etiqueta]) => ({
  valor,
  etiqueta,
}))
const opcionesEstado = Object.entries(etiquetasEstado).map(([valor, etiqueta]) => ({
  valor,
  etiqueta,
}))
const opcionesElegible = [
  { valor: 'true', etiqueta: 'Sí' },
  { valor: 'false', etiqueta: 'No' },
]

export default function Postulaciones() {
  const [filtros, setFiltros] = useState<FiltrosPostulacionesAdmin>({ pagina: 1 })
  const [resultado, setResultado] = useState<PaginaPostulacionesAdmin>()
  const [cargando, setCargando] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    const control = new AbortController()
    listarPostulacionesAdmin(control.signal, filtros)
      .then((data) => {
        if (control.signal.aborted) return
        setResultado(data)
        setError('')
      })
      .catch((fallo: unknown) => {
        if (control.signal.aborted) return
        setError(
          fallo instanceof ApiError
            ? fallo.message
            : 'No se pudo consultar la nómina de postulantes.',
        )
      })
      .finally(() => {
        if (!control.signal.aborted) setCargando(false)
      })
    return () => control.abort()
  }, [filtros])

  const cambiarFiltro = (nombre: 'cargo' | 'estado' | 'elegible', valor: string) => {
    setCargando(true)
    setFiltros((actual) => ({ ...actual, [nombre]: valor, pagina: 1 }))
  }
  const cambiarPagina = (pagina: number) => {
    setCargando(true)
    setFiltros((actual) => ({ ...actual, pagina }))
  }
  const totalPaginas = resultado
    ? Math.max(1, Math.ceil(resultado.total / resultado.por_pagina))
    : 1

  return (
    <main className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-space-lg p-margin-mobile sm:p-margin">
      <div>
        <h1 className="text-headline-lg text-primary">Nómina de postulantes</h1>
        <p className="mt-1 text-body-md text-on-surface-variant">
          Consulte el formulario completo y los certificados de cada postulante.
        </p>
      </div>

      <Card titulo="Filtros">
        <div className="grid gap-space-md md:grid-cols-3">
          <label className="flex flex-col gap-1 text-label-md text-on-surface">
            Cargo
            <Select
              aria-label="Cargo"
              value={filtros.cargo ?? ''}
              placeholder="Todos los cargos"
              opciones={opcionesCargo}
              onChange={(e) => cambiarFiltro('cargo', e.target.value)}
            />
          </label>
          <label className="flex flex-col gap-1 text-label-md text-on-surface">
            Estado
            <Select
              aria-label="Estado"
              value={filtros.estado ?? ''}
              placeholder="Todos los estados"
              opciones={opcionesEstado}
              onChange={(e) => cambiarFiltro('estado', e.target.value)}
            />
          </label>
          <label className="flex flex-col gap-1 text-label-md text-on-surface">
            Elegible
            <Select
              aria-label="Elegible"
              value={filtros.elegible ?? ''}
              placeholder="Todos"
              opciones={opcionesElegible}
              onChange={(e) => cambiarFiltro('elegible', e.target.value)}
            />
          </label>
        </div>
      </Card>

      {error && (
        <Alert tipo="error" titulo="No se pudo cargar la nómina">
          {error}
        </Alert>
      )}
      {cargando && !resultado ? (
        <Loader texto="Cargando postulantes..." />
      ) : (
        resultado && (
          <Card
            titulo={`${resultado.total} postulante${resultado.total === 1 ? '' : 's'}`}
            descripcion={`Página ${resultado.pagina} de ${totalPaginas}`}
          >
            <div className="overflow-x-auto">
              <table className="w-full min-w-[800px] border-collapse text-left text-body-md">
                <thead>
                  <tr className="border-b border-outline-variant text-label-md text-on-surface-variant">
                    <th className="px-3 py-2">Código</th>
                    <th className="px-3 py-2">Postulante</th>
                    <th className="px-3 py-2">Cédula</th>
                    <th className="px-3 py-2">Cargo</th>
                    <th className="px-3 py-2">Estado</th>
                    <th className="px-3 py-2">Elegible</th>
                    <th className="px-3 py-2 text-right">Acción</th>
                  </tr>
                </thead>
                <tbody>
                  {resultado.items.map((item) => (
                    <tr key={item.id} className="border-b border-outline-variant/60">
                      <td className="px-3 py-3 font-mono text-label-md">{item.codigo}</td>
                      <td className="px-3 py-3">
                        {item.nombres} {item.apellidos}
                      </td>
                      <td className="px-3 py-3">{item.cedula}</td>
                      <td className="px-3 py-3">
                        {item.cargo ? etiquetasCargo[item.cargo] : 'Sin seleccionar'}
                      </td>
                      <td className="px-3 py-3">
                        <Badge tono={item.estado === 'enviada' ? 'success' : 'neutral'}>
                          {etiquetasEstado[item.estado]}
                        </Badge>
                      </td>
                      <td className="px-3 py-3">
                        <Badge tono={item.elegible ? 'success' : 'warning'}>
                          {item.elegible ? 'Sí' : 'No'}
                        </Badge>
                      </td>
                      <td className="px-3 py-3 text-right">
                        <Link
                          className="text-body-bold text-primary-container hover:underline"
                          to={`/admin/postulaciones/${item.id}`}
                        >
                          Ver detalle
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {resultado.items.length === 0 && (
                <p className="py-space-xl text-center text-body-md text-on-surface-variant">
                  No hay postulantes con estos filtros.
                </p>
              )}
            </div>
            <div className="mt-space-lg flex items-center justify-between gap-3">
              <Button
                variante="secondary"
                icono="chevron_left"
                disabled={resultado.pagina <= 1 || cargando}
                onClick={() => cambiarPagina(Math.max(1, resultado.pagina - 1))}
              >
                Anterior
              </Button>
              <span className="text-label-md text-on-surface-variant">
                Página {resultado.pagina} de {totalPaginas}
              </span>
              <Button
                variante="secondary"
                icono="chevron_right"
                disabled={resultado.pagina >= totalPaginas || cargando}
                onClick={() => cambiarPagina(resultado.pagina + 1)}
              >
                Siguiente
              </Button>
            </div>
          </Card>
        )
      )}
    </main>
  )
}
