package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"github.com/tinywasm/model"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/router"
	"github.com/tinywasm/router/mock"
	"github.com/tinywasm/storage/mem"
)

func TestMountOps_CreateVisit(t *testing.T) {
	m := setup(t)
	if m.ModelName() != "clinical_encounter" {
		t.Fatalf("expected ModelName %q, got %q", "clinical_encounter", m.ModelName())
	}

	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"patient_id":"pat_1","doctor_id":"doc_1","reason":"Control","attention_at":1700000000,"patient_name_snapshot":"Juan","patient_rut_snapshot":"1-9","doctor_name_snapshot":"Dr. Soto"}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpCreateVisit, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("expected success status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
	if len(ctx.ResponseBody()) == 0 {
		t.Fatal("expected a non-empty response body")
	}
}

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

func TestMountOps_CreateVisit_RBACDenial(t *testing.T) {
	m := setup(t)
	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return false },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"patient_id":"pat_1","doctor_id":"doc_1","reason":"Control","attention_at":1700000000,"patient_name_snapshot":"Juan","patient_rut_snapshot":"1-9","doctor_name_snapshot":"Dr. Soto"}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpCreateVisit, ctx)

	if ctx.Status != 403 {
		t.Fatalf("expected 403 status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
}

func TestMountOps_GetVisit(t *testing.T) {
	m := setup(t)
	// Seed a visit
	rec, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId:           "pat_1",
		DoctorId:            "doc_1",
		Reason:              "Control",
		AttentionAt:         1700000000,
		PatientNameSnapshot: "Juan",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dr. Soto",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}

	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"id":"` + rec.Id + `"}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpGetVisit, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("expected success status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
}

func TestMountOps_ListVisits(t *testing.T) {
	m := setup(t)
	// Seed a visit
	_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId:           "pat_1",
		DoctorId:            "doc_1",
		Reason:              "Control",
		AttentionAt:         1700000000,
		PatientNameSnapshot: "Juan",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dr. Soto",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}

	reg := &mock.Router{}
	m.MountOps(reg)
	reg.Configure(mock.Config{
		Authn:     func(next router.HandlerFunc) router.HandlerFunc { return next },
		Authorize: func(userID string, resource model.Resource, action model.Action) bool { return true },
	})

	ctx := &mock.Context{
		InBody: []byte(`{"patient_id":"pat_1"}`),
	}
	ctx.SetUserID("test-user")
	reg.Invoke("OP", "/"+clinicalencounter.OpListVisitsByPatient, ctx)

	if ctx.Status != 0 && ctx.Status != 200 {
		t.Fatalf("expected success status, got %d, body=%s", ctx.Status, ctx.ResponseBody())
	}
}

func TestCreateVisit_PublishesEvent(t *testing.T) {
	db := orm.New(mem.New())
	pub := &mockPublisher{}
	m, err := clinicalencounter.New(db, clinicalencounter.Deps{IDs: &mockIDGen{}, Publisher: pub})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = m.CreateVisit(clinicalencounter.CreateVisitArgs{
		PatientId:           "pat_1",
		DoctorId:            "doc_1",
		Reason:              "Control rutinario",
		AttentionAt:         1700000000,
		PatientNameSnapshot: "Juan Pérez",
		PatientRutSnapshot:  "1-9",
		DoctorNameSnapshot:  "Dr. Soto",
	})
	if err != nil {
		t.Fatalf("CreateVisit: %v", err)
	}

	if len(pub.Events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(pub.Events))
	}
	event := pub.Events[0]
	if event.Topic != clinicalencounter.TopicVisitCreated {
		t.Fatalf("expected topic %q, got %q", clinicalencounter.TopicVisitCreated, event.Topic)
	}
	payload, ok := event.Payload.(*clinicalencounter.MedicalHistory)
	if !ok {
		t.Fatalf("expected payload type *MedicalHistory, got %T", event.Payload)
	}
	if payload.Reason != "Control rutinario" {
		t.Fatalf("expected payload reason to be 'Control rutinario', got %q", payload.Reason)
	}
}
