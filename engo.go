package engo

import (
	"fmt"
	"log"
	"sync"

	"github.com/EngoEngine/ecs"
)

// BackEnd represents the back end used for the window management / GL Surface
type BackEnd uint

const (
	// BackEndGLFW uses glfw
	BackEndGLFW BackEnd = iota
	// BackEndWeb uses gopherjs
	BackEndWeb
	// BackEndMobile uses gomobile
	BackEndMobile
	// BackEndSDL uses sdl2
	BackEndSDL
	// BackEndHeadless does not use a window manager for the backend
	BackEndHeadless
	// BackEndVulkan uses glfw 3.3 from vulkan-go and vulkan as a render surface
	BackEndVulkan
)

var (
	// Time is the active FPS counter
	Time *Clock

	// Input handles all input: mouse, keyboard and touch
	Input *InputManager

	// Mailbox is used by all Systems to communicate
	Mailbox *MessageManager

	currentUpdater            Updater
	currentScene              Scene
	sceneMutex                = &sync.RWMutex{}
	opts                      RunOptions
	resetLoopTicker           = make(chan bool, 1)
	closeGame                 = make(chan struct{})
	closeGameOnce             sync.Once
	gameWidth, gameHeight     float32
	windowWidth, windowHeight float32
	canvasWidth, canvasHeight float32
	headlessWidth             = 800
	headlessHeight            = 800

	// CurrentBackEnd is the current back end used for window management
	CurrentBackEnd BackEnd
	// ResizeXOffset is how far the screen moves from (0,0) being the top-left corner
	// when the window is resized
	ResizeXOffset = float32(0)
	// ResizeYOffset is how far the screen moves from (0,0) being the top-left corner
	// when the window is resized
	ResizeYOffset = float32(0)
)

const (
	// DefaultVerticalAxis is the name of the default vertical axis, as used internally in `engo` when `StandardInputs`
	// is defined.
	DefaultVerticalAxis = "vertical"

	// DefaultHorizontalAxis is the name of the default horizontal axis, as used internally in `engo` when `StandardInputs`
	// is defined.
	DefaultHorizontalAxis = "horizontal"
	// DefaultMouseXAxis is the name of the default horizontal mouse axis
	DefaultMouseXAxis = "mouse x"
	// DefaultMouseYAxis is the name of the default vertical mouse axis
	DefaultMouseYAxis = "mouse y"
)

