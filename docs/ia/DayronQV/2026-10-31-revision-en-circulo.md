# 2026-10-31: revisión de PR en círculo

- **Autor (usuario de GitHub):** DayronQV
- **Issue:** ninguna (ajuste de las reglas del equipo)

## Prompt (versión redactada)
```
Actúa como administrador del repositorio.

Quiero que la revisión de los Pull Request sea en círculo entre los cinco integrantes:
José revisa a Dayron, Luis a José, Estalin a Luis, Gualter a Estalin y Dayron a Gualter.

Configura lo necesario:
1. Que al abrir un PR se pida la revisión automáticamente a la persona que le toca
   según el autor (un workflow de GitHub Actions).
2. Que baste 1 aprobación y que cualquiera de los cinco pueda aprobar y unir un PR,
   sin exigir un code owner. Elimina el CODEOWNERS.
3. Que quede documentado en AGENTS.md.
```

## Para qué se usó
Automatizar la asignación del revisor y ajustar la regla de protección de las ramas.
