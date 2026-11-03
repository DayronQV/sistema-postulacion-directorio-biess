# 2026-11-03: revisión y correcciones del PR #115 (calificar el examen y reporte PDF)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #44 y #114 (revisión del PR #115)

## Prompt
```
revisa el pr 115
Corrígelos tú
```

## Resultado
- Archivos generados o modificados:
  - `backend/internal/domain/evaluacion_resultado.go` (puerto `RespuestasEvaluacion`).
  - `backend/internal/repository/postgres/evaluacion_repo.go` y `evaluacion_respuestas_integration_test.go`.
  - `backend/internal/service/evaluacion_service.go`, `evaluacion_calificacion.go` y sus pruebas.
  - `backend/cmd/api/wire_evaluacion.go`.
- Resumen de lo que hizo:
  - **Nota coherente con las respuestas:** finalizar leía las respuestas, calificaba y recién después marcaba el examen como finalizado, sin transacción. Una respuesta que llegaba en medio quedaba guardada pero fuera de la nota, y el reporte, que recalcula los aciertos, mostraba otra cosa. Ahora finalizar bloquea el intento (`FOR UPDATE`), lee las respuestas, califica y guarda la nota en una sola transacción con el `TxManager` del proyecto, y guardar una respuesta bloquea el intento en modo compartido (`FOR SHARE`) antes de insertar: la respuesta entra antes de calificar o se rechaza con 409.
  - **Cierre automático:** `FinalizarVencidos` ya no se detiene en el primer examen que falla; sigue con los demás y devuelve los errores juntos. Los resultados del administrador se muestran igual y el error se registra.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go build`, `go vet` y `go test ./internal/...`; pruebas nuevas del servicio (respuesta confirmada antes del bloqueo, transacción al guardar, un vencido que falla no detiene a los demás, resultados aunque uno falle) y una de integración con PostgreSQL que comprueba que una respuesta espera al bloqueo y se rechaza. Las de integración las corre el CI (no había Docker en local).
- Qué se corrigió: lo listado arriba.
- Pendiente: confirmar en el CI la prueba de integración de concurrencia.
