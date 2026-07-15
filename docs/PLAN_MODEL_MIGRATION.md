# PLAN — clinical_encounter: migrar model.go a model.Definition

> This plan is dispatched via the CodeJob workflow. See skill: **agents-workflow**.

✅ **Desbloqueado.** `github.com/tinywasm/model@v0.0.14` (con `orm@v0.9.28`) ya lee `model.Definition`.
`go get github.com/tinywasm/model@v0.0.14 github.com/tinywasm/orm@v0.9.28` antes de regenerar.
⚠️ **Casing puro:** `id`→`Id`, `..._id`→`...Id` (ya no `ID`); actualiza referencias en consumidores
(ver §5).

Eres un agente **sin contexto previo** y **solo tienes este repositorio** (`clinical_encounter`). Plan
autocontenido: todo contrato, regla y ejemplo está inline.

**Nota:** existe un directorio hermano `clinical_encounterOld` (código legado, sin importadores en el
resto del ecosistema) — **no** es este repo y **no** debe tocarse ni migrarse; está fuera de alcance.

---

## 1. Qué cambia y por qué

El ecosistema tinywasm invirtió la generación de modelos: se escribe una definición tipada
(`model.Definition`) a mano, y `ormc` genera el struct concreto + plomería. Migración **mecánica**:
mismo comportamiento, mismas columnas/tabla, mismo JSON.

## 2. Contrato de `github.com/tinywasm/model` (inline)

`Field.Type` **no** es un literal de un enum — es la interfaz `Kind`. Se rellena llamando a un
constructor (`model.Text()`, `model.Int()`, …), nunca asignando `model.FieldText` directamente:

```go
package model

// FieldType es el mapeo determinista de almacenamiento/wire — lo devuelve Kind.Storage(),
// nunca se asigna directamente a Field.Type.
type FieldType int
const (
    FieldText FieldType = iota // string
    FieldInt                   // int64
    FieldFloat                 // float64
    FieldBool                  // bool
    FieldBlob                  // []byte
    FieldStruct                // struct anidado — Kind = model.Struct(ref)
    FieldIntSlice               // []int
    FieldStructSlice            // []T anidado — Kind = model.StructSlice(ref)
    FieldRaw                    // JSON pre-serializado
)

// Kind reemplaza el antiguo par Field.Type-enum + Field.Widget. Implementaciones
// sin estado, seguras para reuso concurrente.
type Kind interface {
    Storage() FieldType          // mapeo determinista a Go/DDL
    Name() string                // nombre semántico: "text", "int", "email", ...
    Validate(value string) error // SIEMPRE presente — fail-closed
}

// Constructores base — devuelven Kind, no un literal FieldType:
func Text() Kind  // storage FieldText
func Int() Kind   // storage FieldInt
func Float() Kind // storage FieldFloat
func Bool() Kind  // storage FieldBool
func Blob() Kind  // storage FieldBlob

type FieldDB struct { PK, Unique, AutoInc bool }

type Field struct {
    Name      string
    Type      Kind        // model.Text(), model.Int(), ... — NUNCA un literal FieldType
    NotNull   bool
    OmitEmpty bool
    DB        *FieldDB    // nil = sin persistencia
    Ref       *Definition // solo FK escalar; en composición (Struct/StructSlice) el ref va
                          // en el constructor del Kind, no aquí
    Exclude   bool
    Permitted
}

type Fields = []Field

type Definition struct {
    Name   string
    Fields Fields
}
```

Mapeo fijo: `model.Text()`→`string`, `model.Int()`→`int64`. Variable de definición debe llamarse
`<Struct>Model`. Este módulo **no** consume el helper de campos tipados `<Struct>_.Campo`. Nota: en
`orm` (actual) ese helper se genera **always-on** para todo modelo con DB — es inofensivo aunque no
se use; no hay que anotarlo (la directiva `// orm:typed_fields` ya no existe).

**Ya no existe `Field.Widget`.** Un Kind con UI (cuando un módulo lo necesite) es un `Kind` de
`github.com/tinywasm/form/input` (p. ej. `input.Text()`) — este módulo no usa widgets, así que le
bastan los Kinds base.

