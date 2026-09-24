# clinical_encounter
<img src="docs/img/badges.svg">

Clinical encounter management bounded context for the Velty ecosystem.

## View and demo

This module exports its browser screen under `ui/` and demo data under `seed/`:

- `ui.ID`: `"clinical_encounter"`
- `ui.Label`: `"Historial Clínico"`
- `ui.Browser(caller, ids, tenantID, userID)`: returns the `platformd.UIModule` for the browser.
- `seed.Load(module, upstream)`: populates demo visits using the module's own methods.

Run `webtyp` at the repository root to open the demo — in-browser, in-memory, no login.
