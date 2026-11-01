import type { ReactNode } from 'react'
import { cx } from '@/lib/cx'
import Button from './Button'
import Icon from './Icon'

export interface ColumnaTabla<T> {
  clave: string
  titulo: string
  render?: (fila: T) => ReactNode
}

interface TablaFilasProps<T> {
  titulo?: string
  columnas: ColumnaTabla<T>[]
  filas: T[]
  obtenerClave: (fila: T) => string | number
  onAgregar?: () => void
  onEditar?: (fila: T) => void
  onEliminar?: (fila: T) => void
  textoAgregar?: string
  textoVacio?: string
  deshabilitado?: boolean
}

// Tabla genérica para las secciones con filas: Agregar, Editar y Eliminar.
export default function TablaFilas<T>({
  titulo,
  columnas,
  filas,
  obtenerClave,
  onAgregar,
  onEditar,
  onEliminar,
  textoAgregar = 'Agregar',
  textoVacio = 'Aún no hay registros.',
  deshabilitado,
}: TablaFilasProps<T>) {
  const hayAcciones = Boolean(onEditar || onEliminar)
  return (
    <div className="flex flex-col gap-4">
      {(titulo || onAgregar) && (
        <div className="flex flex-col justify-between gap-4 rounded-xl bg-surface-container-lowest p-4 shadow-sm sm:flex-row sm:items-center">
          <div className="flex flex-col">
            {titulo && <span className="text-body-bold text-on-surface">{titulo}</span>}
            <span className="text-label-sm text-on-surface-variant">
              {filas.length} {filas.length === 1 ? 'registro' : 'registros'}
            </span>
          </div>
          {onAgregar && (
            <Button icono="add" onClick={onAgregar} disabled={deshabilitado}>
              {textoAgregar}
            </Button>
          )}
        </div>
      )}
      <div className="overflow-x-auto rounded-xl bg-surface-container-lowest shadow-sm">
        <table className="w-full text-left text-body-md">
          <thead>
            <tr className="bg-surface-container-low text-legal-code uppercase text-on-surface-variant">
              {columnas.map((c) => (
                <th key={c.clave} scope="col" className="px-4 py-3.5 font-semibold tracking-wider">
                  {c.titulo}
                </th>
              ))}
              {hayAcciones && (
                <th scope="col" className="px-4 py-3.5 text-right font-semibold tracking-wider">
                  Acciones
                </th>
              )}
            </tr>
          </thead>
          <tbody>
            {filas.length === 0 && (
              <tr>
                <td
                  colSpan={columnas.length + (hayAcciones ? 1 : 0)}
                  className="px-4 py-8 text-center text-on-surface-variant"
                >
                  {textoVacio}
                </td>
              </tr>
            )}
            {filas.map((fila, i) => (
              <tr
                key={obtenerClave(fila)}
                className={cx(
                  'transition-colors hover:bg-surface-container',
                  i % 2 === 1 && 'bg-surface-container-low',
                )}
              >
                {columnas.map((c) => (
                  <td key={c.clave} className="px-4 py-4 text-on-surface">
                    {c.render
                      ? c.render(fila)
                      : String((fila as Record<string, unknown>)[c.clave] ?? '')}
                  </td>
                ))}
                {hayAcciones && (
                  <td className="px-4 py-4 text-right">
                    <div className="inline-flex items-center gap-1">
                      {onEditar && (
                        <button
                          type="button"
                          aria-label="Editar"
                          title="Editar"
                          disabled={deshabilitado}
                          onClick={() => onEditar(fila)}
                          className="rounded-lg p-2 text-primary-container hover:bg-primary-fixed focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary disabled:opacity-50"
                        >
                          <Icon nombre="edit" className="text-[18px]" />
                        </button>
                      )}
                      {onEliminar && (
                        <button
                          type="button"
                          aria-label="Eliminar"
                          title="Eliminar"
                          disabled={deshabilitado}
                          onClick={() => onEliminar(fila)}
                          className="rounded-lg p-2 text-error hover:bg-error-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary disabled:opacity-50"
                        >
                          <Icon nombre="delete" className="text-[18px]" />
                        </button>
                      )}
                    </div>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
