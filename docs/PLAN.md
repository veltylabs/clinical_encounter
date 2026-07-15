# Índice de planes — clinical_encounter

Este repo tiene dos planes pendientes, independientes entre sí. `codejob` despacha solo
`docs/PLAN.md` — cuando toque ejecutar uno, copia/mueve su contenido aquí (o referencia el
archivo correspondiente según lo soporte tu flujo de codejob) antes de despachar.

| Plan | Archivo | Estado | Bloqueante |
|---|---|---|---|
| UI inicial (RightPanel: formulario visita + historial) | [PLAN_UI_INITIAL_VIEW.md](PLAN_UI_INITIAL_VIEW.md) | En curso — 5/7 pasos hechos; quedan "verificar render en navegador" e "integrar tinywasm/form" | Ninguno — continuable ya |
| Migrar `model.go` a `model.Definition` (refactor de modelo tinywasm) | [PLAN_MODEL_MIGRATION.md](PLAN_MODEL_MIGRATION.md) | ✅ Desbloqueado — despachable | Resuelto: `tinywasm/orm@v0.9.24` (+ `fmt@v0.25.1`) lee `Definition`. Ojo **casing puro** (`ID`→`Id`, `..._id`→`...Id`) |

**Orden sugerido:** la UI inicial no depende del refactor de modelo — puede continuarse ya. La
migración de modelo **ya está desbloqueada** (`tinywasm/orm@v0.9.24`).
