package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/demos/demoutils"
)

type DefaultScene struct{}

var (
	// rotationSpeed is the speed at which to rotate
	rotationSpeed float32 = 10

	worldWidth  int = 800
	worldHeight int = 800
)

func (*DefaultScene) Preload() {}

// Setup is called before the main loop is started
func (*DefaultScene) Setup(w *ecs.World) {
	engo.SetBackground(color.White)
	w.AddSystem(&engo.RenderSystem{})
	w.AddSystem(&engo.MouseRotator{RotationSpeed: rotationSpeed})

	// Create a background; this way we'll see when we actually rotate
	demoutils.NewBackground(w, 300, worldHeight, color.RGBA{102, 153, 0, 255}, color.RGBA{102, 173, 0, 255})

	// Create a background; this way we'll see when we actually rotate
	bg2 := demoutils.NewBackground(w, 300, worldHeight, color.RGBA{102, 153, 0, 255}, color.RGBA{102, 173, 0, 255})
	bg2.SpaceComponent.Position.X = 500
}

func (*DefaultScene) Type() string { return "Game" }

func main() {
	opts := engo.RunOptions{
		Title:  "MouseRotation Demo",
		Width:  worldWidth,
		Height: worldHeight,
	}
	engo.Run(opts, &DefaultScene{})
}
