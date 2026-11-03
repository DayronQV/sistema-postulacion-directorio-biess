# 2026-11-03: revisión del PR #119 (rediseño institucional del frontend)

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (el PR es un cambio de diseño acordado en el grupo)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript y Tailwind) revisando como
dueño del repositorio el Pull Request de un compañero de equipo.

Contexto: el PR #119 (rama feat/rediseno-institucional) aplica la línea visual
institucional del BIESS a todas las pantallas. Según su descripción, solo cambia la
presentación: no toca la lógica, la API ni los contratos. El grupo ya no trabaja con
issues para estos cambios, solo con PR.

Hazme lo siguiente:
1. Revisa el diff y confirma que de verdad solo cambia la presentación. Cualquier cambio
   de comportamiento (botones nuevos, textos de botones, flujos) debe seguir cumpliendo
   el requerimiento funcional del IESS, que pide textos y botones concretos.
2. Revisa que los componentes nuevos sean reutilizables, accesibles (nombres accesibles,
   imágenes con alt, uso con teclado) y que tengan sus pruebas.
3. Corre lint, verificación de tipos, todas las pruebas de Vitest y el build.
4. Levanta la app de esa rama contra un backend y una base de datos temporales, con un
   postulante y el administrador de prueba, y revisa cada pantalla a 375, 768, 1024 y
   1280 px: textos cortados o tapados, desbordes horizontales y el menú lateral en móvil.
5. Corrige en la misma rama lo que encuentres, con pruebas para lo nuevo, y verifica cada
   arreglo otra vez en el navegador.
6. Haz los commits en mi nombre, en inglés y con Conventional Commits, sin mencionar IA ni
   agregar coautoría, y publica la revisión en el PR como si la hubiera escrito yo,
   empezando con un saludo al autor. No apruebes ni unas el PR.
7. Deja constancia de la sesión en docs/ia/DayronQV/ con la plantilla del equipo, y
   actualiza mi docs/decisiones.md local.
```

## Resultado
- Archivos generados o modificados: `Domicilio.tsx` y sus dos pruebas, `EncabezadoPagina.tsx`, `EncabezadoSeccion.tsx`, `LayoutAcceso.tsx`, `LayoutPrivado.tsx`, `MarcoInstitucional.tsx` y la prueba nueva `MarcoInstitucional.test.tsx`.
- Resumen de lo que hizo: devolvió el botón "Guardar" de Domicilio (el requerimiento lo pide en el punto 1.2.13) y quitó "Limpiar", que restauraba los valores de cuando se abrió la sección y podía pisar una dirección ya guardada; corrigió tres superposiciones del diseño (botones sobre la descripción entre 768 y 1279 px, título de sección blanco sobre blanco en móvil y logo cortado a 1024 px); hizo que el menú móvil se cierre con Escape, quitó la flecha de un desplegable que no existía y cambió "Usuario" por "Postulante"; agregó pruebas del marco institucional.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npm run lint`, `npx tsc -b`, `npx vitest run` (588 pruebas, 4 nuevas) y `npm run build`; mutaciones a mano sobre `MarcoInstitucional.tsx` (5 de 5 detectadas por la prueba nueva); la app de la rama en el navegador contra un backend y una base temporales, con un postulante registrado por el flujo real y el administrador sembrado, a 375, 768, 1024 y 1280 px, antes y después de cada arreglo.
- Qué se corrigió: botón "Guardar" de Domicilio y botón "Limpiar" que podía pisar datos; superposiciones en tableta y móvil; cierre del menú móvil con Escape; flecha de desplegable falsa y nombre genérico del usuario; prueba que faltaba del componente nuevo más grande.
