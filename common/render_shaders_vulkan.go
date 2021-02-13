//+build vulkan

package common

import (
	"errors"

	vk "github.com/vulkan-go"
)

// LoadShaderModule takes in the shader as a byte slice and returns a
// vk.ShaderModule corresponding to the shader.
func LoadShaderModule(data []byte) (vk.ShaderModule, error) {
	var module vk.ShaderModule
	if res := vk.CreateShaderModule(r.device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint(len(data)),
		PCode:    sliceUint32(data),
	}, nil, &module); res != vk.Success {
		return vk.NullShaderModule, errors.New("unable to create shader module")
	}
	return module, nil
}

var end = "\x00"
var endChar byte = '\x00'

// SafeString returns a string safe for passing to C (Vulkan)
func SafeString(s string) string {
	if len(s) == 0 {
		return end
	}
	if s[len(s)-1] != endChar {
		return s + end
	}
	return s
}

// SafeStrings returns a slice of strings safe for passing to C (Vulkan)
func SafeStrings(list []string) []string {
	for i := range list {
		list[i] = safeString(list[i])
	}
	return list
}
