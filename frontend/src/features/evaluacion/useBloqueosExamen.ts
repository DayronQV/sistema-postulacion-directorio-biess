import { useEffect } from 'react'

// Bloqueos contra fraude mientras dura el examen (EF-14): el botón "Atrás" del navegador,
// copiar, cortar, pegar y el menú contextual, y el aviso al cerrar o recargar la pestaña.
// Son una barrera para el uso normal, no una garantía: la autoridad es el servidor.
export function useBloqueosExamen(activo: boolean) {
  useEffect(() => {
    if (!activo) return

    // "Atrás": se agrega una entrada al historial y se vuelve a empujar cada vez que se retrocede.
    window.history.pushState(null, '', window.location.href)
    const alRetroceder = () => window.history.pushState(null, '', window.location.href)
    window.addEventListener('popstate', alRetroceder)

    const impedir = (e: Event) => e.preventDefault()
    const eventos = ['copy', 'cut', 'paste', 'contextmenu'] as const
    eventos.forEach((nombre) => document.addEventListener(nombre, impedir))

    const alSalir = (e: BeforeUnloadEvent) => e.preventDefault()
    window.addEventListener('beforeunload', alSalir)

    return () => {
      window.removeEventListener('popstate', alRetroceder)
      eventos.forEach((nombre) => document.removeEventListener(nombre, impedir))
      window.removeEventListener('beforeunload', alSalir)
    }
  }, [activo])
}
