package engo

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/EngoEngine/ecs"
)

type testScene struct{}

func (*testScene) Preload() {}

func (t *testScene) Setup(u Updater) {}

func (*testScene) Type() string { return "testScene" }

type testScene2 struct{}

func (*testScene2) Preload() {}

func (*testScene2) Setup(u Updater) {}

func (*testScene2) Type() string { return "testScene2" }

func (*testScene2) Hide() {
	log.Println("Hiding testScene2.")
}

func (*testScene2) Show() {
	log.Println("Showing testScene2.")
}

// The tests for engo.go all have to use the headless option. Non-headless stuff is not
// testable via the cl only, and those are taken care of by building the demos via Travis CI
func TestRunHeadlessNoRunDefaults(t *testing.T) {
	Run(RunOptions{
		NoRun:        true,
		HeadlessMode: true,
	}, &testScene{})

	if opts.FPSLimit != 60 {
		t.Error("FPSLimit was not defaulted to 60")
	}

	if opts.MSAA != 1 {
		t.Error("MSAA was not defaulted to 1")
	}

	if opts.AssetsRoot != "assets" {
		t.Error("AssetsRoot was not defaulted to assets")
	}

	if opts.Width != 800 {
		t.Error("Width was not defaulted to 800")
	}

	if opts.Height != 800 {
		t.Error("Height was not defaulted to 800")
	}

	if opts.ApplicationMajorVersion != 0 {
		t.Error("ApplicationMajorVersion was not defaulted to 0")
	}
	if opts.ApplicationMinorVersion != 0 {
		t.Error("ApplicationMinorVersion was not defaulted to 0")
	}
	if opts.ApplicationRevisionVersion != 0 {
		t.Error("ApplicationRevisionVersion was not defaulted to 0")
	}
	if opts.MaxFramesInFlight != 2 {
		t.Error("MaxFramesInFlight was not defaulted to 2")
	}
}

func TestSetScaleOnResize(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	if opts.ScaleOnResize {
		t.Error("ScaleOnResize didn't default to false.")
	}
	SetScaleOnResize(true)
	if !opts.ScaleOnResize {
		t.Error("SetScaleOnResize didn't set properly.")
	}
}

func TestSetOverrideCloseAction(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	if opts.OverrideCloseAction {
		t.Error("OverrideCloseAction didn't default to false.")
	}
	SetOverrideCloseAction(true)
	if !opts.OverrideCloseAction {
		t.Error("SetOverrideCloseAction didn't set properly.")
	}
}

func TestHeadless(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	if Headless() != opts.HeadlessMode {
		t.Error("Headless didn't return the proper value.")
	}
}

func TestScaleOnResize(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	if ScaleOnResize() != opts.ScaleOnResize {
		t.Error("ScaleOnResize didn't return the proper value.")
	}
}

func TestGameWidthHeight(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
		Width:        100,
		Height:       50,
	}, &testScene{})
	if GameWidth() != 100 {
		t.Error("Width didn't return the proper value.")
	}
	if GameHeight() != 50 {
		t.Error("Height didn't return the proper value.")
	}
}

func TestGetMaxFramesInFlight(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	f := GetMaxFramesInFlight()
	if opts.MaxFramesInFlight != f {
		t.Error("GetMaxFramesInFlight returned an incorrect value.")
	}
}

func TestSetFPSLimit(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	SetFPSLimit(5)
	if opts.FPSLimit != 5 {
		t.Error("SetFPSLimit didn't set properly.")
	}

	expected := "FPS Limit out of bounds. Requires > 0"
	if err := SetFPSLimit(-5); err == nil {
		t.Error("Error wasn't recieved when SetFPSLimit was set to a negative number.")
	} else if err.Error() != expected {
		t.Errorf("Wrong error recieved when SetFPSLimit was set to a negative number. want %v, got %v", expected, err.Error())
	}
}

func TestRunNegativeMSAAPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("No panic when MSAA was set to -5")
		}
		if r != "MSAA has to be greater or equal to 0" {
			t.Errorf("Wrong panic when MSAA was set to -5, got: %v", r)
		}
	}()
	Run(RunOptions{
		NoRun:        true,
		HeadlessMode: true,
		MSAA:         -5,
	}, &testScene{})
}

func TestRunStandardInputs(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	Run(RunOptions{
		NoRun:          true,
		HeadlessMode:   true,
		StandardInputs: true,
	}, &testScene{})

	expected := "Using standard inputs\n"
	if !strings.HasSuffix(buf.String(), expected) {
		t.Error("setting standard inputs did not write expected output to log")
	}
}

type testRunScene struct {
	updates int
}

func (*testRunScene) Preload() {}

func (t *testRunScene) Setup(u Updater) {
	w, _ := u.(*ecs.World)
	w.AddSystem(&testUpdate{updates: t.updates})
}

func (*testRunScene) Type() string { return "testRunScene" }

type testUpdate struct {
	updates, current int
}

func (*testUpdate) Remove(ecs.BasicEntity) {}

func (t *testUpdate) Update(float32) {
	t.current++
	if t.current >= t.updates {
		Exit()
	}
}