// RunOptions are the options used to Run engo
type RunOptions struct {
	// NoRun indicates the Open function should return immediately, without looping
	NoRun bool

	// Title is the Window title
	Title string

	// HeadlessMode indicates whether or not OpenGL calls should be made
	HeadlessMode bool

	// Fullscreen indicates the game should run in fullscreen mode if run on a desktop
	Fullscreen bool

	Width, Height int

	// GlobalScale scales all size/render components by the scale factor
	// Any point passed less than or equal to zero will result in the scale being set to
	// engo.Point{1, 1}.
	// All the systems in common should scale themselves accordingly (collision, camera, render, etc)
	// However, custom systems should be aware of this if this is set.
	GlobalScale Point

	// VSync indicates whether or not OpenGL should wait for the monitor to swp the buffers
	VSync bool

	// Resizable indicates whether or not the Window should be resizable.  Defaults to `false`.
	NotResizable bool

	// ScaleOnResize indicates whether or not engo should make things larger/smaller whenever the screen resizes
	ScaleOnResize bool

	// FPSLimit indicates the maximum number of frames per second
	FPSLimit int

	// OverrideCloseAction indicates that (when true) engo will never close whenever the gamer wants to close the
	// game - that will be your responsibility
	OverrideCloseAction bool

	// StandardInputs is an easy way to map common inputs to actions, such as "jump" being <SPACE>, and "action" being
	// <ENTER>.
	StandardInputs bool

	// MSAA indicates the amount of samples that should be taken. Leaving it blank will default to 1, and you may
	// use any positive value you wish. It may be possible that the operating system / environment doesn't support
	// the requested amount. In that case, GLFW will (hopefully) pick the highest supported sampling count. The higher
	// the value, the bigger the performance cost.
	//
	// Our `RenderSystem` automatically calls `gl.Enable(gl.MULTISAMPLE)` (which is required to make use of it), but
	// if you're going to use your own rendering `System` instead, you will have to call it yourself.
	//
	// Also note that this value is entirely ignored in WebGL - most browsers enable it by default when available, and
	// none of them (at time of writing) allow you to tune it.
	//
	// More info at https://www.opengl.org/wiki/Multisampling
	// "With multisampling, each pixel at the edge of a polygon is sampled multiple times."
	MSAA int

	// AssetsRoot is the path where all resources (images, audio files, fonts, etc.) can be found. Leaving this at
	// empty-string, will default this to `assets`.
	//
	// Whenever using any value that does not start with the directory `assets`, you will not be able to support
	// mobile (Android/iOS), because they **require** all assets to be within the `assets` directory. You may however
	// use any subfolder-structure within that `assets` directory.
	AssetsRoot string

	// MobileWidth and MobileHeight are the width and height given from the Android/iOS OpenGL Surface used for Gomobile bind
	MobileWidth, MobileHeight int

	// Update is the function called each frame during the runLoop to update all of the
	// systems. If left blank, it defaults to &ecs.World{}. Use this if you plan on utilizing
	// engo's window / GL management but don't want to use the ECS paradigm.
	Update Updater

	// ApplicationXXXVersion is the major, minor, and revision versions of the game.
	// defaults to 0.0.0
	ApplicationMajorVersion, ApplicationMinorVersion, ApplicationRevisionVersion int

	// MaxFramesInFlight is the number of frames allowed to be in-flight. Places an
	// upper-bound on the amount of frames drawn between draw calls. Only works on
	// Vulkan. Defaults to 2.
	MaxFramesInFlight int
}

// Run is called to create a window, initialize everything, and start the main loop. Once this function returns,
// the game window has been closed already. You can supply a lot of options within `RunOptions`, and your starting
// `Scene` should be defined in `defaultScene`.
func Run(o RunOptions, defaultScene Scene) {
	// Setting defaults
	if o.FPSLimit == 0 {
		o.FPSLimit = 60
	}

	if o.MSAA < 0 {
		panic("MSAA has to be greater or equal to 0")
	}

	if o.MSAA == 0 {
		o.MSAA = 1
	}

	if len(o.AssetsRoot) == 0 {
		o.AssetsRoot = "assets"
	}

	if o.Update == nil {
		o.Update = &ecs.World{}
	}

	if o.GlobalScale.X <= 0 || o.GlobalScale.Y <= 0 {
		o.GlobalScale = Point{X: 1, Y: 1}
	}

	if o.MaxFramesInFlight <= 0 {
		o.MaxFramesInFlight = 2
	}

	opts = o

	// Create input
	Input = NewInputManager()
	if opts.StandardInputs {
		log.Println("Using standard inputs")

		Input.RegisterButton("jump", KeySpace)
		Input.RegisterButton("action", KeyEnter)

		Input.RegisterAxis(DefaultHorizontalAxis, AxisKeyPair{KeyA, KeyD}, AxisKeyPair{KeyArrowLeft, KeyArrowRight})
		Input.RegisterAxis(DefaultVerticalAxis, AxisKeyPair{KeyW, KeyS}, AxisKeyPair{KeyArrowUp, KeyArrowDown})

		Input.RegisterAxis(DefaultMouseXAxis, NewAxisMouse(AxisMouseHori))
		Input.RegisterAxis(DefaultMouseYAxis, NewAxisMouse(AxisMouseVert))
	}

	Files.SetRoot(opts.AssetsRoot)
	currentUpdater = opts.Update

	// And run the game
	if opts.HeadlessMode {
		if opts.Width == 0 {
			opts.Width = headlessWidth
		}
		if opts.Height == 0 {
			opts.Height = headlessHeight
		}
		windowWidth = float32(opts.Width)
		windowHeight = float32(opts.Height)
		gameWidth = float32(opts.Width)
		gameHeight = float32(opts.Height)
		canvasWidth = float32(opts.Width)
		canvasHeight = float32(opts.Height)

		if !opts.NoRun {
			runHeadless(defaultScene)
		} else {
			SetScene(defaultScene, true)
		}
	} else {
		CreateWindow(opts.Title, opts.Width, opts.Height, opts.Fullscreen, opts.MSAA)
		defer DestroyWindow()

		if !opts.NoRun {
			runLoop(defaultScene, false)
		}
	}
}

