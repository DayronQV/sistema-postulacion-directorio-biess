# 2026-11-02: Dockerfiles del backend y del frontend

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** #47

## Prompt (versión redactada)
```
Necesito empaquetar el backend y el frontend en imágenes Docker livianas para
desplegarlas en el VPS.

Crea:
1. Un Dockerfile multi-stage para el backend en Go: compila un binario estático y lo
   copia a una imagen mínima que corra sin root.
2. Un Dockerfile multi-stage para el frontend: instala dependencias, compila la SPA y
   la sirve con Nginx, con fallback a index.html para las rutas de React Router.
3. Un .dockerignore en cada uno, que deje fuera secretos, dependencias y compilados.
4. VITE_API_URL como argumento de compilación.

Comprueba que ambas imágenes se construyen y corren en local.

La issue #47 del repositorio tiene el detalle completo (archivos a tocar y criterios
de terminado): revísala para cumplirla exactamente.
```

## Para qué se usó
Crear los Dockerfiles, el nginx.conf y los .dockerignore, y verificar en local que las imágenes se construyen y funcionan.
