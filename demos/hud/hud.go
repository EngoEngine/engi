package main

import (
	"image"
	"image/color"

	"github.com/paked/engi"
)

type Game struct{}

var (
	zoomSpeed   float32 = -0.125
	scrollSpeed float32 = 700
	worldWidth  float32 = 800
	worldHeight float32 = 800

	hudBackgroundPriority = engi.PriorityLevel(engi.HUDGround)
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
	fieldSpace := &engi.SpaceComponent{engi.Point{0, 0}, worldWidth, worldHeight}
	field.AddComponent(fieldRender)
	field.AddComponent(fieldSpace)
	return field
}

// generateHUDBackground creates a violet HUD on the left side of the screen - might be inefficient
func generateHUDBackground(width, height float32) *engi.Entity {
	rect := image.Rect(0, 0, int(width), int(height))
	img := image.NewNRGBA(rect)
	c1 := color.RGBA{255, 0, 255, 180}
	for i := rect.Min.X; i < rect.Max.X; i++ {
		for j := rect.Min.Y; j < rect.Max.Y; j++ {
			img.Set(i, j, c1)
		}
	}
	bgTexture := engi.NewImageObject(img)
	field := engi.NewEntity([]string{"RenderSystem"})
	fieldRender := engi.NewRenderComponent(engi.NewRegion(engi.NewTexture(bgTexture), 0, 0, int(width), int(height)), engi.Point{0.5, 0.5}, "HUDBackground1")
	fieldRender.Priority = hudBackgroundPriority
	fieldSpace := &engi.SpaceComponent{engi.Point{-1, -1}, width, height}
	field.AddComponent(fieldRender)
	field.AddComponent(fieldSpace)
	return field
}

// Scroll enables us to zoom in/out, when scrolling our mouse wheel
func (game *Game) Scroll(amount float32) {
	engi.Mailbox.Dispatch(engi.CameraMessage{Axis: engi.ZAxis, Value: amount * zoomSpeed, Incremental: true})
}

func (game *Game) Preload() {}

// Setup is called before the main loop is started
func (game *Game) Setup(w *engi.World) {
	engi.SetBg(0x222222)
	w.AddSystem(&engi.RenderSystem{})

	// Adding KeyboardScroller so we can actually see the difference between background and HUD when scrolling
	w.AddSystem(engi.NewKeyboardScroller(scrollSpeed, engi.W, engi.D, engi.S, engi.A))

	// Create background, so we can see difference between this and HUD
	w.AddEntity(generateBackground())

	// Creating the HUD
	hudWidth := float32(200)   // Can be anything you want
	hudHeight := engi.Height() // Can be anything you want

	// Generate something that uses the PriorityLevel HUDGround or up
	hudBg := generateHUDBackground(hudWidth, hudHeight)
	w.AddEntity(hudBg)
}

func main() {
	engi.Open("HUD Demo", 400, 400, false, &Game{})
}
