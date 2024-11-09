package displaymanager

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mrinny/LGDisplayEmulator/internal/domain"
	"github.com/mrinny/LGDisplayEmulator/internal/eventmessenger"
)

type DisplayManager struct {
	sync.RWMutex
	displays       map[int]*domain.LGDisplay
	eventmessenger *eventmessenger.EventMessenger
}

func New(em *eventmessenger.EventMessenger) *DisplayManager {
	dm := &DisplayManager{
		displays:       make(map[int]*domain.LGDisplay),
		eventmessenger: em,
	}
	go dm.run()
	return dm
}

func (dm *DisplayManager) run() {
	for {
		for id, disp := range dm.displays {
			if disp.RestartFinished() {
				err := disp.PowerOnAfterRestart()
				if err != nil {
					slog.Error("failed to poweron after restart", "error", err, "id", id)
				}
				for _, ev := range disp.Events() {
					dm.eventmessenger.Publish(ev)
				}
				disp.FlushEvents()
			}
		}
		time.Sleep(time.Second)
	}
}

func (dm *DisplayManager) GetDisplay(id int) (*domain.LGDisplay, error) {
	dp, found := dm.displays[id]
	if !found {
		return nil, fmt.Errorf("not found")
	}
	return dp, nil
}

func (dm *DisplayManager) GetDisplays() []*domain.LGDisplay {
	result := make([]*domain.LGDisplay, 0)
	for _, d := range dm.displays {
		result = append(result, d)
	}
	return result
}

func (dm *DisplayManager) NewDisplay() {
	slog.Info("(DisplayManager) NewDisplay")
	dm.Lock()
	var id int
	for i := 1; i <= len(dm.displays)+1; i++ {
		_, found := dm.displays[i]
		if !found {
			id = i
			break
		}
	}
	disp := domain.NewLGDisplay(id)
	dm.displays[id] = disp
	dm.Unlock()
	for _, ev := range disp.Events() {
		dm.eventmessenger.Publish(ev)
	}
	disp.FlushEvents()
}

func (dm *DisplayManager) PowerOnDisplay(id int) {
	slog.Info("(DisplayManager) PowerOnDisplay", "id", id)
	disp, found := dm.displays[id]
	if !found {
		slog.Warn("display not found", "id", id)
		return
	}
	err := disp.PowerOn()
	if err != nil {
		slog.Error("(DisplayManager) PowerOnDisplay", "error", err)
	}
	for _, ev := range disp.Events() {
		dm.eventmessenger.Publish(ev)
	}
	disp.FlushEvents()
}

func (dm *DisplayManager) PowerOffDisplay(id int) {
	slog.Info("(DisplayManager) PowerOffDisplay", "id", id)
	disp, found := dm.displays[id]
	if !found {
		slog.Warn("display not found", "id", id)
		return
	}
	err := disp.PowerOff()
	if err != nil {
		slog.Error("(DisplayManager) PowerOffDisplay", "error", err)
	}
	for _, ev := range disp.Events() {
		dm.eventmessenger.Publish(ev)
	}
	disp.FlushEvents()
}

func (dm *DisplayManager) RestartDisplay(id int) {
	slog.Info("(DisplayManager) RestartDisplay", "id", id)
	disp, found := dm.displays[id]
	if !found {
		slog.Warn("display not found", "id", id)
		return
	}
	err := disp.Restart()
	if err != nil {
		slog.Error("(DisplayManager) RestartDisplay", "error", err)
	}
	for _, ev := range disp.Events() {
		dm.eventmessenger.Publish(ev)
	}
	disp.FlushEvents()
}

func (dm *DisplayManager) SetInput(id int, input string) {
	slog.Info("(DisplayManager) SetInput", "id", id)
	disp, found := dm.displays[id]
	if !found {
		slog.Warn("display not found", "id", id)
		return
	}
	var err error
	switch input {
	case "HDMI1":
		err = disp.SetInput(domain.HDMI1)
	case "HDMI2":
		err = disp.SetInput(domain.HDMI2)
	case "HDMI3":
		err = disp.SetInput(domain.HDMI3)
	case "DISPLAYPORT1":
		err = disp.SetInput(domain.DisplayPort1)
	default:
		err = fmt.Errorf("unsupported input")
	}
	if err != nil {
		slog.Error("(DisplayManager) SetInput", "error", err)
	}
	for _, ev := range disp.Events() {
		dm.eventmessenger.Publish(ev)
	}
	disp.FlushEvents()
}