// This test tests running headless but also letting it go into the runloop
func TestRunHeadless(t *testing.T) {
	testChan := make(chan struct{})
	go func() {
		Run(RunOptions{
			HeadlessMode:        true,
			OverrideCloseAction: true,
		}, &testRunScene{1})
		testChan <- struct{}{}
	}()
	select {
	case <-testChan:
		return
	case <-time.After(1 * time.Second):
		t.Error("Timed out while waiting for Headless Run to return from loop. Exit wasn't called within 1 second.")
	}
}

func TestOverrideCloseAction(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	expected := "[WARNING] default close action set to false, please make sure you manually handle this\n"

	testChan := make(chan struct{})
	go func() {
		Run(RunOptions{
			HeadlessMode:        true,
			OverrideCloseAction: true,
		}, &testRunScene{1})
		testChan <- struct{}{}
	}()
	select {
	case <-testChan:
	case <-time.After(1 * time.Second):
		t.Error("Timed out while waiting for Headless Run to return from loop. Exit wasn't called within 1 second.")
	}

	if !strings.HasSuffix(buf.String(), expected) {
		t.Errorf("calling closeEvent with Override set did not write expected output to log. Wanted: %v, got: %v", expected, buf.String())
	}
}

func TestSetGlobalScale(t *testing.T) {
	data := []struct {
		in  Point
		exp Point
	}{
		{Point{X: 5, Y: 5}, Point{X: 5, Y: 5}},
		{Point{X: -5, Y: 5}, Point{X: 1, Y: 1}},
		{Point{X: 5, Y: -5}, Point{X: 1, Y: 1}},
		{Point{X: -5, Y: -5}, Point{X: 1, Y: 1}},
	}

	for _, d := range data {
		SetGlobalScale(d.in)
		if opts.GlobalScale.X != d.exp.X || opts.GlobalScale.Y != d.exp.Y {
			t.Errorf("SetGlobalScale did not set properly. was: %v, expected: %v", opts.GlobalScale, d.exp)
		}
	}
}

func TestSceneSwitching(t *testing.T) {
	RegisterScene(&testScene2{})
	var buf bytes.Buffer
	log.SetOutput(&buf)
	Run(RunOptions{
		NoRun:        true,
		HeadlessMode: true,
	}, &testScene{})
	SetSceneByName("testScene2", false)
	if CurrentScene().Type() != "testScene2" {
		t.Errorf("CurrentScene got wrong scene. was: %v, expected: %v", CurrentScene().Type(), "testScene2")
	}
	SetSceneByName("testScene", false)
	expected := "Hiding testScene2.\n"
	if !strings.HasSuffix(buf.String(), expected) {
		t.Errorf("Did not properly set testScene1 and hide testScene2. was %v, expected: %v", buf.String(), expected)
	}
	buf.Reset()
	SetSceneByName("testScene2", false)
	expected = "Showing testScene2.\n"
	if !strings.HasSuffix(buf.String(), expected) {
		t.Errorf("Did not properly set and show testScene2. was: %v, expected: %v", buf.String(), expected)
	}
	err := SetSceneByName("doesNotExistScene", true)
	if err == nil {
		t.Error("No error when setting scene that doesn't exist.")
	}
	expected = "scene not registered:"
	if !strings.HasPrefix(err.Error(), expected) {
		t.Errorf("Did not recieve correct error for setting a scene that doesn't exist. was: %v, expected:%v", err.Error(), expected)
	}
}

func TestUtils(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	expectedImpl := "is not yet implemented on this platform\n"
	expectedType := "type not supported\n"
	if notImplemented("testing "); !strings.HasSuffix(buf.String(), expectedImpl) {
		t.Errorf("Did not properly log notImplemented. got: %v", buf.String())
	}
	buf.Reset()
	if unsupportedType(); !strings.HasSuffix(buf.String(), expectedType) {
		t.Errorf("Did not properly log unsupportedType. got: %v", buf.String())
	}
}

func TestSetTitleHeadless(t *testing.T) {
	exp := "Title set to: test title\n"
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &testScene{})
	var buf bytes.Buffer
	log.SetOutput(&buf)
	if SetTitle("test title"); !strings.HasSuffix(buf.String(), exp) {
		t.Errorf("Did not properly log title set. Got: %v, wanted: %v", buf.String(), exp)
	}
}

func TestGetTitle(t *testing.T) {
	Run(RunOptions{
		HeadlessMode: true,
		NoRun:        true,
		Title:        "Title Test",
	}, &testScene{})
	if title := GetTitle(); title != "Title Test" {
		t.Errorf("GetTitle()  did not properly return title. Wanted: %v, got: %v", "Title Test", title)
	}
}

func TestGetApplicationVersion(t *testing.T) {
	Run(RunOptions{
		HeadlessMode:               true,
		NoRun:                      true,
		Title:                      "Title Test",
		ApplicationMajorVersion:    1,
		ApplicationMinorVersion:    2,
		ApplicationRevisionVersion: 3,
	}, &testScene{})
	ver := GetApplicationVersion()
	if ver[0] != 1 || ver[1] != 2 || ver[2] != 3 {
		t.Errorf("Application version did not match. Wanted: %v.%v.%v \n Got: %v.%v.%v\n", 1, 2, 3, ver[0], ver[1], ver[2])
	}
}
