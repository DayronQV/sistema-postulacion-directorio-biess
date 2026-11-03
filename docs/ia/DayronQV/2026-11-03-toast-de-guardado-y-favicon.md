# 2026-11-03: toast de guardado y favicon propio

- **Autor (usuario de GitHub):** DayronQV
- **Herramienta y modelo:** Claude Code
- **Issue:** sin issue (ajustes de presentación del frontend, acordados en el grupo)

## Prompt
```
Actúa como desarrollador frontend sénior (React, TypeScript y Tailwind) afinando la
experiencia de uso del sistema del concurso del Directorio del BIESS.

Quiero dos ajustes de presentación, sin tocar la API ni los contratos:

1. Cuando algo se guarda bien, hoy sale un aviso fijo dentro de la pantalla. Cámbialo por un
   toast: un aviso flotante que aparece unos segundos y se quita solo. Debe cubrir todos los
   guardados, incluidos los que hoy no avisan nada (Inhabilidades y el cierre del
   formulario), y también el panel de administración (guardar una fase y la configuración del
   examen). Los mensajes de error y los de estado (registro exitoso, postulación enviada) se
   quedan donde están.
2. La pestaña del navegador muestra el logo de Vite. Reemplázalo por un icono propio del
   sistema, que se lea bien a 16 px.

Requisitos:
- Diseña el toast como un componente reutilizable con su propio almacén de estado (siguiendo
  el patrón de los almacenes que ya existen en el proyecto): que se quite solo, que se pueda
  cerrar antes, que no apile mensajes iguales, que tenga un máximo de avisos a la vez, que
  sea accesible (región en vivo, alerta para errores) y que respete la preferencia de
  reducir el movimiento. Móntalo una sola vez en App.
- Adapta las pruebas existentes que esperaban el aviso fijo, y agrega las del toast. Después
  rompe el código a propósito, una pieza a la vez, y comprueba que las pruebas lo detectan;
  si alguna mutación sobrevive, agrega la prueba que falte.
- Corre formato, verificación de tipos, lint, todas las pruebas de Vitest y el build, y
  comprueba el resultado en el navegador con un guardado real.
- Cuida el alcance: no reformatees archivos que no cambian.
- Trabaja en una rama nueva desde develop. Haz los commits en mi nombre, en inglés y con
  Conventional Commits, sin mencionar IA ni agregar coautoría, y abre un PR con la plantilla
  del repositorio.
- Deja constancia de la sesión en docs/ia/DayronQV/ con la plantilla del equipo.
```

## Resultado
- Archivos generados o modificados: `store/toasts.ts`, `components/ui/ContenedorToasts.tsx` y `hooks/useAvisoDeGuardado.ts` (nuevos, con sus pruebas), `App.tsx`, `index.css`, `features/postulacion/store.ts`, las secciones del formulario, `TarjetaFase.tsx`, `EvaluacionConfiguracion.tsx`, nueve archivos de prueba adaptados y `public/favicon.svg`.
- Resumen de lo que hizo: creó un sistema de toasts (almacén, contenedor flotante y un hook que avisa cuando un guardado pasa de falso a verdadero) que se quita solo a los 4 s, lo conectó en los guardados del formulario y del panel (y en el guardado común de la página, que cubre Inhabilidades y el cierre), y reemplazó el favicon de Vite por un edificio institucional con las franjas de la bandera, con tres columnas gruesas para que se lea a 16 px.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `npx tsc -b`, `npm run lint`, `npx vitest run` (637 pruebas) y `npm run build`; 14 mutaciones a mano sobre el almacén, el contenedor, el hook, el guardado común y el montaje en `App` (todas detectadas); un guardado real de una fase en el panel contra el backend local, donde se vio el toast con su mensaje y su botón y a los 5 s ya no estaba.
- Qué se corrigió: una mutación sobrevivió al principio (el caso de un componente que nace ya guardado y luego guarda de nuevo) y se agregó su prueba; el hook avisaba si el componente nacía con `guardado` en verdadero, contra lo que decía su comentario, y se corrigió junto con las dos pruebas que dependían de ese caso artificial; un orden de declaración en `TarjetaFase` y el ancho angosto del toast; y se restauraron 28 archivos que Prettier había reescrito solo en finales de línea. La animación de entrada se quedó congelada en el navegador embebido del panel (su reloj no avanza), por lo que se desactivó solo para la captura.
