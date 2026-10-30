# clinical_encounter
<img src="docs/img/badges.svg">

Contexto de límite de gestión de encuentros clínicos para el ecosistema Velty.

## Vista y demo

Este módulo exporta su pantalla de navegador bajo `ui/` y datos de demostración bajo `seed/`:

- `ui.ID`: `"clinical_encounter"`
- `ui.Label`: `"Historial Clínico"`
- `ui.Browser(caller, ids, tenantID, userID)`: devuelve el `platformd.UIModule` para el navegador.
- `seed.Load(module, upstream)`: puebla visitas de demostración utilizando los propios métodos del módulo.

Ejecute `webtyp` en la raíz del repositorio para abrir la demostración — en el navegador, en memoria, sin inicio de sesión.
