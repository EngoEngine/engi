//+build vulkan

package engo

import (
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

var (
	// Window is the glfw.Window used for engo
	Window *glfw.Window

	// Device is the VkDevice used for rendering
	Device *VkDevice

	cursorArrow     *glfw.Cursor
	cursorIBeam     *glfw.Cursor
	cursorCrosshair *glfw.Cursor
	cursorHand      *glfw.Cursor
	cursorHResize   *glfw.Cursor
	cursorVResize   *glfw.Cursor

	scale = float32(1)

	engoVersion = []int{1, 0, 5}
)

func init() {
	runtime.LockOSThread()
}

// fatalErr calls log.Fatal with the given error if it is non-nil.
func fatalErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// CreateWindow sets up the GLFW window and prepares the OpenGL surface for rendering
func CreateWindow(title string, width, height int, fullscreen bool, msaa int) {
	CurrentBackEnd = BackEndVulkan
	err := glfw.Init()
	fatalErr(err)

	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	fatalErr(vk.Init())

	if !opts.HeadlessMode {
		cursorArrow = glfw.CreateStandardCursor(int(glfw.ArrowCursor))
		cursorIBeam = glfw.CreateStandardCursor(int(glfw.IBeamCursor))
		cursorCrosshair = glfw.CreateStandardCursor(int(glfw.CrosshairCursor))
		cursorHand = glfw.CreateStandardCursor(int(glfw.HandCursor))
		cursorHResize = glfw.CreateStandardCursor(int(glfw.HResizeCursor))
		cursorVResize = glfw.CreateStandardCursor(int(glfw.VResizeCursor))
	}

	monitor := glfw.GetPrimaryMonitor()

	var mode *glfw.VidMode
	if monitor != nil {
		mode = monitor.GetVideoMode()
	} else {
		// Initialize default values if no monitor is found
		mode = &glfw.VidMode{
			Width:       1,
			Height:      1,
			RedBits:     8,
			GreenBits:   8,
			BlueBits:    8,
			RefreshRate: 60,
		}
	}

	gameWidth = float32(width)
	gameHeight = float32(height)

	if fullscreen {
		width = mode.Width
		height = mode.Height
		glfw.WindowHint(glfw.Decorated, 0)
	} else {
		monitor = nil
	}

	if opts.HeadlessMode {
		glfw.WindowHint(glfw.Visible, glfw.False)
	}
	if opts.NotResizable {
		glfw.WindowHint(glfw.Resizable, glfw.False)
	}

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)

	if opts.HeadlessMode {
		return
	}

	Window, err = glfw.CreateWindow(width, height, title, monitor, nil)
	fatalErr(err)

	err = Device.init()
	fatalErr(err)

	if !fullscreen {
		Window.SetPos((mode.Width-width)/2, (mode.Height-height)/2)
	}

	width, height = Window.GetSize()
	windowWidth, windowHeight = float32(width), float32(height)

	fw, fh := Window.GetFramebufferSize()
	canvasWidth, canvasHeight = float32(fw), float32(fh)

	if windowWidth <= canvasWidth && windowHeight <= canvasHeight {
		scale = canvasWidth / windowWidth
	}

	Window.SetFramebufferSizeCallback(func(Window *glfw.Window, w, h int) {
		width, height = Window.GetSize()
		windowWidth, windowHeight = float32(width), float32(width)

		oldCanvasW, oldCanvasH := canvasWidth, canvasHeight

		canvasWidth, canvasHeight = float32(w), float32(h)

		ResizeXOffset += oldCanvasW - canvasWidth
		ResizeYOffset += oldCanvasH - canvasHeight

		if windowWidth <= canvasWidth && windowHeight <= canvasHeight {
			scale = canvasWidth / windowWidth
		}
	})

	Window.SetCursorPosCallback(func(Window *glfw.Window, x, y float64) {
		Input.Mouse.X, Input.Mouse.Y = float32(x)/opts.GlobalScale.X, float32(y)/opts.GlobalScale.Y
		if Input.Mouse.Action != Release && Input.Mouse.Action != Press {
			Input.Mouse.Action = Move
		}
	})

	Window.SetMouseButtonCallback(func(Window *glfw.Window, b glfw.MouseButton, a glfw.Action, m glfw.ModifierKey) {
		x, y := Window.GetCursorPos()
		Input.Mouse.X, Input.Mouse.Y = float32(x)/(opts.GlobalScale.X), float32(y)/(opts.GlobalScale.Y)

		// this is only valid because we use an internal structure that is
		// 100% compatible with glfw3.h
		Input.Mouse.Button = MouseButton(b)
		Input.Mouse.Modifer = Modifier(m)

		if a == glfw.Press {
			Input.Mouse.Action = Press
		} else {
			Input.Mouse.Action = Release
		}
	})

	Window.SetScrollCallback(func(Window *glfw.Window, xoff, yoff float64) {
		Input.Mouse.ScrollX = float32(xoff)
		Input.Mouse.ScrollY = float32(yoff)
	})

	Window.SetKeyCallback(func(Window *glfw.Window, k glfw.Key, s int, a glfw.Action, m glfw.ModifierKey) {
		key := Key(k)
		if a == glfw.Press {
			Input.keys.Set(key, true)
		} else if a == glfw.Release {
			Input.keys.Set(key, false)
		}
	})

	Window.SetSizeCallback(func(w *glfw.Window, widthInt int, heightInt int) {
		message := WindowResizeMessage{
			OldWidth:  int(windowWidth),
			OldHeight: int(windowHeight),
			NewWidth:  widthInt,
			NewHeight: heightInt,
		}

		windowWidth = float32(widthInt)
		windowHeight = float32(heightInt)

		// TODO: verify these for retina displays & verify if needed here
		fw, fh := Window.GetFramebufferSize()
		canvasWidth, canvasHeight = float32(fw), float32(fh)

		if !opts.ScaleOnResize {
			gameWidth, gameHeight = float32(widthInt), float32(heightInt)
		}

		Mailbox.Dispatch(message)
	})

	Window.SetCharCallback(func(Window *glfw.Window, char rune) {
		Mailbox.Dispatch(TextMessage{char})
	})

	Window.SetCloseCallback(func(Window *glfw.Window) {
		Exit()
	})
}

