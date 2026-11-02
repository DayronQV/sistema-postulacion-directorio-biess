import type { ReactNode } from 'react'
import type { Postulacion } from '@/services/postulacion'
import { PREGUNTAS_INHABILIDADES, PREGUNTA_REMOVIDO } from '../inhabilidades'
import { formatearDuracion, formatearFecha, formatearFechaHoy, siNo, texto } from './formato'
import { CARGOS, CONSENTIMIENTO, etiqueta } from './textos'
import type { Ubicacion } from './useUbicacion'

interface DocumentoProps {
  postulacion: Postulacion
  // Nombres de país, provincia y ciudad (la API guarda ids de catálogo).
  ubicacion?: Ubicacion
  fecha?: Date
}

function Bloque({ titulo, children }: { titulo: string; children: ReactNode }) {
  return (
    <section className="break-inside-avoid">
      <h2 className="mb-2 border-b border-primary pb-1 text-body-bold uppercase tracking-wide text-primary">
        {titulo}
      </h2>
      {children}
    </section>
  )
}

function Datos({ filas }: { filas: Array<[string, string]> }) {
  return (
    <dl className="grid grid-cols-1 gap-x-6 gap-y-1 text-body-md sm:grid-cols-2">
      {filas.map(([etiqueta, valor]) => (
        <div key={etiqueta} className="flex gap-2">
          <dt className="font-semibold text-on-surface">{etiqueta}:</dt>
          <dd className="text-on-surface">{valor}</dd>
        </div>
      ))}
    </dl>
  )
}

function Fila({ n, children }: { n: number; children: ReactNode }) {
  return (
    <li className="break-inside-avoid text-body-md">
      <span className="font-semibold">{n}.</span> {children}
    </li>
  )
}

function Vacio() {
  return <p className="text-body-md italic text-on-surface-variant">Sin registros.</p>
}

