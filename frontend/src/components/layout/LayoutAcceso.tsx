import type { ReactNode } from 'react'
import { Link, Outlet, useLocation } from 'react-router-dom'
import Icon from '@/components/ui/Icon'
import { RUTA_LOGIN_ADMIN } from '@/routes/rutas'

interface LayoutAccesoProps {
  children?: ReactNode
}

// Pantallas de acceso (login, registro, recuperación): panel institucional a la izquierda
// con el logo, la frase y la foto del edificio, y el contenido en una tarjeta a la derecha.
// En móvil el panel se reduce a una franja superior con el logo.
// Las imágenes están en public/img (logo sin fondo y edificio sin el cuadriculado).
export default function LayoutAcceso({ children }: LayoutAccesoProps) {
  // El personal del BIESS entra por otra pantalla (usuario y contraseña, no cédula): el pie
  // enlaza a ella desde las pantallas del postulante. En la del panel no se muestra, porque la
  // propia tarjeta ya trae el enlace de vuelta ("¿Es postulante? Ingrese aquí").
  const enPanel = useLocation().pathname === RUTA_LOGIN_ADMIN
  return (
    <div className="flex min-h-screen flex-col bg-surface">
      <div className="relative flex flex-1 flex-col lg:flex-row">
        {/* Panel institucional */}
        <aside className="relative overflow-hidden bg-gradient-to-b from-primary via-primary-container to-[#2a5d9c] text-on-primary lg:w-[34%] lg:[clip-path:ellipse(100%_140%_at_0%_50%)]">
          <div className="relative z-10 flex flex-col gap-space-md px-margin-mobile py-space-lg sm:px-margin lg:px-14 lg:pt-16">
            {/* Ancho con tope: a 1024 px el panel mide menos que 340 px más sus márgenes. */}
            <Link to="/" className="block w-56 focus:outline-none lg:w-full lg:max-w-[340px]">
              <img
                src="/img/logo-biess.png"
                alt="BIESS, Banco del Instituto Ecuatoriano de Seguridad Social"
                className="h-auto w-full"
              />
            </Link>
            <span className="hidden h-px w-40 bg-on-primary/40 lg:my-space-md lg:block" />
            <p className="hidden text-[32px] font-light italic leading-tight lg:block">
              Tu bienestar,
              <br />
              <span className="font-bold">nuestra prioridad</span>
            </p>
          </div>

          {/* Foto del edificio teñida con el azul del panel y fundida hacia arriba */}
          <div className="pointer-events-none absolute inset-x-0 bottom-0 hidden h-[52%] lg:block">
            <img
              src="/img/edificio-biess.webp"
              alt=""
              className="absolute bottom-0 left-0 h-full w-auto max-w-none object-contain object-left-bottom opacity-80 mix-blend-luminosity [mask-image:linear-gradient(to_top,black_55%,transparent)]"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-primary/50 via-primary-container/20 to-transparent" />
          </div>
        </aside>

        {/* Contenido */}
        <main className="relative flex flex-1 items-center justify-center overflow-hidden px-margin-mobile py-space-xl sm:px-margin">
          <svg
            aria-hidden="true"
            viewBox="0 0 800 600"
            preserveAspectRatio="none"
            className="pointer-events-none absolute inset-0 h-full w-full text-primary-fixed"
          >
            <path
              d="M0 520 C220 440 420 620 800 470 V600 H0 Z"
              fill="currentColor"
              opacity="0.35"
            />
            <path d="M420 0 C560 90 700 40 800 120 V0 Z" fill="currentColor" opacity="0.3" />
          </svg>
          <div className="relative z-10 w-full max-w-[760px]">{children ?? <Outlet />}</div>
        </main>
      </div>

      <footer className="flex flex-col items-center justify-center gap-3 bg-surface-container-lowest px-margin-mobile py-space-md text-legal-code text-on-surface-variant shadow-[0_-1px_8px_rgba(0,0,0,0.04)] sm:px-margin md:flex-row">
        <div className="flex items-center gap-3">
          {/* shrink-0: en el celular el texto largo ocupa todo el ancho y, sin esto, aplasta
              el escudo hasta cortarlo. */}
          <Icon
            nombre="verified_user"
            relleno
            className="shrink-0 text-[22px] text-primary-container"
          />
          <span className="h-6 w-px shrink-0 bg-outline-variant" />
          <p>
            © {new Date().getFullYear()} Banco del Instituto Ecuatoriano de Seguridad Social
            (BIESS). Sistema del Concurso de Méritos y Oposición.
          </p>
        </div>
        {!enPanel && (
          <>
            <span className="hidden h-6 w-px bg-outline-variant md:block" />
            <Link
              to={RUTA_LOGIN_ADMIN}
              className="inline-flex shrink-0 items-center gap-2 rounded-full border border-primary-container bg-primary-fixed/40 px-4 py-2 text-body-bold text-primary-container hover:bg-primary-container hover:text-on-primary focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
            >
              <Icon nombre="admin_panel_settings" relleno className="text-[22px]" />
              Acceso del personal del BIESS
            </Link>
          </>
        )}
      </footer>
    </div>
  )
}
