package engo

import (
	"sync"
)

const (
	KeyStateUp = iota
	KeyStateDown
	KeyStateJustDown
	KeyStateJustUp
)

// NewKeyManager creates a new KeyManager.
func NewKeyManager() *KeyManager {
	return &KeyManager{
		dirtmap: make(map[Key]Key),
		mapper:  make(map[Key]KeyState),
	}
}

// KeyManager tracks which keys are pressed and released at the current point of time.
type KeyManager struct {
	dirtmap map[Key]Key
	mapper  map[Key]KeyState
	mutex   sync.RWMutex
}

// Set is used for updating whether or not a key is held down, or not held down.
func (km *KeyManager) Set(k Key, state bool) {
	km.mutex.Lock()

	ks := km.mapper[k]
	ks.set(state)
	km.mapper[k] = ks
	km.dirtmap[k] = k

	km.mutex.Unlock()
}

// Get retrieves a keys state.
func (km *KeyManager) Get(k Key) KeyState {
	km.mutex.RLock()
	ks := km.mapper[k]
	km.mutex.RUnlock()

	return ks
}

func (km *KeyManager) update() {
	km.mutex.Lock()

	// Set all keys to their current states
	//for key, state := range km.mapper {
	//	state.set(state.currentState)
	//	km.mapper[key] = state
	//}

	for _, key := range km.dirtmap {
		delete(km.dirtmap, key)

		state := km.mapper[key]
		state.set(state.currentState)
		km.mapper[key] = state
	}

	km.mutex.Unlock()
}

// KeyState is used for detecting the state of a key press.
type KeyState struct {
	lastState    bool
	currentState bool
}

func (key *KeyState) set(state bool) {
	key.lastState = key.currentState
	key.currentState = state
}

// State returns the raw state of a key.
func (key *KeyState) State() int {
	if key.lastState {
		if key.currentState {
			return KeyStateDown
		} else {
			return KeyStateJustUp
		}
	} else {
		if key.currentState {
			return KeyStateJustDown
		} else {
			return KeyStateUp
		}
	}
}

func (key KeyState) Up() bool {
	return (!key.lastState && !key.currentState)
}

func (key KeyState) Down() bool {
	return (key.lastState && key.currentState)
}

func (key KeyState) JustPressed() bool {
	return (!key.lastState && key.currentState)
}

func (key KeyState) JustReleased() bool {
	return (key.lastState && !key.currentState)
}
