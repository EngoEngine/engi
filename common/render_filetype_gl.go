//+build !vulkan

package common

import (
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/gl"
)

type TextureID *gl.Texture

func createTextureID(img Image) TextureID {
	id := engo.Gl.CreateTexture()

	engo.Gl.BindTexture(engo.Gl.TEXTURE_2D, id)

	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_WRAP_S, engo.Gl.CLAMP_TO_EDGE)
	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_WRAP_T, engo.Gl.CLAMP_TO_EDGE)
	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_MIN_FILTER, engo.Gl.LINEAR)
	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_MAG_FILTER, engo.Gl.NEAREST)

	if img.Data() == nil {
		panic("Texture image data is nil.")
	}

	engo.Gl.TexImage2D(engo.Gl.TEXTURE_2D, 0, engo.Gl.RGBA, engo.Gl.RGBA, engo.Gl.UNSIGNED_BYTE, img.Data())
	return id
}

// close texture in an API specific way

func (t Texture) close() {
	engo.Gl.DeleteTexture(t.id)
}