// Formulario final en formato de documento, pensado para imprimirse en A4 y firmarse a mano.
export default function DocumentoPostulacion({
  postulacion: p,
  ubicacion,
  fecha = new Date(),
}: DocumentoProps) {
  const personal = p.informacion_personal
  const dom = p.domicilio
  const nombreCompleto = personal ? `${personal.nombres} ${personal.apellidos}` : ''
  const inhab = p.inhabilidades

  return (
    <article
      aria-label="Formulario de postulación"
      className="mx-auto flex w-full max-w-[210mm] flex-col gap-space-md bg-white p-8 text-on-surface shadow-md print:max-w-none print:p-0 print:shadow-none"
    >
      <header className="flex flex-col items-center gap-1 border-b-2 border-primary pb-4 text-center">
        <p className="text-label-md uppercase tracking-wider text-on-surface-variant">
          Instituto Ecuatoriano de Seguridad Social
        </p>
        <h1 className="text-headline-md text-primary">
          Formulario de postulación al Directorio del BIESS
        </h1>
        <p className="text-body-md text-on-surface-variant">
          Concurso de Méritos y Oposición para la selección de Miembros Principales y Alternos
        </p>
        <div className="mt-2 flex flex-wrap justify-center gap-x-8 gap-y-1 text-body-md">
          <span>
            <strong>Código único:</strong> <span className="font-mono">{texto(p.codigo)}</span>
          </span>
          <span>
            <strong>Fecha:</strong> {formatearFechaHoy(fecha)}
          </span>
        </div>
      </header>

      <Bloque titulo="1. Información personal">
        <Datos
          filas={[
            ['Nombres y apellidos', texto(nombreCompleto)],
            ['Cédula', texto(personal?.cedula)],
            ['Género', texto(personal?.genero)],
            ['Fecha de nacimiento', formatearFecha(personal?.fecha_nacimiento)],
            ['Nacionalidad', texto(personal?.nacionalidad)],
            ['Estado civil', texto(personal?.estado_civil)],
            ...(personal?.pareja
              ? ([
                  ['Nombres y apellidos de la pareja', texto(personal.pareja.nombres_apellidos)],
                  ['Cédula de la pareja', texto(personal.pareja.cedula)],
                ] as Array<[string, string]>)
              : []),
          ]}
        />
      </Bloque>

      <Bloque titulo="2. Información domiciliaria y de contacto">
        <Datos
          filas={[
            ['País', texto(ubicacion?.pais)],
            ['Provincia / Estado', texto(ubicacion?.provincia)],
            ['Ciudad', texto(ubicacion?.ciudad)],
            ['Teléfono celular', texto(dom?.telefono_celular)],
            ['Correo electrónico 1', texto(personal?.email)],
            ['Correo electrónico 2', texto(dom?.email_secundario)],
            ['Calle principal', texto(dom?.calle_principal)],
            ['Número', texto(dom?.numero)],
            ['Calle secundaria', texto(dom?.calle_secundaria)],
            ['Sector', texto(dom?.sector)],
            ['Referencia', texto(dom?.referencia)],
            ['Teléfono de domicilio', texto(dom?.telefono_domicilio)],
          ]}
        />
      </Bloque>

      <Bloque titulo="3. Información adicional">
        <p className="mb-2 text-body-md">
          ¿Tiene familiares hasta el 2.º grado de consanguinidad o 4.º de afinidad trabajando en
          IESS o BIESS? <strong>{siNo(p.tiene_familiares)}</strong>
        </p>
        {p.tiene_familiares && (p.familiares?.length ?? 0) > 0 ? (
          <ol className="flex flex-col gap-1">
            {p.familiares?.map((f, i) => (
              <Fila key={i} n={i + 1}>
                {f.nombres_apellidos} — {f.parentesco}, {f.institucion}, {texto(f.area)} /{' '}
                {texto(f.cargo)}
              </Fila>
            ))}
          </ol>
        ) : null}
      </Bloque>

      <Bloque titulo="4. Selección del cargo">
        <p className="text-body-md">{p.cargo ? CARGOS[p.cargo] : '—'}</p>
        {p.respaldo_organizacion?.nombre && (
          <p className="text-body-md">
            <strong>Organización de respaldo:</strong> {p.respaldo_organizacion.nombre}
          </p>
        )}
      </Bloque>

      <Bloque titulo="5. Formación académica">
        {(p.formacion?.length ?? 0) > 0 ? (
          <ol className="flex flex-col gap-1">
            {p.formacion?.map((f, i) => (
              <Fila key={i} n={i + 1}>
                {f.titulo} ({etiqueta(f.nivel)}) — {f.institucion}, {etiqueta(f.tipo)}. Área:{' '}
                {etiqueta(f.area)}. SENESCYT N.º {texto(f.senescyt_numero)},{' '}
                {formatearFecha(f.senescyt_fecha)}.
              </Fila>
            ))}
          </ol>
        ) : (
          <Vacio />
        )}
      </Bloque>

      <Bloque titulo="6. Experiencia profesional">
        {(p.experiencia?.length ?? 0) > 0 ? (
          <ol className="flex flex-col gap-2">
            {p.experiencia?.map((e, i) => (
              <Fila key={i} n={i + 1}>
                {e.cargo} — {e.institucion} ({etiqueta(e.tipo)}). Del{' '}
                {formatearFecha(e.fecha_inicio)} al {formatearFecha(e.fecha_fin)}.{' '}
                <strong>Tiempo: {formatearDuracion(e.duracion)}.</strong> Actividades relacionadas a{' '}
                {etiqueta(e.actividades)}. {e.descripcion}
              </Fila>
            ))}
            <li className="text-body-md">
              <strong>
                Tiempo total de experiencia: {formatearDuracion(p.duracion_experiencia_total)}.
              </strong>
            </li>
          </ol>
        ) : (
          <Vacio />
        )}
      </Bloque>

      <Bloque titulo="7. Publicaciones">
        <p className="mb-2 text-body-md">
          ¿Cuenta con publicaciones e investigaciones académicas?{' '}
          <strong>{siNo(p.tiene_publicaciones)}</strong>
        </p>
        {p.tiene_publicaciones && (p.publicaciones?.length ?? 0) > 0 ? (
          <ol className="flex flex-col gap-1">
            {p.publicaciones?.map((pub, i) => (
              <Fila key={i} n={i + 1}>
                {texto(pub.titulo)} ({etiqueta(pub.tipo)}) — {texto(pub.medio)},{' '}
                {formatearFecha(pub.fecha)}
              </Fila>
            ))}
          </ol>
        ) : null}
      </Bloque>

      <Bloque titulo="8. Capacitación">
        {(p.capacitaciones?.length ?? 0) > 0 ? (
          <ol className="flex flex-col gap-1">
            {p.capacitaciones?.map((c, i) => (
              <Fila key={i} n={i + 1}>
                {c.nombre} ({etiqueta(c.tipo)}) — {c.institucion}. Del{' '}
                {formatearFecha(c.fecha_inicio)} al {formatearFecha(c.fecha_fin)}, {c.horas} horas.
              </Fila>
            ))}
          </ol>
        ) : (
          <Vacio />
        )}
      </Bloque>

      <Bloque titulo="9. Inhabilidades">
        <ol className="flex flex-col gap-1">
          {PREGUNTAS_INHABILIDADES.map((pregunta, i) => {
            const n = i + 1
            const r = inhab?.[String(n)]
            return (
              <Fila key={n} n={n}>
                {pregunta} <strong>{siNo(typeof r === 'boolean' ? r : undefined)}</strong>
                {n === PREGUNTA_REMOVIDO && r === true && inhab?.removido_detalle && (
                  <span>
                    {' '}
                    — {inhab.removido_detalle.institucion}, {texto(inhab.removido_detalle.cargo)},{' '}
                    {formatearFecha(inhab.removido_detalle.fecha_fin)}:{' '}
                    {inhab.removido_detalle.motivo}
                  </span>
                )}
              </Fila>
            )
          })}
        </ol>
      </Bloque>

      <Bloque titulo="10. Consentimiento y declaración">
        <ol className="flex flex-col gap-1.5 text-justify text-body-md">
          {CONSENTIMIENTO.map((t, i) => (
            <li key={i} className="break-inside-avoid">
              {i + 1}. {t}
            </li>
          ))}
        </ol>
      </Bloque>

      <section aria-label="Firma del postulante" className="break-inside-avoid pt-16">
        <div className="mx-auto flex w-80 max-w-full flex-col items-center gap-1 border-t border-on-surface pt-2 text-center">
          <span className="text-body-bold">Firma del postulante</span>
          <span className="text-body-md">{texto(nombreCompleto)}</span>
          <span className="text-body-md">C.I. {texto(personal?.cedula)}</span>
        </div>
      </section>
    </article>
  )
}
