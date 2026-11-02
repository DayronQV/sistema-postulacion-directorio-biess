# 2026-11-02: pantalla de cierre y envío de la postulación

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #34 (parte de frontend)

## Prompt
```
Sí, empieza #34
(Contexto: la issue #34 pide, en el cierre del formulario, mostrar el consentimiento de 5 numerales con
casilla obligatoria (Guardar Postulación), enviar con confirmación previa, loader y pantalla de éxito con
el código, y listar con enlace las secciones incompletas.)
```

## Resultado
- Archivos generados o modificados: `frontend/src/features/postulacion/secciones/Cierre.tsx`, `frontend/src/features/postulacion/cierre.ts`, `frontend/src/features/postulacion/store.ts` (`enviarFormulario`), `frontend/src/services/postulacion.ts` (`enviarPostulacion`), `frontend/src/features/postulacion/secciones/index.ts`, `frontend/src/features/postulacion/secciones/tipos.ts`, `frontend/src/pages/FormularioPostulacion.tsx` y sus pruebas. Se quitó `SeccionPendiente.tsx`, que ya no se usa.
- Resumen de lo que hizo: el cierre muestra el texto de consentimiento (el mismo de la vista imprimible) con una casilla; "Guardar Postulación" guarda `consentimiento`. "Enviar Postulación" solo se habilita con la aceptación guardada y sin secciones incompletas; pide confirmar ("después de enviar no podrá editar"), llama a `POST /postulacion/enviar` y, si el servidor responde 422, lista las secciones que faltan con un enlace a cada una. Ya enviada, muestra el éxito con el código y el enlace a la vista imprimible.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `tsc`, `eslint` y `vitest` en verde; flujo completo contra la API real (formulario completo por API, 422 solo por falta de consentimiento, envío, 409 al reenviar, correo en el log) y la pantalla de éxito en el navegador.
- Qué se corrigió: nada todavía.
- Pendiente: revisión del código por la persona autora; el botón de enviar y la lista de secciones incompletas solo se probaron con pruebas unitarias, no a mano en el navegador.
