package tests

import (
	"strings"
	"testing"
	"webtyp.com/orm"
	"webtyp.com/storage/mem"

	clinicalencounter "github.com/veltylabs/clinical_encounter"
)

func TestNew_ConfigValidation(t *testing.T) {
	db := orm.New(mem.New())
	fakeRecipes := &fakeRecipes{}
	fakeRooms := &fakeRooms{}
	fakeStock := &fakeStock{}
	fakeIDs := &fakeIDs{}

	cases := []struct {
		name        string
		deps        clinicalencounter.Deps
		wantErr     bool
		errContains string
	}{
		{
			name:    "no supply deps -> ok, supplyOff",
			deps:    clinicalencounter.Deps{},
			wantErr: false,
		},
		{
			name: "all supply deps -> ok, supplyOn",
			deps: clinicalencounter.Deps{
				Recipes:      fakeRecipes,
				Rooms:        fakeRooms,
				Supplies:     fakeStock,
				IDs:          fakeIDs,
				SuppliesArea: "supplies_area",
			},
			wantErr: false,
		},
		{
			name: "missing Rooms -> error",
			deps: clinicalencounter.Deps{
				Recipes:  fakeRecipes,
				Supplies: fakeStock,
			},
			wantErr:     true,
			errContains: "requires Recipes, Rooms and Supplies together",
		},
		{
			name: "missing IDs when supplyOn -> error",
			deps: clinicalencounter.Deps{
				Recipes:      fakeRecipes,
				Rooms:        fakeRooms,
				Supplies:     fakeStock,
				SuppliesArea: "supplies_area",
			},
			wantErr:     true,
			errContains: "IDs generator is missing",
		},
		{
			name: "missing SuppliesArea when supplyOn -> error",
			deps: clinicalencounter.Deps{
				Recipes:  fakeRecipes,
				Rooms:    fakeRooms,
				Supplies: fakeStock,
				IDs:      fakeIDs,
			},
			wantErr:     true,
			errContains: "SuppliesArea is empty",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := clinicalencounter.New(db, tc.deps)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("expected error to contain %q, got %v", tc.errContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