---

## 3. Estado actual (`model.go`, a portar)

```go
package clinical_encounter

type MedicalHistory struct {
	ID                      string
	PatientID               string `db:"not_null"`
	DoctorID                string `db:"not_null"`
	ReservationID           string // soft ref — optional
	Status                  string `db:"not_null"` // see Status* constants
	AttentionAt             int64  `db:"not_null"` // Unix timestamp
	Reason                  string `db:"not_null"`
	Diagnostic              string
	Prescription            string
	Cie10Code               string
	StartedAt               int64
	FinishedAt              int64
	PatientNameSnapshot     string `db:"not_null"`
	PatientRutSnapshot      string `db:"not_null"`
	DoctorNameSnapshot      string `db:"not_null"`
	DoctorSpecialtySnapshot string
	UpdatedAt               int64 `db:"not_null"`
}

type CreateVisitArgs struct {
	PatientID               string
	DoctorID                string
	AttentionAt             int64
	Reason                  string
	PatientNameSnapshot     string
	PatientRutSnapshot      string
	DoctorNameSnapshot      string
	ReservationID           string
	Diagnostic              string
	Prescription            string
	DoctorSpecialtySnapshot string
}
```

Nota: `MedicalHistory.ID` no tiene `db:"pk"` en el struct actual, pero el `model_orm.go` generado hoy
SÍ marca `{Name: "id", Type: fmt.FieldText, DB: &fmt.FieldDB{PK: true}}` — el generador actual infiere
PK por la convención de nombre `ID` incluso sin tag explícito. En la `Definition` nueva esto se
**escribe explícito** (§4) para no depender de inferencia.

## 4. Estado objetivo (`model.go` reescrito)

`MedicalHistoryModel` se queda con Kinds base (`model.Text()`/`model.Int()`) — no tiene ningún
widget hoy en el `model_orm.go` actual y no se envía como tal por un form (es el registro
histórico; se muestra, no se edita directo). `CreateVisitArgsModel` sí gana widgets aquí, de
forma **proactiva**: `view.go` ya documenta la intención (`"In production, use
tinywasm/form.New(parentID, CreateVisitArgs{})"`, ver `PLAN_UI_INITIAL_VIEW.md`), y sin widget
ese `form.New` fallaría vacío el día que se conecte — el mismo defecto ya corregido en
`service_catalog`. `diagnostic`/`prescription` usan `input.Textarea()` (texto clínico largo,
mismo criterio que `description` en `service_catalog`); el resto, `input.Text()`/`input.Number()`.

```go
package clinical_encounter

import (
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/model"
)

var MedicalHistoryModel = model.Definition{
	Name: "medical_history",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "patient_id", Type: model.Text(), NotNull: true},
		{Name: "doctor_id", Type: model.Text(), NotNull: true},
		{Name: "reservation_id", Type: model.Text()},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "attention_at", Type: model.Int(), NotNull: true},
		{Name: "reason", Type: model.Text(), NotNull: true},
		{Name: "diagnostic", Type: model.Text()},
		{Name: "prescription", Type: model.Text()},
		{Name: "cie10_code", Type: model.Text()},
		{Name: "started_at", Type: model.Int()},
		{Name: "finished_at", Type: model.Int()},
		{Name: "patient_name_snapshot", Type: model.Text(), NotNull: true},
		{Name: "patient_rut_snapshot", Type: model.Text(), NotNull: true},
		{Name: "doctor_name_snapshot", Type: model.Text(), NotNull: true},
		{Name: "doctor_specialty_snapshot", Type: model.Text()},
		{Name: "updated_at", Type: model.Int(), NotNull: true},
	},
}

var CreateVisitArgsModel = model.Definition{
	Name: "create_visit_args",
	Fields: model.Fields{
		{Name: "patient_id", Type: input.Text()},
		{Name: "doctor_id", Type: input.Text()},
		{Name: "attention_at", Type: input.Number()},
		{Name: "reason", Type: input.Text()},
		{Name: "patient_name_snapshot", Type: input.Text()},
		{Name: "patient_rut_snapshot", Type: input.Text()},
		{Name: "doctor_name_snapshot", Type: input.Text()},
		{Name: "reservation_id", Type: input.Text()},
		{Name: "diagnostic", Type: input.Textarea()},
		{Name: "prescription", Type: input.Textarea()},
		{Name: "doctor_specialty_snapshot", Type: input.Text()},
	},
}
```

