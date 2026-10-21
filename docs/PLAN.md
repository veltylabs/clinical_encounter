---
PLAN: "refactor: typed instants of webtyp/time v0.6 in clinical_encounter — every instant field is time.Unix or time.Nano, no unit arithmetic"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 10891109495993698691
PR: https://github.com/veltylabs/clinical_encounter/pull/8
REVIEWED_COMMIT: 6af6de7ae1082724481bef3edc128505eec7267c
---

> Read AGENTS.md at the repo root before any change. This plan is complete: do not guess; if something does not fit, stop and ask in the PR.

## Goal

`webtyp.com/time` v0.6 put the unit of an instant in its **type** and removed `time.Now()`. This module no longer compiles against it. Make it compile, typing every instant field, without changing any behaviour and **without migrating stored data**: a field keeps the unit it already stores.

## Git rules (strict)

- Work on the branch codejob created. Never run `git reset`, `git checkout -- <file>`, `git clean`, `git stash` or any command that discards files. Commit often (`git add -A && git commit -m "..."`).
- Do not edit `model_orm.go` by hand: it is generated (Stage 2).
- **Never pause to ask permission or to report progress.** When the delivery check passes: commit, push and open the pull request. If something still fails after you tried for 30 minutes, open the pull request anyway and list in it exactly what fails (command and output).
- If `gotest` cannot run its browser WASM lane (`wasmbrowsertest` missing), run `GOOS=js GOARCH=wasm go vet ./...` instead and say so in the pull request; it does not block delivery.
- Never use `goimports` on a file that needs `webtyp.com/time`: it imports the standard library `time` instead. Write the import by hand: `"webtyp.com/time"`.

## Tools

```bash
go install webtyp.com/ormc/cmd/ormc@v0.1.23
go install webtyp.com/devflow/cmd/gotest@latest
```

## Dependencies (Stage 0)

```bash
GOFLAGS=-mod=mod go get webtyp.com/time@v0.6.2 webtyp.com/model@v0.2.4 webtyp.com/unixid@v0.3.1 webtyp.com/auth@v0.0.76 github.com/veltylabs/patient_directory@latest github.com/veltylabs/service_catalog@latest github.com/veltylabs/staff_manager@latest github.com/veltylabs/appointment_booking@v0.3.11
go mod tidy
```

## Fields to type in `model.go` (Stage 1)

In `MedicalHistoryModel`:
- `attention_at`: `model.UnixTime()` (date and time of medical attention, used as unix seconds)
- `started_at`: `model.NanoTime()`
- `finished_at`: `model.NanoTime()`
- `updated_at`: `model.NanoTime()`

In `CreateVisitArgsModel`:
- `attention_at`: `input.Number()` (keep as is if input widget, or convert at boundary)

## Code updates (Stage 2)

1. Run `ormc` to regenerate `model_orm.go`.
2. In `visit.go`:
   - `if args.AttentionAt == 0 { args.AttentionAt = time.NowUnix() }` (or convert if `CreateVisitArgs.AttentionAt` is `int64` / `time.Unix`)
   - `UpdatedAt: time.NowNano()`
3. In `view.go`:
   - `leadFromUnix(sec time.Unix)`:
     - `compact := time.FormatCompact(sec)` (FormatCompact takes `time.Unix` directly in v0.6; no `* 1e9`!)
     - `weekday := time.Weekday(sec)`
4. In `seed/seed.go`:
   - `thirtyDaysAgo := time.NowUnix() - seedVisitAgeDays*time.Day`
5. In tests:
   - Update any tests using `time.Now()` or integer timestamps to use typed `time.Unix` or `time.NowUnix()`/`time.NowNano()`.
   - Remember storage/mem rule: queries against numeric time columns require explicit `int64(...)` cast in ORM query conditions.

## Delivery Check

```bash
gotest ./...
```
Must report `race ✅, tests ✅, wasm ✅`.

