# 2026-11-03: botones de guardado de las secciones y selección en listas

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (ajustes de comportamiento del formulario, acordados en el grupo)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript y Tailwind) revisando la
experiencia de uso del formulario de postulación del concurso del Directorio del BIESS.

Encontré dos problemas al probar a mano el formulario como postulante:

1. Botones redundantes en las secciones con tabla (Formación académica, Experiencia
   profesional, Capacitación, Publicaciones, Referencias personales y Familiares). Con el
   formulario de una fila abierto aparecen tres botones (Cancelar, Agregar y Guardar
   información) y "Guardar información" se ve siempre, incluso con la lista vacía, cuando no
   tiene nada que guardar. Quiero que con el formulario abierto solo haya Cancelar y Guardar,
   y que "Guardar" valide, agregue o actualice la fila y la guarde en el servidor en un solo
   paso; si el servidor la rechaza, el formulario debe seguir abierto con lo escrito.
   "Guardar información" solo debe verse cuando haya algo que guardar y tenga sentido:
   después de eliminar filas, o al cambiar la respuesta Sí/No en las secciones que la tienen.
   Con "Sí" y sin filas no debe verse, porque no hay nada que guardar hasta agregar una.
2. Al elegir una opción en una lista desplegable (por ejemplo el nivel de instrucción en
   Formación o el tipo de evento en Capacitación), la primera elección no se queda: la lista
   vuelve a "Seleccione..." y aparece el puntito amarillo de cambios sin guardar; hay que
   elegir una segunda vez. Encuentra la causa de raíz, que afecta a todas las pantallas con
   listas, y corrígela en un solo lugar.

Requisitos:
- Para el problema 2, escribe primero una prueba que lo reproduzca en la página real y
  comprueba que falla antes del arreglo.
- Mantén los textos y botones que pide el requerimiento funcional del IESS.
- Adapta las pruebas existentes y agrega las que falten. Después rompe el código a propósito,
  una pieza a la vez, y comprueba que las pruebas lo detectan; si alguna mutación sobrevive,
  agrega la prueba que falte.
- Corre formato, verificación de tipos, lint, todas las pruebas de Vitest y el build.
- Trabaja en una rama nueva desde develop. Haz los commits en mi nombre, en inglés y con
  Conventional Commits, sin mencionar IA ni agregar coautoría.
- Deja constancia de la sesión en docs/ia/DayronQV/ con la plantilla del equipo.
```

## Resultado
- Archivos generados o modificados: `Formacion.tsx`, `Experiencia.tsx`, `Capacitacion.tsx`, `Publicaciones.tsx`, `Referencias.tsx` y `Familiares.tsx` (en `frontend/src/features/postulacion/secciones/`) con sus pruebas, y `FormularioPostulacion.tsx` con su prueba.
- Resumen de lo que hizo: unificó el flujo de guardado de las seis secciones con tabla (Cancelar y Guardar en el formulario, que guarda la fila al momento; "Guardar información" solo con cambios pendientes, y con "Sí" solo si queda alguna fila); y corrigió la primera selección perdida en las listas quitando el `onInput` del contenedor de la página, que marcaba "cambios sin guardar" con el primer evento y hacía que React devolviera la lista a su valor anterior antes de que llegara `change`.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npx tsc -b`, `npm run lint`, `npx vitest run` (618 pruebas) y `npm run build`; una prueba de regresión del problema de las listas, vista fallar antes del arreglo y pasar después; mutaciones a mano en cada sección (más de 40, todas detectadas tras agregar las pruebas que faltaban); revisión del comportamiento de la sección en el navegador a cargo de la persona que pidió el cambio.
- Qué se corrigió: varias mutaciones sobrevivieron al principio (botón visible con el formulario abierto, agregar que pisaba las filas existentes, lista que no se actualizaba tras guardar, formulario que se cerraba aunque el guardado fallara) y se agregaron las pruebas que faltaban; "Guardar información" se ocultó también con "Sí" y sin filas tras la observación de que ahí no tenía nada que guardar. Dos pruebas de otros archivos (`ImprimirPostulacion` y `CargaMasiva`) fallaron una vez por los tiempos intermitentes ya conocidos (issue #89) y pasaron en la repetición.
