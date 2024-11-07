package displaymanager

import "github.com/mrinny/LGDisplayEmulator/internal/domain"

type DisplayManager struct {
	displays map[int]domain.LGDisplay
}

func New() *DisplayManager {
	return &DisplayManager{
		displays: make(map[int]domain.LGDisplay),
	}
}
