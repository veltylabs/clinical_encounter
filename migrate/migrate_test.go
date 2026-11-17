package migrate_test

import (
	"testing"
	"webtyp.com/storage"
	"webtyp.com/storage/mem"

	"github.com/veltylabs/clinical_encounter/migrate"
	"webtyp.com/ddl"
	"webtyp.com/model"
)

// dummyExecer registra cada sentencia DDL y lee a través de mem: ddl reconstruye text_search leyendo filas.
type dummyExecer struct {
	storage.Conn
	calls []string
}

func (d *dummyExecer) Exec(query string, args ...any) (storage.Result, error) {
	d.calls = append(d.calls, query)
	return storage.Affected(0), nil
}

type dummyCompiler struct{}

func (d *dummyCompiler) CompileDDL(stmt ddl.Stmt, m model.Model) (string, []any, error) {
	if stmt.Table == "text_search_state" { // tabla propia de ddl, no de este repo
		return "", nil, nil
	}
	return stmt.Table, nil, nil
}

func TestMigrate_CreatesMedicalHistoryTable(t *testing.T) {
	execer := &dummyExecer{Conn: mem.New()}
	if err := migrate.Migrate(execer, &dummyCompiler{}); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	hasMedicalHistory := false
	hasEncounterSupply := false
	for _, call := range execer.calls {
		if call == "medical_history" {
			hasMedicalHistory = true
		}
		if call == "encounter_supply" {
			hasEncounterSupply = true
		}
	}
	if !hasMedicalHistory || !hasEncounterSupply {
		t.Errorf("got %d Exec calls, tables: %v, want medical_history and encounter_supply", len(execer.calls), execer.calls)
	}
}
