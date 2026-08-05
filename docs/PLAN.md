---
PLAN: "test: clinical_encounter translate comments to Spanish, cover untested error branches"
TAG: v0.1.1
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 1979765201827893236
PR: https://github.com/veltylabs/clinical_encounter/pull/2
---

> This plan is dispatched via the CodeJob workflow. See skill: **agents-workflow**.

# PLAN — clinical_encounter: comentarios en español + cobertura de ramas de error

Eres un agente **sin contexto previo** y **solo tienes este repositorio** (`clinical_encounter`). El
módulo ya adoptó el patrón del arnés reutilizable en una ronda anterior; este plan es un ajuste
pequeño y autocontenido sobre ese trabajo ya mergeado.

## 1. Por qué existe este plan

Dos cosas quedaron pendientes de la ronda anterior:

1. **Comentarios todavía en inglés** en `module.go`, `ops.go`, `view.go` y `model.go` — el resto del
   batch (`business_hours`, `provider_payouts`, `agent_switch`, `work_schedule`) ya tradujo los suyos
   a español, manteniendo el código (identificadores) en inglés. Este módulo no lo hizo.
2. **Cobertura 48.7%**, medida con `go test ./tests/... -coverpkg=github.com/veltylabs/clinical_encounter
   -cover`. El objetivo del ecosistema es >=80%, pero **igual que en `work_schedule`/`agent_switch`,
   buena parte de lo que falta es plumbing generado estructuralmente inalcanzable en este módulo — no
   lo persigas más allá de lo que este plan pide**. Este plan agrega 4 pruebas de valor real (ramas de
   error del Op handler y de `New`, hoy sin ejercitar) que suben la cobertura a ~53%.

## 2. Traducir comentarios a español

Traduce **todos** los comentarios existentes en `module.go`, `ops.go`, `view.go` y `model.go` a
español, sin tocar ningún identificador, string literal de negocio, ni lógica. Aplica el mismo
criterio usado en el resto del batch. Ejemplos concretos de los comentarios a traducir en este
módulo:

**`module.go`:**
```go
// ANTES
// Deps are the module's infrastructure ports — never a concrete implementation.
type Deps struct {
	IDs       model.IDGenerator // required — the module never builds its own
	Publisher events.Publisher  // optional — nil disables publishing silently
}
...
	// ddl.Compiler is an optional capability — only SQL backends (sqlt, postgres) implement it.
	// storage/mem (this module's own tests, Stage 6) creates tables lazily and needs no DDL, so a
	// type assertion — not an unconditional call — is how the module stays backend-agnostic here.

// DESPUÉS
// Deps son los puertos de infraestructura del módulo — nunca una implementación concreta.
type Deps struct {
	IDs       model.IDGenerator // requerido — el módulo nunca lo construye por sí mismo
	Publisher events.Publisher  // opcional — nil deshabilita la publicación silenciosamente
}
...
	// ddl.Compiler es una capacidad opcional — solo los backends SQL (sqlt, postgres) la implementan.
	// storage/mem (las pruebas propias de este módulo) crea tablas de forma perezosa y no necesita DDL,
	// así que una aserción de tipo — no una llamada incondicional — es cómo el módulo se mantiene agnóstico aquí.
```

**`ops.go`** — el comentario de convención de estado dentro de `opCreateVisit`:
```go
// ANTES
		// Status convention (ecosystem-wide): 400 = invalid input, 404 = not found,
		// 500 = genuine internal error only — never collapse client errors into 500.

// DESPUÉS
		// Convención de estado (en todo el ecosistema): 400 = entrada inválida, 404 = no encontrado,
		// 500 = solo error interno genuino — nunca colapsar errores de cliente en 500.
```

**`view.go`** — ambos docstrings (`Item` y `NewView`):
```go
// ANTES
// Item implements view.Itemizer — the ONLY view-specific code this record carries. The
// Presenter indexes rows by ID from this during Reload; there is no manual byID/WithFill lookup.
func (it *MedicalHistory) Item() view.Item {

// NewView builds the medical-history Presenter — the tech-agnostic engine a renderer
// (tinywasm/layout/rightpanel, or any other) wraps. This module builds it (view+model+router
// only); the app decides which renderer draws it.

// DESPUÉS
// Item implementa view.Itemizer — el ÚNICO código específico de view que carga este registro. El
// Presenter indexa las filas por ID a partir de esto durante Reload; no hay lookup manual byID/WithFill.
func (it *MedicalHistory) Item() view.Item {

// NewView construye el Presenter del historial médico — el motor agnóstico de tecnología que envuelve
// un renderer (tinywasm/layout/rightpanel, o cualquier otro). Este módulo lo construye (solo
// view+model+router); la app decide qué renderer lo dibuja.
```

