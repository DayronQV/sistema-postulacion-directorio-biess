# 2026-11-01: revisión del login con JWT

- **Autor (usuario de GitHub):** gualter302
- **Herramienta y modelo:** Codex
- **Issue:** #14

## Prompt
```
Revisar el PR asignado y aplicar directamente las correcciones necesarias antes de aprobarlo; el usuario realizará el merge.
```

## Resultado
- Archivos generados o modificados: `backend/internal/service/auth_service_test.go` y este registro.
- Resumen de lo que hizo: se revisó el login, el bloqueo por intentos y la emisión del JWT. Se agregó una prueba que usa el token devuelto por `AuthService.Login` en una ruta protegida real con autenticación y rol de postulante.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `go test ./...`, `go vet ./...` y revisión del diff contra la issue #14.
- Qué se corrigió: la prueba anterior validaba el token directamente con el emisor, pero no demostraba su uso mediante el middleware de una ruta protegida.
