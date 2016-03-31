package main

import (
	"image"
	"image/color"
	"log"

	"github.com/paked/engi"
	"github.com/paked/engi/ecs"
)

type Game struct{}

var (
	zoomSpeed   float32 = -0.125
	worldWidth  float32 = 800
	worldHeight float32 = 800
)

// generateBackground creates a background of green tiles - might not be the most efficient way to do this
func generateBackground() *ecs.Entity {
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
	field := ecs.NewEntity([]string{"RenderSystem"})
	fieldRender := engi.NewRenderComponent(engi.NewTexture(bgTexture), engi.Point{1, 1}, "Background1")
	fieldRender.SetPriority(engi.Background)
	fieldSpace := &engi.SpaceComponent{Position: engi.Point{0, 0}, Width: worldWidth, Height: worldHeight}
	field.AddComponent(fieldRender)
	field.AddComponent(fieldSpace)
	return field
}

func (game *Game) Preload() {}

// Setup is called before the main loop is started
func (game *Game) Setup(w *ecs.World) {
	engi.SetBg(color.White)
	w.AddSystem(&engi.RenderSystem{})
	w.AddSystem(&engi.MouseZoomer{zoomSpeed})

	// Create the background; this way we'll see when we actually zoom
	err := w.AddEntity(generateBackground())
	if err != nil {
		log.Println(err)
	}
}

func (*Game) Hide()        {}
func (*Game) Show()        {}
func (*Game) Type() string { return "Game" }

func main() {
	opts := engi.RunOptions{
		Title:  "Zoom Demo",
		Width:  400,
		Height: 400,
	}
	engi.Run(opts, &Game{})
}