// DestroyWindow handles the termination of windows
func DestroyWindow() {
	glfw.Terminate()
}

// SetTitle sets the title of the window
func SetTitle(title string) {
	if opts.HeadlessMode {
		log.Println("Title set to:", title)
	} else {
		Window.SetTitle(title)
	}
}

// RunIteration runs one iteration per frame
func RunIteration() {
	Time.Tick()

	// First check for new keypresses
	if !opts.HeadlessMode {
		Input.update()
		glfw.PollEvents()
	}

	// Then update the world and all Systems
	currentUpdater.Update(Time.Delta())

	// Lastly, forget keypresses and swap buffers
	if !opts.HeadlessMode {
		// reset values to avoid catching the same "signal" twice
		Input.Mouse.ScrollX, Input.Mouse.ScrollY = 0, 0
		Input.Mouse.Action = Neutral
	}
}

// RunPreparation is called automatically when calling Open. It should only be called once.
func RunPreparation(defaultScene Scene) {
	Time = NewClock()
	SetScene(defaultScene, false)
}

func runLoop(defaultScene Scene, headless bool) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		closeEvent()
	}()

	RunPreparation(defaultScene)
	ticker := time.NewTicker(time.Duration(int(time.Second) / opts.FPSLimit))

	// Start tick, minimize the delta
	Time.Tick()

	for {
		select {
		case <-ticker.C:
			RunIteration()
		case <-resetLoopTicker:
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(int(time.Second) / opts.FPSLimit))
		case <-closeGame:
			ticker.Stop()
			closeEvent()
			return
		}
	}
}

// CursorPos returns the current cursor position
func CursorPos() (x, y float32) {
	w, h := Window.GetCursorPos()
	return float32(w), float32(h)
}

// WindowSize gets the current window size
func WindowSize() (w, h int) {
	return Window.GetSize()
}

// WindowWidth gets the current window width
func WindowWidth() float32 {
	return windowWidth
}

// WindowHeight gets the current window height
func WindowHeight() float32 {
	return windowHeight
}

// CanvasWidth gets the width of the current OpenGL Framebuffer
func CanvasWidth() float32 {
	return canvasWidth
}

// CanvasHeight gets the height of the current OpenGL Framebuffer
func CanvasHeight() float32 {
	return canvasHeight
}

// CanvasScale gets the ratio of the canvas to the window sizes
func CanvasScale() float32 {
	return scale
}

// SetCursor sets the pointer of the mouse to the defined standard cursor
func SetCursor(c Cursor) {
	var cur *glfw.Cursor
	switch c {
	case CursorNone:
		cur = nil // just for the documentation, this isn't required
	case CursorArrow:
		cur = cursorArrow
	case CursorCrosshair:
		cur = cursorCrosshair
	case CursorHand:
		cur = cursorHand
	case CursorIBeam:
		cur = cursorIBeam
	case CursorHResize:
		cur = cursorHResize
	case CursorVResize:
		cur = cursorVResize
	}
	Window.SetCursor(cur)
}

//SetCursorVisibility sets the visibility of the cursor.
//If true the cursor is visible, if false the cursor is not.
func SetCursorVisibility(visible bool) {
	if visible {
		glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	} else {
		glfw.GetCurrentContext().SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	}
}

// openFile is the desktop-specific way of opening a file
func openFile(url string) (io.ReadCloser, error) {
	return os.Open(url)
}

// IsAndroidChrome tells if the browser is Chrome for android
func IsAndroidChrome() bool {
	return false
}

