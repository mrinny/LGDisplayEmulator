package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const LGDisplayRebootTime time.Duration = time.Second * 30

const (
	DisplayInputChangedEventKey      string = "DisplayInputChangedEvent"
	DisplayPowerStateChangedEventKey string = "DisplayPowerStateChangedEvent"
	NewDisplayEventKey               string = "newDisplayEvent"
)

type PowerState string

const (
	ON         PowerState = "on"
	OFF        PowerState = "off"
	RESTARTING PowerState = "restarting"
)

type LGDisplayInput string

const (
	HDMI1        LGDisplayInput = "HDMI1"
	HDMI2        LGDisplayInput = "HDMI2"
	HDMI3        LGDisplayInput = "HDMI3"
	DisplayPort1 LGDisplayInput = "DP1"
)

type LGDisplay struct {
	id              int
	serial          string
	input           LGDisplayInput
	power           PowerState
	powerchangetime time.Time
	events          []Event
}

func NewLGDisplay(id int) *LGDisplay {
	disp := &LGDisplay{
		id:              id,
		serial:          uuid.NewString(),
		input:           HDMI1,
		power:           OFF,
		powerchangetime: time.Now(),
		events:          make([]Event, 0),
	}
	disp.AddEvent(&NewDisplayEvent{Id: disp.id})
	return disp
}

func (l *LGDisplay) AddEvent(ev Event) {
	l.events = append(l.events, ev)
}

func (l *LGDisplay) Events() []Event {
	return l.events
}

func (l *LGDisplay) FlushEvents() {
	l.events = make([]Event, 0)
}

func (l *LGDisplay) GetId() int {
	return l.id
}

func (l *LGDisplay) GetSerial() string {
	return l.serial
}

func (l *LGDisplay) RestartFinished() bool {
	return l.power == RESTARTING && time.Since(l.powerchangetime) > LGDisplayRebootTime
}

func (l *LGDisplay) PowerOn() error {
	if l.power != OFF {
		return fmt.Errorf("cannot power on device which is not off")
	}
	l.power = ON
	l.powerchangetime = time.Now()
	l.AddEvent(&DisplayPowerStateChangedEvent{Id: l.id, NewPowerState: ON})
	return nil
}

func (l *LGDisplay) PowerOnAfterRestart() error {
	if l.power != RESTARTING {
		return fmt.Errorf("cannot power on device which is not restarting")
	}
	l.power = ON
	l.powerchangetime = time.Now()
	l.AddEvent(&DisplayPowerStateChangedEvent{Id: l.id, NewPowerState: ON})
	return nil
}

func (l *LGDisplay) PowerOff() error {
	if l.power != ON {
		return fmt.Errorf("cannot power off device which is not on")
	}
	l.power = OFF
	l.powerchangetime = time.Now()
	l.AddEvent(&DisplayPowerStateChangedEvent{Id: l.id, NewPowerState: OFF})
	return nil
}

func (l *LGDisplay) Restart() error {
	if l.power != ON {
		return fmt.Errorf("cannot restart device which is not on")
	}
	l.power = RESTARTING
	l.powerchangetime = time.Now()
	l.AddEvent(&DisplayPowerStateChangedEvent{Id: l.id, NewPowerState: RESTARTING})
	return nil
}

func (l *LGDisplay) GetPowerState() PowerState {
	return l.power
}

func (l *LGDisplay) SetInput(input LGDisplayInput) error {
	if l.input == input {
		return fmt.Errorf("input already set")
	}
	l.input = input
	l.AddEvent(&DisplayInputChangedEvent{Id: l.id, NewInput: input})
	return nil
}

func (l *LGDisplay) GetInput() LGDisplayInput {
	return l.input
}

type NewDisplayEvent struct {
	Id int
}

func (NewDisplayEvent) Key() string { return NewDisplayEventKey }

type DisplayInputChangedEvent struct {
	Id       int
	NewInput LGDisplayInput
}

func (DisplayInputChangedEvent) Key() string { return DisplayInputChangedEventKey }

type DisplayPowerStateChangedEvent struct {
	Id            int
	NewPowerState PowerState
}

func (DisplayPowerStateChangedEvent) Key() string { return DisplayPowerStateChangedEventKey }
