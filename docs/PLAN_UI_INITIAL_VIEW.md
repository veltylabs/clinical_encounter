# PLAN: clinical_encounter UI — vista inicial

## Objetivo

Montar la UI del módulo `clinical_encounter` usando `tinywasm/dom` y `tinywasm/layout/rightpanel`.
El resultado debe renderizar en el navegador vía WASM y quedar listo para conectar a la DB real.

---

## Layout general (RightPanel)

```
┌──────────────────────────────────────────────────────────────────┐
│ HEAD: "Ficha Paciente"    HEAD: [Juan Pérez González — 12.345.678-9] │
├───────────────────────────────┬──────────────────────────────────┤
│ ARTICLE: Formulario visita    │ ASIDE CONTROLS: [🔍 buscar]      │
│                               ├──────────────────────────────────┤
│  Razón       [textarea]       │ ASIDE: Historial                 │
│  Diagnóstico [textarea]       │  ┌─ 22 abr 2026 · en progreso   │
│  Prescripción[textarea]       │  ├─ 01 abr 2026 · llegado        │
│  Fecha       [date]           │  ├─ 14 feb 2026 · completado     │
│  Doctor:  Dr. Nombre [display]│  └─ 10 mar 2026 · completado     │
│                               │                                  │
│  [Guardar]  [Cancelar]        │  (click → carga en formulario)   │
└───────────────────────────────┴──────────────────────────────────┘
```

### Slots del `rightpanel.RightPanel`

| Slot           | Contenido                                                    |
|----------------|--------------------------------------------------------------|
| `Title`        | `"Ficha Paciente"`                                           |
| `Head`         | `PatientHeadView` — nombre y RUT del paciente activo         |
| `HeadControls` | vacío en esta iteración                                      |
| `Article`      | `VisitFormView` — formulario de creación/edición             |
| `AsideControls`| `HistorySearchView` — input search simple                    |
| `Aside`        | `HistoryListView` — lista de 4 cards mock                    |

---

## Archivos a crear / modificar

### 1. `model.go`
- `Cie10Code` comentado en `MedicalHistory` ✓ (ya hecho)
- No se agrega a `CreateVisitArgs` en esta iteración

### 2. `view.go` — componentes UI (`package clinical_encounter`)

```go
// PatientHeadView — Head slot: nombre + RUT
type PatientHeadView struct {
    *dom.Element
    Name string
    RUT  string
}

// VisitFormView — Article slot: formulario crear/editar visita
type VisitFormView struct {
    *dom.Element
    Record     *MedicalHistory // nil = formulario vacío (nueva visita)
    DoctorName string
    onSave     func(args CreateVisitArgs)
    onCancel   func()
}

// HistoryListView — Aside slot: lista de registros
type HistoryListView struct {
    *dom.Element
    Items    []*MedicalHistory
    onSelect func(id string)
}

// HistorySearchView — AsideControls slot: búsqueda simple
type HistorySearchView struct {
    *dom.Element
}
```

**Formulario — campos visibles (Article):**

| Campo           | Elemento       | Editable |
|-----------------|----------------|----------|
| `Reason`        | `<textarea>`   | sí       |
| `Diagnostic`    | `<textarea>`   | sí       |
| `Prescription`  | `<textarea>`   | sí       |
| `AttentionAt`   | `<input date>` | sí       |
| Doctor name     | `<span>`       | no (display) |
| Botones         | Guardar / Cancelar | —    |

Campos no renderizados (se pasan por contexto desde `client.go`):
- `PatientID`, `DoctorID`, `ReservationID`, `*Snapshot`

**Historial card — campos visibles (Aside):**
- Fecha de atención formateada (día mes año)
- Estado con clase CSS de color
- Razón truncada a ~40 chars
- Click → `onSelect(id)` → recarga formulario

**CSS de status (nuevas clases en `view.go`):**

