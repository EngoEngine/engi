//+build vulkan

package common

import "sync/atomic"

type TextureID *VkTextureID

type VkTextureID uint64

// TexMap is a map of the texture IDs to the Image data. This is used for making
// custom shaders that use textures in Vulkan.
var TexMap = make(map[*VkTextureID]Image)

var idInc uint64

func createTextureID(img Image) TextureID {
	id := VkTextureID(atomic.AddUint64(&idInc, 1))
	TexMap[&id] = img
	return &id
}

// close texture in an api specific way
func (t Texture) close() {
	delete(texMap, t.Texture())
}
