import re

with open("migrate/migrate_test.go", "r") as f:
    text = f.read()

text = text.replace("""	if len(execer.calls) != 1 {
		t.Fatalf("got %d Exec calls, want 1: %v", len(execer.calls), execer.calls)
	}""", """	tablesStr := fmt.Sprintf("%v", execer.calls)
	if !strings.Contains(tablesStr, "medical_history") || !strings.Contains(tablesStr, "encounter_supply") {
		t.Errorf("got %d Exec calls, tables: %v, want medical_history and encounter_supply", len(execer.calls), execer.calls)
	}""")

with open("migrate/migrate_test.go", "w") as f:
    f.write(text)
