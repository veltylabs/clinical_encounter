import re

with open("tests/patients_test.go", "r") as f:
    text = f.read()

# Make sure all CreateVisits in patients_test.go have TenantId: "t1"
text = text.replace('{Id: "v1"', '{TenantId: "t1", Id: "v1"')
text = text.replace('{Id: "v2"', '{TenantId: "t1", Id: "v2"')
text = text.replace('{Id: "v3"', '{TenantId: "t1", Id: "v3"')
text = text.replace('Id: "v" + string(rune(i+65)),', 'TenantId: "t1", Id: "v" + string(rune(i+65)),')

text = text.replace('m.ListRecentPatients(0)', 'm.ListRecentPatients("t1", 0)')

with open("tests/patients_test.go", "w") as f:
    f.write(text)

with open("migrate/migrate_test.go", "r") as f:
    text = f.read()

# We need to expect two CreateTable commands: one for MedicalHistory, one for EncounterSupply
text = text.replace("""if len(ops) == 0 {
		t.Errorf("got %d Exec calls, want 1: %v", len(ops), tables)
	}""", """// We are now syncing two tables
	tablesStr := fmt.Sprintf("%v", tables)
	if !strings.Contains(tablesStr, "medical_history") || !strings.Contains(tablesStr, "encounter_supply") {
		t.Errorf("got %d Exec calls, tables: %v, want medical_history and encounter_supply", len(ops), tables)
	}""")

text = text.replace('import (', 'import (\n\t"fmt"\n\t"strings"\n', 1)

with open("migrate/migrate_test.go", "w") as f:
    f.write(text)