| Status       | Clase CSS            | Color sugerido |
|--------------|----------------------|----------------|
| `created`    | `.ce-status-created`     | gris           |
| `arrived`    | `.ce-status-arrived`     | azul claro     |
| `triaged`    | `.ce-status-triaged`     | amarillo       |
| `in_progress`| `.ce-status-in_progress` | naranja        |
| `completed`  | `.ce-status-completed`   | verde          |
| `cancelled`  | `.ce-status-cancelled`   | rojo           |

---

### 3. `web/mockdb/` — DB en memoria compatible con `tinywasm/orm`

Subdirectorio `web/mockdb/mockdb.go` con package `mockdb`.

Implementa las interfaces `orm.Executor` y `orm.Compiler` sobre un slice en memoria de `*MedicalHistory`.
El `web/client.go` lo instancia y pasa a `orm.New(exec, compiler)` — el `Module` recibe un `*orm.DB` real,
no un nil, por lo que el código de producción queda intacto.

```go
// web/mockdb/mockdb.go
package mockdb

// MemExecutor implementa orm.Executor en memoria
// MemCompiler implementa orm.Compiler (queries no-op que delegan al MemExecutor)
```

Datos mock fijos (4 registros):

| # | Razón             | Status       | AttentionAt (Unix) | Doctor            |
|---|-------------------|--------------|--------------------|-------------------|
| 1 | Control rutinario | completed    | 2026-03-10         | Dr. Rodrigo Soto  |
| 2 | Dolor de cabeza   | completed    | 2026-02-14         | Dr. Rodrigo Soto  |
| 3 | Revisión exámenes | in_progress  | 2026-04-01         | Dr. Rodrigo Soto  |
| 4 | Consulta general  | arrived      | 2026-04-22         | Dr. Rodrigo Soto  |

Paciente mock fijo:
```
Name: "Juan Pérez González"
RUT:  "12.345.678-9"
```

---

### 4. `web/client.go` — entrypoint WASM (`//go:build wasm`)

Responsabilidades:
1. Instanciar `mockdb.New()` → `*orm.DB`
2. Instanciar `Module{db, uid, nil}` (sin publisher en esta iteración)
3. Construir views con datos mock
4. Armar `rightpanel.RightPanel` con los slots
5. Llamar `dom.Render("body", panel.Render())`
6. Registrar handler: click en card del Aside → `dom.Update(VisitFormView)`

---

## Dependencias a agregar en `go.mod`

```
github.com/tinywasm/dom
github.com/tinywasm/layout
```

> Mismas versiones que usa el workspace tinywasm. El `web/client.go` forma parte del
> mismo módulo (`github.com/cdvelop/clinical_encounter`), sin `go.mod` propio.

---

## Interacción: click en historial

```
usuario click card[id] en Aside
  → HistoryListView.onSelect(id)
    → VisitFormView.Record = buscar por id en slice mock
    → dom.Update(VisitFormView)  // re-render formulario con datos del registro
```

---

## Estado de ejecución

- [x] Actualizar `go.mod` con `tinywasm/dom` y `tinywasm/layout`
- [x] Crear `web/mockdb/mockdb.go` (Executor + Compiler + datos mock)
- [x] Implementar `view.go` (4 componentes + CSS inline)
- [x] Reemplazar `web/client.go` con entrypoint real
- [x] Verificar compilación (`GOOS=js GOARCH=wasm go build -o /tmp/app.wasm ./web/`)
- [x] Module campos públicos (DB, UID, Pub)
- [ ] Verificar render en navegador (esperar logs del MCP)
- [ ] Integrar `tinywasm/form` para formulario real en next iteration

## Próximos pasos

1. **Formulario con tinywasm/form:** Reemplazar `VisitFormView` con un formulario real generado por `tinywasm/form.New()`
2. **Manejo de clicks en historial:** Implementar click handlers para cargar registros en el formulario
3. **Cie10Code field:** Agregar a `CreateVisitArgs` cuando se integre el formulario completo
4. **Conectar a DB real:** Reemplazar `mockdb` con ejecutor real (SQLite, PostgreSQL, etc.)
