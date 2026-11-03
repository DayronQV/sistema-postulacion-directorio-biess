# 2026-11-01: vista imprimible del formulario final

- **Autor (usuario de GitHub):** LuisAnchundia
- **Herramienta y modelo:** Claude Code (Claude Sonnet 5.5)
- **Issue:** #35

## Prompt
```
Haz lo que puedas hacer sin dependencia de los demás.
(Contexto: página /postulacion/imprimir con el formulario en formato de documento, apartado de firma
y estilos de impresión A4; usa los datos del store de la postulación.)
```

## Resultado
- Archivos generados o modificados:
  - `frontend/src/pages/ImprimirPostulacion.tsx` y su prueba.
  - `frontend/src/features/postulacion/imprimible/` (`DocumentoPostulacion.tsx`, `formato.ts`, `textos.ts` y pruebas).
  - `frontend/src/routes/index.tsx` (ruta privada sin menús) y `frontend/src/index.css` (`@page` A4).
- Resumen de lo que hizo: el documento reúne encabezado institucional, código único, fecha, las secciones con sus datos, el tiempo de experiencia que calcula el servidor, el texto de consentimiento y, al final, el apartado de firma con nombres y cédula. Los controles llevan `print:hidden` y los bloques evitan cortarse entre páginas.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npm test` y `npm run build`.
- Qué se corrigió: una prueba buscaba el nombre en todo el documento y aparecía también en la firma; se acotó a su sección.
- Pendiente: no se revisó la impresión real en el navegador (vista previa A4); la forma de las publicaciones está supuesta porque `docs/api.md` no la define; falta el enlace desde el formulario hacia esta vista (se agrega con la sección de cierre, #34). Revisión del código por la persona autora.
