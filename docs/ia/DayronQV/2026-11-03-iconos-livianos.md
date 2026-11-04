# 2026-11-03: fuente de iconos liviana para que no se vean sus nombres al cargar

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (detectado al probar en producción)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript, Vite) del sistema del concurso
del Directorio del BIESS.

En producción, al cargar la página, a veces se ven letras sueltas donde van los iconos
("ba", "ho", "ch"...) y al rato aparecen los iconos. Busca la causa y corrígela para que no
le pase a nadie, también la primera vez y en el celular con datos.

Requisitos:
- Sin dependencias nuevas ni cambios en la API.
- Si la solución depende de una lista de iconos, que una prueba avise cuando alguien use un
  icono nuevo y no lo agregue.
- Mide el antes y el después, corre tipos, lint, todas las pruebas y el build, y compruébalo
  en el navegador.
- Commit en mi nombre, en inglés y con Conventional Commits, y PR con la plantilla hacia
  develop.
```

## Resultado
- Archivos generados o modificados: `frontend/index.html` y `src/components/ui/iconos.test.ts` (nuevo).
- Resumen de lo que hizo: los iconos son la fuente Material Symbols, que dibuja palabras ("badge", "home"); se pedía completa a Google Fonts (~3,9 MB) y, con `display=block`, el navegador espera solo unos 3 s antes de mostrar la palabra como texto. Ahora `index.html` pide solo los 66 iconos que usa el sistema (`icon_names`, en orden alfabético, como exige Google), y la fuente baja a ~65 KB. La prueba nueva recorre el código con `import.meta.glob` (las cuatro formas de nombrar un icono: `<Icon nombre>`, los props `icono`/`iconoEtiqueta`, las listas `icono:` y los avisos) y falla si falta alguno en la lista o si la lista no está ordenada.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): tamaño medido con Google Fonts: 3.917 KB con todos los iconos y 65 KB con la lista, con respuesta 200; `npx tsc -b`, `npm run lint`, `npx vitest run` (665 pruebas) y `npm run build` (el `index.html` compilado conserva `icon_names`); 4 mutaciones (quitar un icono de la lista, desordenarla, no leer las listas `icono:` y usar un icono nuevo sin agregarlo), todas detectadas; en el navegador, la fuente cargada pesó 65 KB y cada icono midió lo mismo que su tamaño de letra, es decir, se dibujó y no se vio como texto.
- Qué se corrigió: la primera versión de la prueba leía los archivos con `node:fs`, pero el proyecto no tiene los tipos de Node y `tsc` fallaba; se cambió por `import.meta.glob` y `?raw` de Vite, que ya estaban disponibles. El primer recorrido no encontraba `undo` ni `verified`, que se eligen con una condición dentro de un prop, y se amplió la regla.
