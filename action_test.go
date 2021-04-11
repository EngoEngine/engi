package engo

import "testing"

var (
	actions = [...]InputAction{
		InputAction{Name: "Action 1", Keys: []Key{KeyA, KeyC}, MouseButtons: []MouseButton{MouseButtonLeft}},
		InputAction{Name: "Action 2", Keys: []Key{KeyB, KeyD}, MouseButtons: []MouseButton{MouseButtonLeft}},
		InputAction{Name: "Action 3", Keys: []Key{KeyF2, KeyF5}, MouseButtons: []MouseButton{MouseButtonLeft}},
		InputAction{Name: "Action 4", Keys: []Key{KeyF4, KeyF6}, MouseButtons: []MouseButton{MouseButtonLeft}},
		InputAction{Name: "Action 5", Keys: []Key{KeyOne, KeyFour}, MouseButtons: []MouseButton{MouseButtonLeft}},
		InputAction{Name: "Action 6", Keys: []Key{KeyTwo, KeyFive}, MouseButtons: []MouseButton{MouseButtonLeft}},
	}

	// Expected button state @ pass 0
	keyPass0 = [6]btnState{
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
	}

	// Expected button state @ pass 1
	keyPass1 = [6]btnState{
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: true},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: true},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: true},
	}

	// Expected button state @ pass 2
	keyPass2 = [6]btnState{
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: true, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: true, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: true, justUp: false, justDown: false},
	}

	// Expected button state @ pass 3
	keyPass3 = [6]btnState{
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: true, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: true, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: true, justDown: false},
	}

	// Expected button state @ pass 4
	keyPass4 = [6]btnState{
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
	}

	// Expected button state @ pass 0
	mousePass0 = [6]btnState{
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: true, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: true, justUp: false, justDown: false},
		btnState{down: false, justUp: false, justDown: false},
		btnState{down: true, justUp: false, justDown: false},
	}
)

func runActionKeyChecks(msg string, t *testing.T, expect [6]btnState) {
	for i, action := range actions {
		exp := expect[i]
		act := Input.Action(action.Name)
		if exp.down != act.Pressed() {
			t.Error(msg, " Invalid on: ", action.Name, " - Down")
		}
		if exp.justUp != act.JustReleased() {
			t.Error(msg, " Invalid on: ", action.Name, " - Just Released")
		}
		if exp.justDown != act.JustPressed() {
			t.Error(msg, " Invalid on: ", action.Name, " - Just Pressed")
		}
	}
}

func TestInputAction_Keys(t *testing.T) {
	Input = NewInputManager()

	for _, action := range actions {
		Input.RegisterAction(action)
	}
	Input.update()
	runActionKeyChecks("Init (0.0)", t, keyPass0)
	// Empty update pass0
	Input.update()
	runActionKeyChecks("Pass (0.1)", t, keyPass0)
	Input.update()
	runActionKeyChecks("Pass (0.2)", t, keyPass0)
	Input.update()
	runActionKeyChecks("Pass (0.3)", t, keyPass0)

	// Set even true pass1
	Input.update()
	Input.keys.Set(actions[1].Keys[0], true)
	Input.keys.Set(actions[3].Keys[0], true)
	Input.keys.Set(actions[5].Keys[0], true)

	// FixMe: this causes an error ? Because the the static
	// arrays get filled with values before engo rewrites them!
	//Input.keys.Set(F10, true)

	runActionKeyChecks("Pass (1.0)", t, keyPass1)

	// Keeps state on pass2
	Input.update()
	runActionKeyChecks("Pass (2.0)", t, keyPass2)
	Input.update()
	runActionKeyChecks("Pass (2.1)", t, keyPass2)
	Input.update()
	runActionKeyChecks("Pass (2.2)", t, keyPass2)
	Input.update()
	runActionKeyChecks("Pass (2.3)", t, keyPass2)

	// Set even true pass3
	Input.update()
	Input.keys.Set(actions[1].Keys[0], false)
	Input.keys.Set(actions[3].Keys[0], false)
	Input.keys.Set(actions[5].Keys[0], false)

	runActionKeyChecks("Pass (3.0)", t, keyPass3)

	// Keeps state on pass4
	Input.update()
	runActionKeyChecks("Pass (4.0)", t, keyPass4)
	Input.update()
	runActionKeyChecks("Pass (4.1)", t, keyPass4)
	Input.update()
	runActionKeyChecks("Pass (4.2)", t, keyPass4)
	Input.update()
	runActionKeyChecks("Pass (4.3)", t, keyPass4)
}
