package tests

import (
	"errors"
	"testing"
	"webtyp.com/notify"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestCompleteVisit_Stage5(t *testing.T) {
	cases := []struct {
		name          string
		roomFound     bool
		stockCost     int64
		stockErr      error
		visitStatus   string
		supplyOff     bool
		expectedCost  int64
		expectAlert   bool
		expectPending bool
		expectErr     error
	}{
		{
			name:          "stock ok -> deducted, no alert",
			roomFound:     true,
			stockCost:     100,
			stockErr:      nil,
			visitStatus:   clinicalencounter.StatusCreated,
			supplyOff:     false,
			expectedCost:  100,
			expectAlert:   false,
			expectPending: false,
		},
		{
			name:          "no shift -> pending, alert",
			roomFound:     false,
			stockCost:     0,
			stockErr:      nil,
			visitStatus:   clinicalencounter.StatusCreated,
			supplyOff:     false,
			expectedCost:  0,
			expectAlert:   true,
			expectPending: true,
		},
		{
			name:          "stock empty error -> pending, alert",
			roomFound:     true,
			stockCost:     0,
			stockErr:      errors.New("not enough stock"),
			visitStatus:   clinicalencounter.StatusCreated,
			supplyOff:     false,
			expectedCost:  0,
			expectAlert:   true,
			expectPending: true,
		},
		{
			name:          "stock warning (cost > 0 but err) -> deducted",
			roomFound:     true,
			stockCost:     70,
			stockErr:      errors.New("some warning"),
			visitStatus:   clinicalencounter.StatusCreated,
			supplyOff:     false,
			expectedCost:  70,
			expectAlert:   false,
			expectPending: false,
		},
		{
			name:          "visit closed -> ErrVisitClosed",
			roomFound:     true,
			visitStatus:   clinicalencounter.StatusCompleted,
			expectErr:     clinicalencounter.ErrVisitClosed,
		},
		{
			name:          "supply off -> ok, no lines",
			visitStatus:   clinicalencounter.StatusCreated,
			supplyOff:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pub := &mockPublisher{}
			rooms := &fakeRooms{found: tc.roomFound, roomId: "r1"}
			stock := &fakeStock{cost: tc.stockCost, err: tc.stockErr}
			recipes := &fakeRecipes{items: []string{"i1"}, qtys: []float64{1}}
			ids := &fakeIDs{}

			db := orm.New(mem.New())
			deps := clinicalencounter.Deps{
				Publisher: pub,
			}
			if !tc.supplyOff {
				deps.Recipes = recipes
				deps.Rooms = rooms
				deps.Supplies = stock
				deps.IDs = ids
				deps.SuppliesArea = "supplies_area"
			}
			m, _ := clinicalencounter.New(db, deps)

			// Create visit
			_, err := m.CreateVisit(clinicalencounter.CreateVisitArgs{
				TenantId: "t1", Id: "v1", PatientId: "p1", DoctorId: "d1",
				Reason: "reason", PatientNameSnapshot: "P", PatientRutSnapshot: "1-9", DoctorNameSnapshot: "D", ServiceId: "s1",
			})
			if err != nil {
				t.Fatalf("CreateVisit: %v", err)
			}

			// Manually change status if needed to test "already completed"
			if tc.visitStatus == clinicalencounter.StatusCompleted {
				// complete it once legitimately
				_, _, err = m.CompleteVisit("t1", "v1", "by1")
				if err != nil && tc.expectErr == nil {
					t.Fatalf("CompleteVisit setup: %v", err)
				}
			}

			pub.Events = nil // reset events for the actual call

			resVisit, resLines, err := m.CompleteVisit("t1", "v1", "by1")
			if tc.expectErr != nil {
				if err == nil || err.Error() != tc.expectErr.Error() {
					t.Fatalf("expected error %v, got %v", tc.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CompleteVisit failed: %v", err)
			}

			if resVisit.Status != clinicalencounter.StatusCompleted {
				t.Errorf("expected visit status completed, got %s", resVisit.Status)
			}

			if tc.supplyOff {
				if resLines != nil {
					t.Errorf("expected nil lines when supply off")
				}
				return
			}

			if len(resLines) != 1 {
				t.Fatalf("expected 1 line, got %d", len(resLines))
			}

			line := resLines[0]
			if tc.expectPending {
				if line.Status != clinicalencounter.SupplyPendingDeduction {
					t.Errorf("expected status pending, got %s", line.Status)
				}
			} else {
				if line.Status != clinicalencounter.SupplyDeducted {
					t.Errorf("expected status deducted, got %s", line.Status)
				}
				if line.Cost != tc.expectedCost {
					t.Errorf("expected cost %d, got %d", tc.expectedCost, line.Cost)
				}
			}

			alertCount := 0
			for _, ev := range pub.Events {
				if ev.Topic == notify.TopicAlert {
					alertCount++
				}
			}

			if tc.expectAlert && alertCount != 1 {
				t.Errorf("expected 1 alert, got %d", alertCount)
			}
			if !tc.expectAlert && alertCount > 0 {
				t.Errorf("expected no alert, got %d", alertCount)
			}
		})
	}
}
