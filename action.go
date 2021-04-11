package engo

// An InputAction is an abstract input which can be mapped to any number of keys and/or mouse buttons.InputAction
// InputActions represent player intent rather than a specific button. For instance, "move_up", "jump", "talk",
// "attack", etc... are all examples of appropriate action names. Similar to Buttons, at any given time, an action
// can be in the state JustPressed, JustReleased, or Pressed. While JustPressed and JustReleased are mutually exclusive,
// if an InputAction can be JustPressed and Pressed at the same time. Similarly, an Input action can be Pressed but
// neither JustPressed or JustReleased.
//
// Going forward, Actions are the intended means of managing input, as they provide the most flexibility to both
// developers and players.
type InputAction struct {
	Name         string
	Keys         []Key
	MouseButtons []MouseButton
}

// JustPressed checks whether at least one of the keys or mouse buttons associated with the action
// was just pressed in the previous frame.
func (a InputAction) JustPressed() bool {
	for _, key := range a.Keys {
		v := Input.keys.Get(key).JustPressed()
		if v {
			return v
		}
	}
	for _, btn := range a.MouseButtons {
		v := Input.mouse.Get(btn).JustPressed()
		if v {
			return v
		}
	}
	return false
}

// JustReleased checks whether at least one of the keys or mouse buttons associated with the action
// was just released in the previous frame.
func (a InputAction) JustReleased() bool {
	for _, key := range a.Keys {
		v := Input.keys.Get(key).JustReleased()
		if v {
			return v
		}
	}
	for _, btn := range a.MouseButtons {
		v := Input.mouse.Get(btn).JustReleased()
		if v {
			return v
		}
	}
	return false
}

// Pressed checks whether at least one of the keys or mouse buttons associated with the action
// is curently being held down.
func (a InputAction) Pressed() bool {
	for _, key := range a.Keys {
		v := Input.keys.Get(key).Down()
		if v {
			return v
		}
	}
	for _, btn := range a.MouseButtons {
		v := Input.mouse.Get(btn).Down()
		if v {
			return v
		}
	}
	return false
}
