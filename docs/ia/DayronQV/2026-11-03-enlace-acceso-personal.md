# 2026-11-03: enlace al acceso del personal y parpadeo de los iconos

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (ajustes de presentación del frontend, acordados en el grupo)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript y Tailwind) trabajando sobre el
frontend ya rediseñado del sistema del concurso del Directorio del BIESS.

Quiero dos ajustes de presentación, sin tocar la API ni los contratos:

1. El panel de administración tiene su propio login (/admin/login, con usuario y contraseña),
   distinto del login del postulante (/login, con cédula). Hoy ninguna pantalla pública
   enlaza a él, así que quien entre al sitio desplegado no sabe que existe. Agrega en las
   pantallas de acceso un enlace claramente visible, con aspecto de botón, que lleve al login
   del panel. En el propio login del panel no debe repetirse, porque ya trae el enlace de
   vuelta al del postulante.
2. Al cargar o recargar cualquier pantalla se ven por una fracción de segundo los nombres de
   los iconos escritos como texto (por ejemplo "badge" o "lock") antes de que aparezca el
   dibujo. Corrígelo de raíz, sin que el diseño se mueva mientras la fuente carga.

Requisitos:
- Mantén los textos y botones que pide el requerimiento funcional del IESS.
- Agrega pruebas para lo nuevo, comprueba que detectan fallos rompiendo el código a
  propósito, y corre lint, verificación de tipos, todas las pruebas de Vitest y el build.
- Revisa el resultado en el navegador a escritorio y a móvil (375 px): sin desbordes ni
  iconos recortados, en el login del postulante, en el del panel y dentro del panel.
- Trabaja en una rama nueva desde develop. Haz los commits en mi nombre, en inglés y con
  Conventional Commits, sin mencionar IA ni agregar coautoría.
- Deja constancia de la sesión en docs/ia/DayronQV/ con la plantilla del equipo.
```

## Resultado
- Archivos generados o modificados: `frontend/src/components/layout/LayoutAcceso.tsx`, su prueba nueva `LayoutAcceso.test.tsx`, `frontend/index.html` y `frontend/src/index.css`.
- Resumen de lo que hizo: agregó en el pie de las pantallas de acceso del postulante el botón "Acceso del personal del BIESS", que lleva a `/admin/login` y no se muestra en esa misma pantalla; y corrigió el parpadeo de los nombres de los iconos pidiendo la fuente Material Symbols con `display=block` en vez de `swap` y fijando el icono en un cuadro de 1 em con el desborde recortado, para que el nombre escrito no se vea ni ensanche el icono mientras la fuente carga.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npx tsc -b`, `npx vitest run` (591 pruebas, 3 nuevas) y `npm run build` (el HTML compilado pide la fuente de iconos con `display=block`); mutaciones a mano sobre el enlace (3 de 3 detectadas por la prueba); la app en el navegador, a escritorio y a 375 px, en el login del postulante, el registro, el login del panel y dentro del panel, comprobando que el enlace lleva a donde debe, que no hay scroll horizontal y que los iconos se ven completos.
- Qué se corrigió: se quitó el enlace de vuelta que se había puesto en el pie del login del panel, porque duplicaba el que ya tiene su tarjeta; se pasó Prettier a `index.html` y `index.css`. Tres pruebas de `Inhabilidades` fallaron una vez por los tiempos intermitentes ya conocidos (issue #89), pasaron solas y en la segunda corrida completa.
