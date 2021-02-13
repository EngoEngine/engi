//+build !vulkan

package common

import (
	"image/color"

	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/gl"
)

type BufferData struct {
	// Buffer represents the buffer object itself
	// Avoid using it unless your are writing a custom shader
	Buffer *gl.Buffer
	// BufferContent contains the buffer data
	// Avoid using it unless your are writing a custom shader
	BufferContent []float32
}

func clearScreen() {
	engo.Gl.Clear(engo.Gl.COLOR_BUFFER_BIT)
}

func setBackground(c color.Color) {
	r, g, b, a := c.RGBA()

	engo.Gl.ClearColor(float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff)
}

func enableMultisample() {
	engo.Gl.Enable(engo.Gl.MULTISAMPLE)
}
