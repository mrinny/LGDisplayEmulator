package displaymanager

import (
	"fmt"

	"github.com/mrinny/LGDisplayEmulator/internal/domain"
	"github.com/mrinny/LGDisplayEmulator/internal/eventmessenger"
)

type DisplayManager struct {
	displays       map[int]domain.LGDisplay
	eventmessenger *eventmessenger.EventMessenger
}

func New(em *eventmessenger.EventMessenger) *DisplayManager {
	return &DisplayManager{
		displays:       make(map[int]domain.LGDisplay),
		eventmessenger: em,
	}
}

func (dm *DisplayManager) GetDisplay(id int) (*domain.LGDisplay, error) {
	dp, found := dm.displays[id]
	if !found {
		return nil, fmt.Errorf("not found")
	}
	return &dp, nil
}

func (dm *DisplayManager) GetDisplays() []*domain.LGDisplay {
	result := make([]*domain.LGDisplay, 0)
	for _, d := range dm.displays {
		result = append(result, &d)
	}
	return result
}

func (dm *DisplayManager) NewDisplay() {
	var id int
	for i := 0; i <= len(dm.displays)+1; i++ {
		_, found := dm.displays[i]
		if !found {
			id = i
			break
		}
	}

	disp := domain.NewLGDisplay(id)
	for _, ev := range disp.Events() {
		dm.eventmessenger.Publish(ev)
	}
}
