package webapp

import (
	"log/slog"
	"reflect"
	"sync"

	"github.com/mrinny/LGDisplayEmulator/internal/displaymanager"
	"github.com/mrinny/LGDisplayEmulator/internal/domain"
	"github.com/mrinny/LGDisplayEmulator/internal/eventmessenger"
)

type Hub struct {
	sync.RWMutex
	clients        map[*Client]bool
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
	actions        chan *ActionRequest
	eventmessenger *eventmessenger.EventMessenger
	displaymanager *displaymanager.DisplayManager
}

func NewHub(
	messenger *eventmessenger.EventMessenger,
	displaymessenger *displaymanager.DisplayManager,
) *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		broadcast:      make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		actions:        make(chan *ActionRequest),
		eventmessenger: messenger,
		displaymanager: displaymessenger,
	}
}

func (h *Hub) Run() {
	h.subscribeDomain()
	for {
		select {
		case cl := <-h.register:
			h.Lock()
			h.clients[cl] = true
			h.Unlock()
			for _, display := range h.displaymanager.GetDisplays() {
				cl.send <- renderDisplayTemplate(*display)
			}
		case cl := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[cl]; ok {
				close(cl.send)
				delete(h.clients, cl)
			}
			h.Unlock()
		case msg := <-h.broadcast:
			h.RLock()
			for cl := range h.clients {
				select {
				case cl.send <- msg:
				default:
					slog.Warn("removing client")
					close(cl.send)
					delete(h.clients, cl)
				}
			}
			h.RUnlock()
		case req := <-h.actions:
			slog.Info("handing action")
			switch req.Action {
			case "AddDisplay":
				h.displaymanager.NewDisplay()
			case "PowerOnDisplay":
				h.displaymanager.PowerOnDisplay(req.Id)
			case "PowerOffDisplay":
				h.displaymanager.PowerOffDisplay(req.Id)
			case "PowerRestartDisplay":
				h.displaymanager.RestartDisplay(req.Id)
			case "SetInput":
				h.displaymanager.SetInput(req.Id, req.Input)
			default:
				slog.Warn("unhandled action", "action", req.Action)
			}
		}
	}
}

func (h *Hub) subscribeDomain() {
	h.eventmessenger.Subscribe(domain.NewDisplayEventKey, h.HandleDomainEvent)
	h.eventmessenger.Subscribe(domain.DisplayPowerStateChangedEventKey, h.HandleDomainEvent)
	h.eventmessenger.Subscribe(domain.DisplayInputChangedEventKey, h.HandleDomainEvent)
}

func (h *Hub) HandleDomainEvent(ev domain.Event) {
	slog.Info("received domain event", "event_type", ev.Key())
	switch event := ev.(type) {
	case *domain.NewDisplayEvent:
		display, err := h.displaymanager.GetDisplay(event.Id)
		if err != nil {
			slog.Warn("display not found")
		}
		h.broadcast <- renderDisplayTemplate(*display)
	case *domain.DisplayPowerStateChangedEvent:
		display, err := h.displaymanager.GetDisplay(event.Id)
		if err != nil {
			slog.Warn("display not found")
		}
		h.broadcast <- renderDisplayUpdateTemplate(*display)
	case *domain.DisplayInputChangedEvent:
		display, err := h.displaymanager.GetDisplay(event.Id)
		if err != nil {
			slog.Warn("display not found")
		}
		h.broadcast <- renderDisplayUpdateTemplate(*display)
	default:
		slog.Warn("unexpected domain event", "type", reflect.TypeOf(ev))
	}
}
