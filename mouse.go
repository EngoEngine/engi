package engo

import "sync"

// NewMouseManager returns a new MouseManager that manages the provided Mouse.
func NewMouseManager(mouse *Mouse) *MouseManager {
	return &MouseManager{
		mouse:  mouse,
		mapper: make(map[MouseButton]KeyState),
	}
}

// MouseManager tracks which mouse buttons are pressed and released at the current point of time.
type MouseManager struct {
	mouse      *Mouse
	mapper     map[MouseButton]KeyState
	mu         sync.RWMutex
	lastAction Action
	lastButton MouseButton
}

// Get retrieves the current state of a MouseButton.
func (mm *MouseManager) Get(btn MouseButton) KeyState {
	mm.mu.RLock()
	state := mm.mapper[btn]
	mm.mu.RUnlock()
	return state
}

// Update updates the MouseManager with the current state of the mouse.
func (mm *MouseManager) Update() {
	mm.mu.Lock()
	mouse := mm.mouse
	if mm.lastAction == mouse.Action && mm.lastButton == mouse.Button {
		mm.mu.Unlock()
		return
	}
	state := mm.mapper[mouse.Button]
	if mouse.Action == Press {
		state.set(true)
	} else if mouse.Action == Release {
		state.set(false)
	}
	mm.mu.Unlock()
}
