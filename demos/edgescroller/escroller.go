package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/demos/demoutils"
)

type DefaultScene struct{}

var (
	edgeMargin  float64 = 20
	scrollSpeed float32 = 700

<<<<<<< HEAD
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

	bgTexture := engo.NewImageObject(img)
	field := ecs.NewEntity("RenderSystem")
	fieldRender := engo.NewRenderComponent(engo.NewTexture(bgTexture), engo.Point{1, 1})
	fieldSpace := &engo.SpaceComponent{
		Position: engo.Point{0, 0},
		Width:    worldWidth,
		Height:   worldHeight,
	}

	field.AddComponent(fieldRender)
	field.AddComponent(fieldSpace)
	return field
}
=======
	worldWidth  int = 800
	worldHeight int = 800
)
>>>>>>> 28393c45ef7ce198babe3c6854931398faaba25c

func (*DefaultScene) Preload() {}

// Setup is called before the main loop is started
func (*DefaultScene) Setup(w *ecs.World) {
	engo.SetBackground(color.White)
	w.AddSystem(&engo.RenderSystem{})

	// The most important line in this whole demo:
	w.AddSystem(&engo.EdgeScroller{scrollSpeed, edgeMargin})

	// Create the background; this way we'll see when we actually scroll
	demoutils.NewBackground(w, worldWidth, worldHeight, color.RGBA{102, 153, 0, 255}, color.RGBA{102, 173, 0, 255})
}

<<<<<<< HEAD
func (*Game) Hide()        {}
func (*Game) Show()        {}
func (*Game) Exit()        {}
func (*Game) Type() string { return "Game" }
=======
func (*DefaultScene) Type() string { return "Game" }
>>>>>>> 28393c45ef7ce198babe3c6854931398faaba25c

func main() {
	opts := engo.RunOptions{
		Title:  "EdgeScroller Demo",
<<<<<<< HEAD
		Width:  1024,
		Height: 640,
=======
		Width:  worldWidth,
		Height: worldHeight,
>>>>>>> 28393c45ef7ce198babe3c6854931398faaba25c
	}
	engo.Run(opts, &DefaultScene{})
}
