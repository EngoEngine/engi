//+build vulkan

package engo

import (
	vk "github.com/vulkan-go/vulkan"
)

func clamp(high, low, value uint32) uint32 {
	var ret uint32
	if value > high {
		ret = high
	} else {
		ret = value
	}
	if ret < low {
		return low
	}
	return ret
}

type swapChainSupportDetails struct {
	capabilities vk.SurfaceCapabilities
	formats      []vk.SurfaceFormat
	presentModes []vk.PresentMode
}

var details swapChainSupportDetails

var end = "\x00"
var endChar byte = '\x00'

func safeString(s string) string {
	if len(s) == 0 {
		return end
	}
	if s[len(s)-1] != endChar {
		return s + end
	}
	return s
}

func safeStrings(list []string) []string {
	for i := range list {
		list[i] = safeString(list[i])
	}
	return list
}
