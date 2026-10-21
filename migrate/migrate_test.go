package migrate_test

import (
	"fmt"
	"strings"

	"testing"

	"github.com/veltylabs/clinical_encounter/migrate"
	"webtyp.com/ddl"
	"webtyp.com/model"
)

type dummyExecer struct{ calls []string }

func (d *dummyExecer) Exec(query string, args ...any) error {
	d.calls = append(d.calls, query)
	return nil
}

type dummyCompiler struct{}

func (d *dummyCompiler) CompileDDL(stmt ddl.Stmt, m model.Model) (string, []any, error) {
	return stmt.Table, nil, nil
}

func TestMigrate_CreatesMedicalHistoryTable(t *testing.T) {
	execer := &dummyExecer{}
	if err := migrate.Migrate(execer, &dummyCompiler{}); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	tablesStr := fmt.Sprintf("%v", execer.calls)
	if !strings.Contains(tablesStr, "medical_history") || !strings.Contains(tablesStr, "encounter_supply") {
		t.Errorf("got %d Exec calls, tables: %v, want medical_history and encounter_supply", len(execer.calls), execer.calls)
	}
}
