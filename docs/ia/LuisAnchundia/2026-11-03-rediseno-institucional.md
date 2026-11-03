# 2026-11-03: rediseño institucional del frontend

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** ninguna (cambio de diseño acordado en el grupo; va en un PR aparte)

## Prompt
```
Replicar el diseño de las capturas de referencia: primero el login (panel azul con el logo y la foto
del edificio, tarjeta con el círculo), luego el formulario de postulación (barra lateral, barra
superior con el usuario, encabezado con la foto y las curvas de la bandera, avance en tarjeta,
menú de secciones en línea de tiempo y franja azul en cada sección) y después llevarlo a todas
las demás pantallas manteniendo la estructura de cada una.
```

## Resultado
- Archivos generados o modificados: componentes nuevos en `frontend/src/components/layout/` (`LayoutAcceso`, `TarjetaAcceso`, `MarcoInstitucional`, `EncabezadoPagina`), `frontend/src/features/postulacion/EncabezadoFormulario.tsx` y `secciones/EncabezadoSeccion.tsx`; `LayoutPrivado` y `LayoutAdmin` pasan a usar el marco común; `Input` y `Campo` aceptan iconos; páginas de acceso, del formulario, de la evaluación y del panel; imágenes en `frontend/public/img/`.
- Resumen de lo que hizo: las pantallas públicas usan el panel institucional con la tarjeta de acceso; las del postulante y del panel comparten la barra lateral (logo, menú, frase, edificio y ondas amarilla, azul y roja) y la barra superior; cada pantalla con sesión tiene el encabezado con la foto del edificio; las 11 secciones del formulario van en una tarjeta con la franja azul (el título propio de cada sección queda solo para lectores de pantalla). El logo y la foto se tomaron de los PNG entregados: se les quitó el fondo azul y el cuadriculado pintado, y la foto se pasó a WebP.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `tsc`, `eslint`, `vitest` (584 pruebas) y `npm run build`; revisión en el navegador contra la API local con un postulante en borrador, uno con la postulación enviada y el administrador.
- Qué se corrigió: el orden de los colores de las ondas (amarillo, azul y rojo), la altura del encabezado y que el encabezado nuevo también apareciera en la postulación enviada.
- Pendiente: revisión por un compañero; "Mis postulaciones", "Ayuda" y "Configuración" de las capturas no existen en el sistema y no se agregaron al menú.
