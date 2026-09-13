---
PLAN: "fix: MedicalHistory declara widgets y su vista deja de fallar al construirse"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 13042852609281692566
PR: https://github.com/veltylabs/clinical_encounter/pull/4
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> Repo rules (import whitelist, no stdlib, `gotest`): see `AGENTS.md` at the
> repo root. The whitelist **already allows `form/input`**, with this exact
> case as its stated reason: *"Only when a `model.Definition` field needs a
> widget for a form"*.

# Plan — darle widgets a `MedicalHistoryModel`

## 1. El problema, concreto

El módulo no llega a aparecer en el chasis de la app que lo compone
(`veltylabs/mjosefa-cms`). No es un problema de esa app: es este repo.

Cadena exacta, verificada:

1. La app construye la vista con `crudview.New(...)` pasándole
   `clinicalencounter.NewView(caller)` (este repo, `view.go:68`).
2. `crudview.New` genera el formulario con
   `form.New(cfg.ParentID, cfg.Presenter.Record(), cfg.IDs)` — y el
   `Record()` es `&MedicalHistory{}`.
3. `form.New` recorre `Schema()` y **salta** todo campo cuyo `Type` no sea un
   `input.Input` (`form.go:297`: `continue // skip fields with no UI
   binding`). Si al final no quedó ninguno, devuelve error:
   `"form.New: %s has no renderable field — every Field.Type is a plain
   model.Kind, not a form input.Input."`
4. `MedicalHistoryModel` (`model.go:9`) declara **los 17 campos** con
   `model.Text()` / `model.Int()`. Ninguno es un widget. Entonces `form.New`
   falla, `NewView` propaga el error, y la app registra el fallo y **omite el
   módulo**: nunca se dibuja.

Ironía que confirma el diagnóstico: `CreateVisitArgsModel`, justo debajo en
el mismo archivo, **sí** usa `input.Text()`, `input.Textarea()` e
`input.Number()`. Los widgets están en el modelo de transporte y faltan en el
modelo persistido: están al revés de lo que `crudview` necesita.

Comparar con los módulos hermanos que sí se dibujan
(`veltylabs/device_manager`, `veltylabs/staff_manager`): los dos declaran
`input.*` en los campos editables de su modelo persistido.

## 2. Restricción que manda sobre todo lo demás

**Ningún cambio puede alterar el tipo de columna en la base.** El tipo de
almacenamiento sale de `Kind.Storage()`, que resuelve por el nombre HTML del
widget: `"number"` → `FieldInt`, `"checkbox"` → `FieldBool`, **cualquier otro
→ `FieldText`**.

Consecuencias, ambas obligatorias:

- Los campos que hoy son `model.Text()` solo pueden recibir widgets de texto
  (`input.Text`, `input.Textarea`, `input.Select`, `input.Rut`). Todos
  resuelven `FieldText`: la columna no se mueve.
- Los campos que hoy son `model.Int()` **se dejan exactamente como están**.
  El único widget que preservaría `FieldInt` es `input.Number()`, y una caja
  de números no es forma de elegir una hora de atención. Quedan fuera del
  formulario (que es lo correcto: son marcas de tiempo y auditoría que pone
  el servidor, no cosas que alguien teclea). Ver §5.

## 3. El cambio — tabla cerrada, sin decisiones pendientes

En `model.go`, dentro de `MedicalHistoryModel` **y solo ahí**, cambiar el
`Type` de estos campos. Todo lo demás del campo (`Name`, `NotNull`, `DB`,
orden) se deja intacto:

| Campo | Hoy | Queda | Por qué |
|---|---|---|---|
| `patient_id` | `model.Text()` | `input.Text()` | Referencia al paciente. Texto libre por ahora: el selector de pacientes está deferido (ver el comentario de `newView` en el repo de la app) |
| `doctor_id` | `model.Text()` | `input.Text()` | Igual que el anterior |
| `reservation_id` | `model.Text()` | `input.Text()` | Referencia opcional |
| `status` | `model.Text()` | `input.Select(...)` | Conjunto cerrado de 6 valores — ver §4 |
| `reason` | `model.Text()` | `input.Textarea()` | Motivo de consulta: texto largo, con saltos de línea |
| `diagnostic` | `model.Text()` | `input.Textarea()` | Ídem |
| `prescription` | `model.Text()` | `input.Textarea()` | Ídem |
| `cie10_code` | `model.Text()` | `input.Text()` | Código CIE-10 corto (ej. `J06.9`); el `.` está en el charset de `Text()` |
| `patient_name_snapshot` | `model.Text()` | `input.Text()` | Se envía al crear la visita (`CreateVisitArgsModel` lo exige `NotNull`) |
| `patient_rut_snapshot` | `model.Text()` | `input.Rut()` | Es un RUT chileno y el ecosistema ya tiene el widget con su dígito verificador |
| `doctor_name_snapshot` | `model.Text()` | `input.Text()` | Igual que el del paciente |
| `doctor_specialty_snapshot` | `model.Text()` | `input.Text()` | Ídem, opcional |

