# 2026-11-02: docker-compose de producción

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #48

## Prompt (versión redactada)
```
Necesito definir el entorno de producción completo en un solo archivo.

Crea:
1. Un docker-compose.prod.yml con el backend, el frontend, PostgreSQL, Redis y el proxy
   (usa las imágenes de la issue #47).
2. PostgreSQL y Redis sin puertos expuestos a internet, con volúmenes persistentes,
   reinicio automático (unless-stopped) y healthchecks.
3. Las variables desde un .env del servidor, y documenta las de producción en
   .env.example. No cargues en producción la semilla de administración, porque trae
   contraseñas públicas.
4. Un script de respaldo diario de PostgreSQL con pg_dump, con rotación.

Comprueba que todo el sistema levanta con un solo comando y que el respaldo se puede
restaurar.

La issue #48 del repositorio tiene el detalle completo (archivos a tocar y criterios de
terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Escribir el compose de producción y el script de respaldo, y probarlos en local (arranque con un comando, respaldo, restauración, rotación y fallo).
