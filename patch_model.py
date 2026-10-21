with open("model.go", "r") as f:
    text = f.read()

text = text.replace(
    '{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},',
    '{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}},\n\t\t{Name: "tenant_id", Type: model.Text(), NotNull: true},',
    1
)
text = text.replace(
    '{Name: "reservation_id", Type: input.Text()},',
    '{Name: "reservation_id", Type: input.Text()},\n\t\t{Name: "service_id", Type: input.Text()},',
    1
)
text = text.replace(
    '{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},',
    '{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, NotNull: true},\n\t\t{Name: "tenant_id", Type: model.Text(), NotNull: true},',
    1
)
text = text.replace(
    '{Name: "reservation_id", Type: input.Text()},',
    '{Name: "reservation_id", Type: input.Text()},\n\t\t{Name: "service_id", Type: input.Text()},',
    1
)
text = text.replace(
    'Name: "get_visit_args",\n\tFields: model.Fields{\n',
    'Name: "get_visit_args",\n\tFields: model.Fields{\n\t\t{Name: "tenant_id", Type: model.Text()},\n',
    1
)
text = text.replace(
    'Name: "list_visits_args",\n\tFields: model.Fields{\n',
    'Name: "list_visits_args",\n\tFields: model.Fields{\n\t\t{Name: "tenant_id", Type: model.Text()},\n',
    1
)
text = text.replace(
    'Name: "list_recent_patients_args",\n\tFields: model.Fields{\n',
    'Name: "list_recent_patients_args",\n\tFields: model.Fields{\n\t\t{Name: "tenant_id", Type: model.Text()},\n',
    1
)

text += """
var EncounterSupplyModel = model.Definition{
	Name: "encounter_supply",
	Fields: model.Fields{
		{Name: "id", Type: model.Text(), DB: &model.FieldDB{PK: true}, OmitEmpty: true},
		{Name: "tenant_id", Type: model.Text(), NotNull: true},
		{Name: "encounter_id", Type: model.Text(), Ref: &MedicalHistoryModel, DB: &model.FieldDB{RefColumn: "id"}, NotNull: true},
		{Name: "item_id", Type: model.Text(), NotNull: true},
		{Name: "quantity", Type: input.Decimal(), Label: "Quantity", NotNull: true},
		{Name: "billable", Type: input.Checkbox(), Label: "Billable"},
		{Name: "from_recipe", Type: model.Bool()},
		{Name: "status", Type: model.Text(), NotNull: true},
		{Name: "room_id", Type: model.Text()},
		{Name: "cost", Type: model.Int()},
		{Name: "failure", Type: model.Text()},
		{Name: "deducted_at", Type: model.NanoTime(), OmitEmpty: true},
		{Name: "updated_at", Type: model.NanoTime(), OmitEmpty: true},
	},
}
"""

with open("model.go", "w") as f:
    f.write(text)
