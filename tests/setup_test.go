package tests

import (
	"testing"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
	"webtyp.com/events"
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"
)

type mockIDGen struct{ counter int }

func (g *mockIDGen) NewID() string {
	g.counter++
	return "test-id-" + fmt.Convert(g.counter).String() // tinywasm/fmt — stdlib strconv is banned, tests included
}

var _ model.IDGenerator = (*mockIDGen)(nil)

type mockPublisher struct{ Events []events.Event }

// events.Publisher.Publish is fire-and-forget: NO error return (verified against
// webtyp.com/events@v0.0.2 — a `Publish(e Event) error` signature does not compile).
func (p *mockPublisher) Publish(e events.Event) {
	p.Events = append(p.Events, e)
}

var _ events.Publisher = (*mockPublisher)(nil)

func setup(t *testing.T) *clinicalencounter.Module {
	t.Helper()
	db := orm.New(mem.New())
	m, err := clinicalencounter.New(db, clinicalencounter.Deps{IDs: &mockIDGen{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return m
}
