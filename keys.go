package engo

import (
	"sync"
)

const (
	// KeyStateUp is a state for when the key is not currently being pressed
	KeyStateUp = iota
	// KeyStateDown is a state for when the key is currently being pressed
	KeyStateDown
	// KeyStateJustDown is a state for when a key was just pressed
	KeyStateJustDown
	// KeyStateJustUp is a state for when a key was just released
	KeyStateJustUp
)

// NewKeyManager creates a new KeyManager.
func NewKeyManager() *KeyManager {
	return &KeyManager{
		mapper: make(map[Key]KeyState),
	}
}

// KeyManager tracks which keys are pressed and released at the current point of time.
type KeyManager struct {
	mapper map[Key]KeyState
	mutex  sync.RWMutex
}

// Set is used for updating whether or not a key is held down, or not held down.
func (km *KeyManager) Set(k Key, state bool) {
	km.mutex.Lock()

	ks := km.mapper[k]
	ks.set(state)
	km.mapper[k] = ks

	km.mutex.Unlock()
}

// Get retrieves a keys state.
func (km *KeyManager) Get(k Key) KeyState {
	km.mutex.RLock()
	defer km.mutex.RUnlock()

	ks, ok := km.mapper[k]
	if !ok {
		return KeyState{false, false}
	}

	return ks
}

func (km *KeyManager) update() {
	km.mutex.Lock()

	// Set all keys to their current states
	for key, state := range km.mapper {
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
	if !key.lastState && key.currentState {
		return KeyStateJustDown
	} else if key.lastState && !key.currentState {
		return KeyStateJustUp
	} else if key.lastState && key.currentState {
		return KeyStateDown
	} else if !key.lastState && !key.currentState {
		return KeyStateUp
	}

	return KeyStateUp
}

// JustPressed returns whether a key was just pressed
func (key KeyState) JustPressed() bool {
	return key.State() == KeyStateJustDown
}

// JustReleased returns whether a key was just released
func (key KeyState) JustReleased() bool {
	return key.State() == KeyStateJustUp
}

// Up returns wheter a key is not being pressed
func (key KeyState) Up() bool {
	return key.State() == KeyStateUp
}

// Down returns wether a key is being pressed
func (key KeyState) Down() bool {
	return key.State() == KeyStateDown
}
