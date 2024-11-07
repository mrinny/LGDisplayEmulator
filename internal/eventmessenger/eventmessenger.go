package eventmessenger

import "github.com/mrinny/LGDisplayEmulator/internal/domain"

type EventMessenger struct {
	subs map[string][]domain.EventCallback
}

func New() *EventMessenger {
	return &EventMessenger{
		subs: make(map[string][]domain.EventCallback),
	}
}

func (em *EventMessenger) Subscribe(key string, cb domain.EventCallback) {
	cbs, found := em.subs[key]
	if found {
		cbs = append(cbs, cb)
	} else {
		cbs = make([]domain.EventCallback, 1)
		cbs[1] = cb
		em.subs[key] = cbs
	}
}

func (em *EventMessenger) Publish(ev domain.Event) {
	cbs, found := em.subs[ev.Key()]
	if !found {
		return
	}
	for _, cb := range cbs {
		go cb(ev)
	}
}