**Sin tocar:** `id` (es PK; `form.New` la salta siempre), `attention_at`,
`started_at`, `finished_at`, `updated_at` (ver §2).

Agregar el import `"webtyp.com/input"` a `model.go` si no está.

## 4. Las opciones de `status`

`input.Select(opts ...fmt.KeyValue)` recibe el conjunto cerrado inline, igual
que `input.Radio` en `veltylabs/device_manager/model.go`. Las claves **deben
ser las constantes que ya existen** en `const.go` — nunca los literales:

```go
{Name: "status", Type: input.Select(
    fmt.KeyValue{Key: StatusCreated, Value: "Agendada"},
    fmt.KeyValue{Key: StatusArrived, Value: "En recepción"},
    fmt.KeyValue{Key: StatusTriaged, Value: "En triage"},
    fmt.KeyValue{Key: StatusInProgress, Value: "En atención"},
    fmt.KeyValue{Key: StatusCompleted, Value: "Completada"},
    fmt.KeyValue{Key: StatusCancelled, Value: "Cancelada"},
), NotNull: true},
```

Las claves son los códigos alineados a FHIR que ya se guardan y **no
cambian**; solo las etiquetas son nuevas. Van en español porque el consumidor
de este módulo es un CMS clínico chileno y la etiqueta es lo que lee quien
atiende. (Etiquetas multi-idioma serían un cambio distinto, con
`webtyp/fmt/lang`; **no entra en este plan**.)

`model.go` ya importa `webtyp.com/fmt`; si no, agregarlo.

## 5. Lo que este plan NO hace

- **No toca `model_orm.go`.** Es generado, pero solo referencia el
  `Definition` (`Schema() []model.Field { return MedicalHistoryModel.Fields }`)
  y no codifica el `Type` de cada campo; los tipos Go de la struct dependen
  de `Storage()`, que §2 mantiene idéntico. **No ejecutar `ormc`**: el último
  commit del repo fue justamente limpiar ruido de generación
  (`chore: drop the Schema()/Pointers() stubs from list types`) y volver a
  correrlo puede reintroducirlo.
- **No agrega un selector de pacientes.** Está deferido a propósito y
  documentado en la app que compone este módulo.
- **No cambia `CreateVisitArgsModel`**, ni las ops, ni `visit.go`.
- **No agrega una migración**, porque §2 garantiza que ninguna columna cambia
  de tipo.
- **No cambia widgets por campos `model.Int()`** (§2).

## 6. Criterios de aceptación

1. `go build ./...`, `go vet ./...` y `gotest ./...` en verde.
2. `git diff --name-only` devuelve **exactamente** `model.go` (más
   `docs/PLAN.md`). Si aparece `model_orm.go`, se corrió `ormc`: revertirlo.
3. `grep -n "model.Text()" model.go` no debe devolver ninguna línea de
   `MedicalHistoryModel` entre `patient_id` y `doctor_specialty_snapshot`.
4. `grep -n '"created"\|"arrived"\|"triaged"\|"in_progress"\|"completed"\|"cancelled"' model.go`
   → vacío: las claves salen de las constantes de `const.go`, nunca de
   literales.
5. Test nuevo en `tests/model_test.go` que fija lo que este plan arregla —
   es la regresión que importa, porque el síntoma real (un módulo que no
   aparece) ocurre en otro repo:
   ```go
   // MedicalHistory debe poder generar un formulario: crudview lo construye
   // con form.New sobre este Record, y form.New falla si NINGÚN campo
   // declara un widget (input.Input).
   func TestMedicalHistoryModel_HasRenderableWidgets(t *testing.T) {
       var widgets int
       for _, f := range MedicalHistoryModel.Fields {
           if _, ok := f.Type.(input.Input); ok {
               widgets++
           }
       }
       if widgets == 0 {
           t.Fatal("MedicalHistoryModel no declara ningún widget: crudview no podrá construir su formulario")
       }
   }
   ```
   El test importa `webtyp.com/input`, permitido por el whitelist.
6. Los campos `model.Int()` siguen siendo `model.Int()`:
   `grep -c "model.Int()" model.go` no disminuye respecto de `main`.

| Etapa | Archivos | Acción |
|---|---|---|
| 1 | `model.go` | Widgets de la tabla §3 + opciones de `status` (§4) |
| 2 | `tests/model_test.go` | Test de §6.5 |
