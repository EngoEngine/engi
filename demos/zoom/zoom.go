package main

import (
	"image"
	"image/color"

	"github.com/paked/engi"
)

type Game struct{}

var (
	zoomSpeed   float32 = -0.125
	worldWidth  float32 = 800
	worldHeight float32 = 800
)

// generateBackground creates a background of green tiles - might not be the most efficient way to do this
func generateBackground() *engi.Entity {
	rect := image.Rect(0, 0, int(worldWidth), int(worldHeight))
	img := image.NewNRGBA(rect)
	c1 := color.RGBA{102, 153, 0, 255}
	c2 := color.RGBA{102, 173, 0, 255}
	for i := rect.Min.X; i < rect.Max.X; i++ {
		for j := rect.Min.Y; j < rect.Max.Y; j++ {
			if i%40 > 20 {
				if j%40 > 20 {
					img.Set(i, j, c1)
				} else {
					img.Set(i, j, c2)
				}
			} else {
				if j%40 > 20 {
					img.Set(i, j, c2)
				} else {
					img.Set(i, j, c1)
				}
			}
		}
	}
	bgTexture := engi.NewImageObject(img)
	field := engi.NewEntity([]string{"RenderSystem"})
	fieldRender := engi.NewRenderComponent(engi.NewRegion(engi.NewTexture(bgTexture), 0, 0, int(worldWidth), int(worldHeight)), engi.Point{1, 1}, "Background1")
	fieldRender.Priority = engi.Background
	fieldSpace := engi.SpaceComponent{engi.Point{0, 0}, worldWidth, worldHeight}
	field.AddComponent(&fieldRender)
	field.AddComponent(&fieldSpace)
	return field
}

// TODO: deprecated
// Scroll is called whenever the mouse wheel scrolls
func (game *Game) Scroll(amount float32) {
	// Adding this line, allows for zooming on scrolling the mouse wheel
	engi.Mailbox.Dispatch(engi.CameraMessage{Axis: engi.ZAxis, Value: amount * zoomSpeed, Incremental: true})
}

func (game *Game) Preload() {}

// Setup is called before the main loop is started
func (game *Game) Setup(w *engi.World) {
	engi.SetBg(0x222222)
	w.AddSystem(&engi.RenderSystem{})

	// Explicitly set WorldBounds for better default CameraSystem values
	engi.WorldBounds.Max = engi.Point{worldWidth, worldHeight}

	// Create the background; this way we'll see when we actually zoom
	w.AddEntity(generateBackground())
}

func main() {
	engi.Open("Zoom Demo", 400, 400, false, &Game{})
}
