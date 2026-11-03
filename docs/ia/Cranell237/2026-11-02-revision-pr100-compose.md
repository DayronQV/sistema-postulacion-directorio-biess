# 2026-11-02: revisión del PR #100 (docker-compose de producción y respaldo)

- **Autor (usuario de GitHub):** Cranell237
- **Herramienta y modelo:** Claude Code (Claude Opus 5.5)
- **Issue:** #48 (revisión del PR #100)

## Prompt
```
revisa ahora el pr 100
```

## Resultado
- Archivos generados o modificados: `scripts/backup.sh`.
- Resumen de lo que hizo:
  - Revisó `docker-compose.prod.yml`, `scripts/backup.sh` y `.env.example` contra la issue #48: solo el proxy publica puertos, PostgreSQL y Redis quedan en la red interna con volúmenes, `restart` y healthchecks, y las variables que arma el compose coinciden con las que lee `backend/internal/config/config.go`.
  - Agregó `umask 077` al inicio de `backup.sh`: el respaldo contiene datos personales y antes el archivo parcial y la carpeta se creaban con los permisos por defecto del servidor; el `chmod 600` llegaba recién al final.

## Revisión humana
- Qué se verificó y cómo (pruebas, ejecución): `docker compose config` con un `.env` de prueba (válido; solo el proxy publica 80 y 443) y `bash -n scripts/backup.sh`. No se levantó el entorno: Docker no estaba encendido.
- Qué se corrigió: lo listado arriba.
- Pendiente:
