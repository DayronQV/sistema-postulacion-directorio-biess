import { useState, type ComponentProps } from 'react'
import { CampoTexto } from '@/components/form/campos'
import Icon from '@/components/ui/Icon'

type CampoClaveProps = Omit<ComponentProps<typeof CampoTexto>, 'type' | 'accion'>

// Campo de contraseña con el botón del ojo para mostrarla u ocultarla. Cada campo tiene su
// propio botón: mostrar la contraseña no muestra la confirmación, y al revés.
export default function CampoClave(props: CampoClaveProps) {
  const [visible, setVisible] = useState(false)
  return (
    <CampoTexto
      {...props}
      type={visible ? 'text' : 'password'}
      accion={
        <button
          type="button"
          onClick={() => setVisible((v) => !v)}
          aria-label={`${visible ? 'Ocultar' : 'Mostrar'} ${props.etiqueta.toLowerCase()}`}
          aria-pressed={visible}
          className="flex h-9 w-9 items-center justify-center rounded-lg text-on-surface-variant hover:bg-surface-container focus:outline-none focus-visible:ring-2 focus-visible:ring-secondary"
        >
          <Icon nombre={visible ? 'visibility' : 'visibility_off'} className="text-[22px]" />
        </button>
      }
    />
  )
}
