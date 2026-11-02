# 2026-11-01: contenedor del formulario de postulación

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #32

## Prompt
```
Haz lo que puedas hacer sin dependencia de los demás.
(Contexto: construir el contenedor del formulario con menú de secciones, avance, código único,
store y servicio de postulación, usando respuestas simuladas hasta que el backend de #20 esté listo.)
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/pages/FormularioPostulacion.tsx` (menú lateral y desplegable móvil, avance, código y estado, aviso de cambios sin guardar).
  - `frontend/src/features/postulacion/store.ts` (carga con `GET /postulacion`, guardado con `PUT /postulacion`).
  - `frontend/src/features/postulacion/secciones/` (`index.ts` con el registro, `tipos.ts`, `SeccionPendiente.tsx`, `PieSeccion.tsx`).
  - `frontend/src/services/postulacion.ts` (tipos de `docs/api.md` y llamadas).
  - Pruebas del store y de la página.
  - `frontend/src/features/postulacion/secciones/*Seccion.tsx` (envoltorios que montan Familiares, Referencias, Formación, Experiencia, Publicaciones y Capacitación, ya unidos a `develop`; Información personal, Domicilio y Cargo usan los suyos).
- Resumen de lo que hizo: el contenedor carga la postulación, muestra el avance por secciones completas y deja que cada sección se registre en `secciones/index.ts`. Guardar envía el estado completo editable, porque `PUT` reemplaza todo.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `tsc`, `eslint` y `vitest` en verde; además, el formulario se probó en el navegador contra la API real (login, `GET`/`PUT /postulacion`, Información personal y Domicilio dentro del contenedor). Al volver a leer la postulación al salir de cada sección, el avance refleja lo que guardan las secciones por su cuenta.
- Qué se corrigió: la lista de secciones sigue `docs/especificacion.md`, que no incluye "Referencias personales" aunque la issue la menciona; se dejó como sección 9 porque ya existe su componente en `develop`.
- Pendiente: revisión del código por la persona autora. Faltan las secciones Inhabilidades (#33) y Cierre (#34), que usan `SeccionPendiente` hasta entonces.
