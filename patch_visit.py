with open("visit.go", "r") as f:
    text = f.read()

text = text.replace("""if args.Id == "" || args.PatientId == "" || args.DoctorId == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, ErrMissingArgs
	}""", """if args.TenantId == "" {
		return nil, ErrTenantRequired
	}
	if args.Id == "" || args.PatientId == "" || args.DoctorId == "" || args.Reason == "" ||
		args.PatientNameSnapshot == "" || args.PatientRutSnapshot == "" || args.DoctorNameSnapshot == "" {
		return nil, ErrMissingArgs
	}""")

text = text.replace('if existing, err := m.GetVisit(args.Id); err == nil {', 'if existing, err := m.GetVisit(args.TenantId, args.Id); err == nil {')

text = text.replace("""Id:                      args.Id,
		PatientId:               args.PatientId,""", """Id:                      args.Id,
		TenantId:                args.TenantId,
		PatientId:               args.PatientId,""")

text = text.replace("""ReservationId:           args.ReservationId,
		Status:                  StatusCreated,""", """ReservationId:           args.ReservationId,
		ServiceId:               args.ServiceId,
		Status:                  StatusCreated,""")

recipeLogic = """if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicVisitCreated, Payload: record})
	}
	if m.supplyOn && record.ServiceId != "" {
		itemIds, quantities, err := m.recipes.ServiceRecipe(record.TenantId, record.ServiceId)
		if err != nil {
			return record, err
		}
		for i, itemId := range itemIds {
			line := &EncounterSupply{
				Id:          m.ids.NewID(),
				TenantId:    record.TenantId,
				EncounterId: record.Id,
				ItemId:      itemId,
				Quantity:    quantities[i],
				Billable:    false,
				FromRecipe:  true,
				Status:      SupplyPlanned,
				UpdatedAt:   time.NowNano(),
			}
			if err := m.db.Create(line); err != nil {
				return record, err
			}
		}
	}
	return record, nil"""

text = text.replace("""if m.pub != nil {
		m.pub.Publish(events.Event{Topic: TopicVisitCreated, Payload: record})
	}
	return record, nil""", recipeLogic)

text = text.replace("""func (m *Module) GetVisit(id string) (*MedicalHistory, error) {
	var rec MedicalHistory
	qb := m.db.Query(&rec).Where(MedicalHistory_.Id).Eq(id)""", """func (m *Module) GetVisit(tenantId, id string) (*MedicalHistory, error) {
	if tenantId == "" {
		return nil, ErrTenantRequired
	}
	var rec MedicalHistory
	qb := m.db.Query(&rec).Where(MedicalHistory_.Id).Eq(id).Where(MedicalHistory_.TenantId).Eq(tenantId)""")

text = text.replace("""func (m *Module) ListVisitsByPatient(patientId string) ([]*MedicalHistory, error) {
	var rec MedicalHistory
	qb := m.db.Query(&rec).
		Where(MedicalHistory_.PatientId).Eq(patientId).""", """func (m *Module) ListVisitsByPatient(tenantId, patientId string) ([]*MedicalHistory, error) {
	if tenantId == "" {
		return nil, ErrTenantRequired
	}
	var rec MedicalHistory
	qb := m.db.Query(&rec).
		Where(MedicalHistory_.PatientId).Eq(patientId).
		Where(MedicalHistory_.TenantId).Eq(tenantId).""")

with open("visit.go", "w") as f:
    f.write(text)
