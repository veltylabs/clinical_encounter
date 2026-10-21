package tests

import (
	"errors"
	"fmt"
	"webtyp.com/time"
)

type fakeRecipes struct {
	fail bool
	items []string
	qtys []float64
}

func (f *fakeRecipes) ServiceRecipe(tenantId, serviceId string) ([]string, []float64, error) {
	if f.fail {
		return nil, nil, errors.New("recipe error")
	}
	return f.items, f.qtys, nil
}

type fakeRooms struct {
	fail bool
	roomId string
	found bool
}

func (f *fakeRooms) RoomOfOccupant(tenantID, occupantID string, at time.Unix) (string, bool, error) {
	if f.fail {
		return "", false, errors.New("room error")
	}
	return f.roomId, f.found, nil
}

type fakeStock struct {
	cost int64
	err error
	calls []string
}

func (f *fakeStock) UseInRoom(tenantId, roomId, itemId string, quantity float64, docType, docId, by string) (int64, error) {
	f.calls = append(f.calls, fmt.Sprintf("%s:%s:%s:%.1f:%s:%s:%s", tenantId, roomId, itemId, quantity, docType, docId, by))
	return f.cost, f.err
}

type fakeIDs struct {
	counter int
}

func (f *fakeIDs) NewID() string {
	f.counter++
	return fmt.Sprintf("id_%d", f.counter)
}
