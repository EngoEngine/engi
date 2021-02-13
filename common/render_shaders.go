package common

import (
	"image/color"
	"log"
	"sync"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo/math"
)

var (
	// UnicodeCap is the amount of unicode characters the fonts will be able to use, starting from index 0.
	UnicodeCap = 200

	// DefaultShader is the shader picked when no other shader is used.
	DefaultShader = &basicShader{cameraEnabled: true}
	// HUDShader is the shader used for HUD elements.
	HUDShader = &basicShader{cameraEnabled: false}
	// LegacyShader is the shader used for drawing shapes.
	LegacyShader = &legacyShader{cameraEnabled: true}
	// LegacyHUDShader is the shader used for drawing shapes on the HUD.
	LegacyHUDShader = &legacyShader{cameraEnabled: false}
	// TextShader is the shader used to draw fonts from a FontAtlas
	TextShader = &textShader{cameraEnabled: true}
	// TextHUDShader is the shader used to draw fonts from a FontAtlas on the HUD.
	TextHUDShader = &textShader{cameraEnabled: false}
	// BlendmapShader is a shader used to create blendmaps
	BlendmapShader = &blendmapShader{cameraEnabled: true}
	shadersSet     bool
	atlasCache     = make(map[Font]FontAtlas)
	shaders        = []Shader{
		DefaultShader,
		HUDShader,
		LegacyShader,
		LegacyHUDShader,
		TextShader,
		TextHUDShader,
		BlendmapShader,
	}
)

const (
	// MaxSprites is the maximum number of sprites that can comprise a single batch.
	// 32767 is the max vertex index in OpenGL. Since each sprite has 4 vertices,
	// 32767 / 4 = 8191 max sprites.
	MaxSprites = 8191
	spriteSize = 20

	bufferSize = 10000
)

// Shader when implemented can be used in the RenderSystem as an OpenGl Shader.
//
// Setup holds the actual OpenGL shader data, and prepares any matrices and OpenGL calls for use.
//
// Pre is called just before the Draw step.
//
// Draw is the Draw step.
//
// Post is called just after the Draw step.
type Shader interface {
	Setup(*ecs.World) error
	Pre()
	Draw(*RenderComponent, *SpaceComponent)
	Post()
	SetCamera(*CameraSystem)
}

// CullingShader when implemented can be used in the RenderSystem to test if an entity should be rendered.
type CullingShader interface {
	Shader
	// PrepareCulling is called once per frame for the shader to initialize culling calculation.
	PrepareCulling()
	ShouldDraw(*RenderComponent, *SpaceComponent) bool
}

// TextureShader when implemented stores the textures in the shader. This is because
// vulkan ties the textures to the graphics pipeline, so one texture may be
// implemented several times depending on how many shaders use textures.
type TextureShader interface {
	AddTexture(d Drawable)
}

func setBufferValue(buffer []float32, index int, value float32, changed *bool) {
	if buffer[index] != value {
		buffer[index] = value
		*changed = true
	}
}

// colorToFloat32 returns the float32 representation of the given color
func colorToFloat32(c color.Color) float32 {
	colorR, colorG, colorB, colorA := c.RGBA()
	colorR >>= 8
	colorG >>= 8
	colorB >>= 8
	colorA >>= 8

	red := colorR
	green := colorG << 8
	blue := colorB << 16
	alpha := colorA << 24

	return math.Float32frombits((alpha | blue | green | red) & 0xfeffffff)
}

// AddShader adds a shader to the list of shaders for initalization. They should
// be added before the Rendersystem is added, such as in the scene's Preload.
func AddShader(s Shader) {
	shaders = append(shaders, s)
}

var shaderInitMutex sync.Mutex

func initShaders(w *ecs.World) error {
	shaderInitMutex.Lock()
	defer shaderInitMutex.Unlock()

	if !shadersSet {
		var err error

		for _, shader := range shaders {
			err = shader.Setup(w)
			if err != nil {
				return err
			}
		}

		shadersSet = true
	}
	return nil
}

func newCamera(w *ecs.World) {
	shaderInitMutex.Lock()
	defer shaderInitMutex.Unlock()
	var cam *CameraSystem
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *CameraSystem:
			cam = sys
		}
	}
	if cam == nil {
		log.Println("Camera system was not found when changing scene!")
		return
	}
	for _, shader := range shaders {
		shader.SetCamera(cam)
	}
}
