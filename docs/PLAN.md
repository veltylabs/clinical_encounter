---
PLAN: "feat!: PatientReader uses patient_directory's GetPatient"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 18185247049495220193
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to
> proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan — the encounter reads a patient with the name patient_directory exports

## Tools

```bash
go install webtyp.com/devflow/cmd/gotest@latest
```

## Why

`patient_directory` now exports only plain `<Verb><Entity>` methods (rule module-layout «Six verbs»);
`GetPatient` does not exist there. In `interfaces.go` replace `PatientReader` with:

```go
// PatientReader es la parte de patient_directory que este módulo usa.
// Ejemplo: name, doc, _, _, ok, err := p.GetPatient("t1", "p1")
type PatientReader interface {
	GetPatient(tenantID, id string) (name, documentNumber, phone, email string, found bool, err error)
}
```

## Stage 1 — call site and fakes

- `agenda.go:65`: `name, doc, ok, err := m.patients.GetPatient(tenantID, patID)` →
  `name, doc, _, _, ok, err := m.patients.GetPatient(tenantID, patID)`.
- `tests/agenda_test.go:44` `fakePatientReader` and `web/browser/main.go:41`: `GetPatient` with the new signature,
  same data (phone and email "").

## Stage 2 — docs (Spanish)

`README.md`/`docs/ARCHITECTURE.md`: `PatientReader{GetPatient}`.

## Delivery check

1. `gotest` passes.
2. `grep -rn "GetPatient" --include=*.go --include=*.md .` → empty.

Commit after every stage. Never edit `model_orm.go` by hand; a build error there means the hand-written code is wrong.
Never run `git reset`, `git restore`, `git checkout -- <file>` or `git stash` on your own work. Stuck 30 minutes: open
the PR anyway listing what fails. A PR that does not compile, or whose tests could not run, is a failed delivery.
Environment: `GOPRIVATE=webtyp.com/*,github.com/veltylabs/*`, `GOFLAGS=-mod=mod`, `go mod tidy` before `ormc`.
If a signature in the published tag differs from what this plan states, stop and name the symbol in the PR; never adapt it by guessing.

## Stages

| Stage | Files | Done when |
|---|---|---|
| 1 | `interfaces.go`, `agenda.go`, `tests/agenda_test.go`, `web/browser/main.go` | tests pass |
| 2 | `README.md`, `docs/ARCHITECTURE.md` | done |
## Executor notes
No unresolved issues or unmet tasks.
