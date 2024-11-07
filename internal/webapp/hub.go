package webapp

import (
	"log/slog"
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
					close(cl.send)
					delete(h.clients, cl)
				}
			}
			h.RUnlock()
		case req := <-h.actions:
			slog.Info("handing action")
			if req.Action == "AddDisplay" {
				h.displaymanager.NewDisplay()
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
	switch event := ev.(type) {
	case domain.NewDisplayEvent:
		display, err := h.displaymanager.GetDisplay(event.Id)
		if err != nil {
			slog.Warn("display not found")
		}
		h.broadcast <- renderDisplayTemplate(*display)
	case domain.DisplayPowerStateChangedEvent:
		display, err := h.displaymanager.GetDisplay(event.Id)
		if err != nil {
			slog.Warn("display not found")
		}
		h.broadcast <- renderDisplayTemplate(*display)
	case domain.DisplayInputChangedEvent:
		display, err := h.displaymanager.GetDisplay(event.Id)
		if err != nil {
			slog.Warn("display not found")
		}
		h.broadcast <- renderDisplayTemplate(*display)
	default:
		slog.Warn("unexpected domain event")
	}
}