// GetKeyName returns the string returned from the given Key. So you can write "Press W to move forward"
// and get a W for QWERTY and Z for AZERTY
func GetKeyName(k Key) string {
	m := map[Key]glfw.Key{
		KeyGrave:        glfw.KeyGraveAccent,
		KeyDash:         glfw.KeyMinus,
		KeyApostrophe:   glfw.KeyApostrophe,
		KeySemicolon:    glfw.KeySemicolon,
		KeyEquals:       glfw.KeyEqual,
		KeyComma:        glfw.KeyComma,
		KeyPeriod:       glfw.KeyPeriod,
		KeySlash:        glfw.KeySlash,
		KeyBackslash:    glfw.KeyBackslash,
		KeyBackspace:    glfw.KeyBackspace,
		KeyTab:          glfw.KeyTab,
		KeyCapsLock:     glfw.KeyCapsLock,
		KeySpace:        glfw.KeySpace,
		KeyEnter:        glfw.KeyEnter,
		KeyEscape:       glfw.KeyEscape,
		KeyInsert:       glfw.KeyInsert,
		KeyPrintScreen:  glfw.KeyPrintScreen,
		KeyDelete:       glfw.KeyDelete,
		KeyPageUp:       glfw.KeyPageUp,
		KeyPageDown:     glfw.KeyPageDown,
		KeyHome:         glfw.KeyHome,
		KeyEnd:          glfw.KeyEnd,
		KeyPause:        glfw.KeyPause,
		KeyScrollLock:   glfw.KeyScrollLock,
		KeyArrowLeft:    glfw.KeyLeft,
		KeyArrowRight:   glfw.KeyRight,
		KeyArrowDown:    glfw.KeyDown,
		KeyArrowUp:      glfw.KeyUp,
		KeyLeftBracket:  glfw.KeyLeftBracket,
		KeyLeftShift:    glfw.KeyLeftShift,
		KeyLeftControl:  glfw.KeyLeftControl,
		KeyLeftSuper:    glfw.KeyLeftSuper,
		KeyLeftAlt:      glfw.KeyLeftAlt,
		KeyRightBracket: glfw.KeyRightBracket,
		KeyRightShift:   glfw.KeyRightShift,
		KeyRightControl: glfw.KeyRightControl,
		KeyRightSuper:   glfw.KeyRightSuper,
		KeyRightAlt:     glfw.KeyRightAlt,
		KeyZero:         glfw.Key0,
		KeyOne:          glfw.Key1,
		KeyTwo:          glfw.Key2,
		KeyThree:        glfw.Key3,
		KeyFour:         glfw.Key4,
		KeyFive:         glfw.Key5,
		KeySix:          glfw.Key6,
		KeySeven:        glfw.Key7,
		KeyEight:        glfw.Key8,
		KeyNine:         glfw.Key9,
		KeyF1:           glfw.KeyF1,
		KeyF2:           glfw.KeyF2,
		KeyF3:           glfw.KeyF3,
		KeyF4:           glfw.KeyF4,
		KeyF5:           glfw.KeyF5,
		KeyF6:           glfw.KeyF6,
		KeyF7:           glfw.KeyF7,
		KeyF8:           glfw.KeyF8,
		KeyF9:           glfw.KeyF9,
		KeyF10:          glfw.KeyF10,
		KeyF11:          glfw.KeyF11,
		KeyF12:          glfw.KeyF12,
		KeyA:            glfw.KeyA,
		KeyB:            glfw.KeyB,
		KeyC:            glfw.KeyC,
		KeyD:            glfw.KeyD,
		KeyE:            glfw.KeyE,
		KeyF:            glfw.KeyF,
		KeyG:            glfw.KeyG,
		KeyH:            glfw.KeyH,
		KeyI:            glfw.KeyI,
		KeyJ:            glfw.KeyJ,
		KeyK:            glfw.KeyK,
		KeyL:            glfw.KeyL,
		KeyM:            glfw.KeyM,
		KeyN:            glfw.KeyN,
		KeyO:            glfw.KeyO,
		KeyP:            glfw.KeyP,
		KeyQ:            glfw.KeyQ,
		KeyR:            glfw.KeyR,
		KeyS:            glfw.KeyS,
		KeyT:            glfw.KeyT,
		KeyU:            glfw.KeyU,
		KeyV:            glfw.KeyV,
		KeyW:            glfw.KeyW,
		KeyX:            glfw.KeyX,
		KeyY:            glfw.KeyY,
		KeyZ:            glfw.KeyZ,
		KeyNumLock:      glfw.KeyNumLock,
		KeyNumMultiply:  glfw.KeyKPMultiply,
		KeyNumDivide:    glfw.KeyKPDivide,
		KeyNumAdd:       glfw.KeyKPAdd,
		KeyNumSubtract:  glfw.KeyKPSubtract,
		KeyNumZero:      glfw.KeyKP0,
		KeyNumOne:       glfw.KeyKP1,
		KeyNumTwo:       glfw.KeyKP2,
		KeyNumThree:     glfw.KeyKP3,
		KeyNumFour:      glfw.KeyKP4,
		KeyNumFive:      glfw.KeyKP5,
		KeyNumSix:       glfw.KeyKP6,
		KeyNumSeven:     glfw.KeyKP7,
		KeyNumEight:     glfw.KeyKP8,
		KeyNumNine:      glfw.KeyKP9,
		KeyNumDecimal:   glfw.KeyKPDecimal,
		KeyNumEnter:     glfw.KeyKPEnter,
	}
	return glfw.GetKeyName(m[k], 0)
}