Las constantes `Status*`/`Event*` y la FSM `visitTransitions` (definidas en `const.go`, no en
`model.go`) **no** se tocan — no forman parte del schema.

## 5. Pasos

> **Dependencias:** `go get github.com/tinywasm/model@v0.0.14 github.com/tinywasm/orm@v0.9.28 github.com/tinywasm/form@v0.2.15`
> (`model` directa nueva, antes solo se llegaba transitivamente vía `orm`; `form` es dependencia
> **nueva** en este repo — la trae `CreateVisitArgsModel` al ganar widgets, ver §4).

1. Reescribe `model.go` con el contenido de §4, **sin directivas**.
2. Regenera `model_orm.go` con `ormc` (instalado/actual). Los tipos no cambian (todo ya `string`/`int64`),
   pero ⚠️ **el casing sí:** `MedicalHistory.ID`→`MedicalHistory.Id` y cualquier `..._id`→`...Id`
   (casing puro, no `ID`).
3. Ajusta consumidores por el rename de casing: referencias `.ID`→`.Id` en `visit.go`/`module.go`/
   `viewCtrl.go` (y tests). Columnas/JSON del wire no cambian.
4. Crea `model_test.go` con `var _ model.Model = (*CreateVisitArgs)(nil)` y un test que construya
   `form.New("clinical_encounter", &CreateVisitArgs{})` y falle si `len(f.Inputs) == 0` — el mismo
   test que ya existe en `service_catalog` tras corregir su defecto equivalente. Es lo que impide
   que `CreateVisitArgsModel` vuelva a quedarse sin widgets en silencio.

## 6. Fuera de alcance

- No tocar `clinical_encounterOld` (código legado sin importadores — fuera de este repo).
- No cambiar nombres de tabla/columna ni comportamiento.
- No tocar `const.go` (constantes de estado/eventos, FSM).
- **No añadir** la directiva `// orm:typed_fields` (ya no existe).
- No le pongas widget a `MedicalHistoryModel` — no se envía por form, solo se muestra; solo
  `CreateVisitArgsModel` los necesita (§4).
- No implementes aquí la vista real (`view.go`/`viewCtrl.go` más allá del casing) — eso es
  `PLAN_UI_INITIAL_VIEW.md`, fuera de alcance de esta migración de modelo.

## 7. Criterio de aceptación

- `gotest ./...` verde con `go.mod` en `model v0.0.14` / `orm v0.9.28` / `form v0.2.15`.
- `model_orm.go` regenerado compila; mismos campos y tipos salvo el **casing puro** (`Id`, `...Id`),
  actualizado en todos los consumidores.
- No queda struct plano con tags `db:` ni directiva en `model.go`.
- `CreateVisitArgsModel` tiene widget en cada campo (`input.Text()`/`input.Number()`/
  `input.Textarea()`, ver §4); `MedicalHistoryModel` sigue sin ninguno.
- El test de `model_test.go` (`form.New` con `CreateVisitArgs{}`, inputs > 0) pasa.

## 8. Etapas

| # | Etapa | Salida | Criterio |
|---|---|---|---|
| 1 | `go get` model v0.0.14 + orm v0.9.28 + form v0.2.15 + reescribir `model.go` | 2 Definitions de §4, sin directiva (`CreateVisitArgsModel` con widgets) | compila |
| 2 | Regenerar `model_orm.go` | struct + plomería | tipos iguales, casing puro |
| 3 | Actualizar casing en consumidores | `visit.go`/`module.go`/`viewCtrl.go` (`.ID`→`.Id`) | `gotest ./...` verde |
| 4 | `model_test.go` | `var _ model.Model` + el form de `CreateVisitArgs` no sale vacío | `gotest ./...` verde |