**`model.go`** — cuatro comentarios (nota: los dos que mencionan "Stage 3" citan al plan original ya
borrado — no reproduzcas esa referencia, usa la versión de abajo que la quita):
```go
// ANTES
		// status: valid values are ONLY the Status* constants in const.go — the literals live
		// in those constants and nowhere else (anti magic-string rule; see item_catalog review).
...
// NotNull mirrors exactly the required-argument set CreateVisit enforces (visit.go) — the
// Definition is the single place the contract is declared; the service's manual checks are a
// defense-in-depth duplicate of the same set, never a different one.
...
// GetVisitArgsModel and ListVisitsArgsModel are transport-only (Field.DB is nil throughout) —
// see Stage 3 for the ops that consume them.
...
// Domain event topics — <module>.<entity>.<past-tense-verb>, tenant/id data goes in the
// payload, never in the topic name (n/a here — see "no TenantId field" note in ARCHITECTURE.md).

// DESPUÉS
		// status: los valores válidos son ÚNICAMENTE las constantes Status* en const.go — los literales
		// viven solo en esas constantes y en ningún otro lugar (regla anti magic-string; ver la revisión de item_catalog).
...
// NotNull refleja exactamente el conjunto de argumentos requeridos que CreateVisit exige (visit.go) —
// el Definition es el único lugar donde se declara el contrato; las validaciones manuales del servicio
// son un duplicado de defensa en profundidad del mismo conjunto, nunca uno distinto.
...
// GetVisitArgsModel y ListVisitsArgsModel son solo de transporte (Field.DB es nil en todos los campos) —
// ver ops.go para las operaciones que los consumen.
...
// Topics de eventos de dominio — <módulo>.<entidad>.<verbo-en-pasado>, los datos de tenant/id van en
// el payload, nunca en el nombre del topic (no aplica aquí — ver la nota "no TenantId field" en ARCHITECTURE.md).
```

`visit.go` no tiene comentarios — no requiere cambios.

## 3. Analizado y descartado — no agregar pruebas para esto

Igual razonamiento que en `work_schedule`/`agent_switch`: alrededor de la mitad de los métodos
generados por `ormc` para los 4 `Definition`s de este módulo nunca se invocan por ningún camino real
(p. ej. `MedicalHistory.DecodeFields`/`Validate` — el registro se construye por struct literal en
`CreateVisit`, nunca se decodifica ni valida vía el generado; `ModelName`/`Schema`/`Pointers` en los
3 tipos de args de transporte — el codec JSON usa `EncodeFields`/`DecodeFields` directamente; las 4
variantes `XxxList` — ninguna se usa como lista de nivel superior salvo `MedicalHistoryList`, que sí
se ejercita). **No escribas pruebas para cerrar esos huecos llamando a los métodos directamente.**

## 4. Las 4 pruebas de valor real a agregar — en `tests/ops_test.go`

### 4.1 — Extender `TestMountOps_CreateVisit` con una aserción de `ModelName`

Al inicio de la función, justo después de `m := setup(t)`, agrega:

```go
	if m.ModelName() != "clinical_encounter" {
		t.Fatalf("expected ModelName %q, got %q", "clinical_encounter", m.ModelName())
	}
```

### 4.2 — Tres funciones nuevas, agregadas después de `TestMountOps_CreateVisit`

Cubren tres ramas de error que hoy solo se prueban a nivel de servicio (`CreateVisit`/`GetVisit`
directo) pero nunca a través del Op/`router.Context`:

```go
func TestMountOps_CreateVisit_DecodeError(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{InBody: []byte(`not valid json`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpCreateVisit, ctx)

	if ctx.Status != 400 {
		t.Fatalf("expected 400 for a malformed body, got %d", ctx.Status)
	}
}

func TestMountOps_CreateVisit_MissingArgs(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{InBody: []byte(`{"patient_id":"pat_1"}`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpCreateVisit, ctx)

	if ctx.Status != 400 {
		t.Fatalf("expected 400 for missing required args, got %d", ctx.Status)
	}
}

func TestMountOps_GetVisit_NotFound(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{InBody: []byte(`{"id":"does-not-exist"}`)}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpGetVisit, ctx)

	if ctx.Status != 404 {
		t.Fatalf("expected 404 for a non-existent id, got %d", ctx.Status)
	}
}

func TestNew_RequiresIDs(t *testing.T) {
	db := orm.New(mem.New())
	if _, err := clinicalencounter.New(db, clinicalencounter.Deps{}); err == nil {
		t.Fatal("expected an error when Deps.IDs is nil")
	}
}
```

`TestNew_RequiresIDs` necesita `"github.com/tinywasm/orm"` y `"github.com/tinywasm/storage/mem"` —
ya están en el bloque de imports de `tests/ops_test.go`, no hace falta agregar nada.

## 5. Fuera de alcance

- No tocar `visit.go`, `model_orm.go` — este plan es traducción + pruebas.
- No agregar pruebas para lo descartado en §3.
- No perseguir 80% — ~53% es el resultado esperado de aplicar exactamente §4.

## 6. Criterio de aceptación

- `go build ./...` y `GOOS=js GOARCH=wasm go build ./...` limpios.
- `gotest ./...` verde.
- `grep -n "^\s*//" module.go ops.go view.go model.go` — ningún comentario en inglés remanente.
- `go test ./tests/... -coverpkg=github.com/veltylabs/clinical_encounter -cover` reporta ~53% (no
  80% — ver §1/§3).
- `git status` limpio tras el commit.