// SetScaleOnResize can be used to change the value in the given `RunOpts` after already having called `engo.Run`.
func SetScaleOnResize(b bool) {
	opts.ScaleOnResize = b
}

// SetOverrideCloseAction can be used to change the value in the given `RunOpts` after already having called `engo.Run`.
func SetOverrideCloseAction(value bool) {
	opts.OverrideCloseAction = value
}

// SetFPSLimit can be used to change the value in the given `RunOpts` after already having called `engo.Run`.
func SetFPSLimit(limit int) error {
	if limit <= 0 {
		return fmt.Errorf("FPS Limit out of bounds. Requires > 0")
	}
	opts.FPSLimit = limit
	resetLoopTicker <- true
	return nil
}

// Headless indicates whether or not OpenGL-calls should be made
func Headless() bool {
	return opts.HeadlessMode
}

// ScaleOnResize indicates whether or not the screen should resize (i.e. make things look smaller/bigger) whenever
// the window resized. If `false`, then the size of the screen does not affect the size of the things drawn - it just
// makes less/more objects visible
func ScaleOnResize() bool {
	return opts.ScaleOnResize
}

// Exit is the safest way to close your game, as `engo` will correctly attempt to close all windows, handlers and contexts
func Exit() {
	closeGameOnce.Do(func() {
		close(closeGame)
	})
}

// GameWidth returns the current game width
func GameWidth() float32 {
	return gameWidth
}

// GameHeight returns the current game height
func GameHeight() float32 {
	return gameHeight
}

func closeEvent() {
	sceneMutex.RLock()
	for _, scenes := range scenes {
		if exiter, ok := scenes.scene.(Exiter); ok {
			exiter.Exit()
		}
	}
	sceneMutex.RUnlock()

	if !opts.OverrideCloseAction {
		Exit()
	} else {
		log.Println("[WARNING] default close action set to false, please make sure you manually handle this")
	}
}

func runHeadless(defaultScene Scene) {
	runLoop(defaultScene, true)
}

// GetGlobalScale returns the GlobalScale factor set in the RunOptions or via
// SetGlobalScale()
func GetGlobalScale() Point {
	return opts.GlobalScale
}

// SetGlobalScale sets the GlobalScale to the given dimensions. If either dimension is
// less than or equal to zero, GlobalScale is set to (1, 1).
func SetGlobalScale(p Point) {
	if p.X <= 0 || p.Y <= 0 {
		opts.GlobalScale = Point{X: 1, Y: 1}
		return
	}
	opts.GlobalScale = p
}

// GetTitle returns the title of the game.
func GetTitle() string {
	return opts.Title
}

// GetApplicationVersion returns the major, minor, and revision of the game.
func GetApplicationVersion() [3]int {
	return [3]int{opts.ApplicationMajorVersion, opts.ApplicationMinorVersion, opts.ApplicationRevisionVersion}
}

// GetMaxFramesInFlight returns the maximum number of frames to be drawn while drawing.
// Only works on Vulkan.
func GetMaxFramesInFlight() int {
	return opts.MaxFramesInFlight
}
