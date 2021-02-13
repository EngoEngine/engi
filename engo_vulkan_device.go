//+build vulkan

package engo

import (
	"errors"

	vk "github.com/vulkan-go/vulkan"
)

// VkDevice represents the Vulkan Device as well as the state of the device necessary
// to run.
type VkDevice struct {
	instance             vk.Instance
	surface              vk.Surface
	gpu                  vk.PhysicalDevice
	device               vk.Device
	graphicsIdx          uint32
	graphicsQueue        vk.Queue
	presentIdx           uint32
	presentQueue         vk.Queue
	swapChain            vk.Swapchain
	images               []vk.Image
	swapChainImageFormat vk.Format
	swapChainExtent      vk.Extent2D
	swapChainImageViews  []vk.ImageView
}

// Device returns the vulkan virtual device
func (d *VkDevice) Device() vk.Device {
	return d.device
}

// GPU returns the vulkan physical device
func (d *VkDevice) GPU() vk.PhysicalDevice {
	return d.gpu
}

func (d *VkDevice) SwapChainImageFormat() vk.Format {
	return d.swapChainImageFormat
}

func (d *VkDevice) SwapChainExtent() vk.Extent2D {
	return d.swapChainExtent
}

func (d *VkDevice) GraphicsQueueIndex() uint32 {
	return d.graphicsIdx
}

func (d *VkDevice) GraphicsQueue() vk.Queue {
	return d.graphicsQueue
}

func (d *VkDevice) PresentQueueIndex() uint32 {
	return d.presentIdx
}

func (d *VkDevice) PresentQueue() vk.Queue {
	return d.presentQueue
}

func (d *VkDevice) init() error {
	if err := d.initVulkan(); err != nil {
		return err
	}
	if err := d.createSwapChain(); err != nil {
		return err
	}
	if err := d.createImageViews(); err != nil {
		return err
	}
	return nil
}

func (d *VkDevice) initVulkan() error {
	version := GetApplicationVersion()
	appInfo := vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PApplicationName:   safeString(GetTitle()),
		ApplicationVersion: vk.MakeVersion(version[0], version[1], version[2]),
		PEngineName:        safeString("engo engine"),
		EngineVersion:      vk.MakeVersion(engoVersion[0], engoVersion[1], engoVersion[2]),
		ApiVersion:         vk.ApiVersion10,
	}
	wantedExtensions := []string{
		vk.KhrSwapchainExtensionName,
	}
	createInfo := vk.InstanceCreateInfo{}
	createInfo.SType = vk.StructureTypeInstanceCreateInfo
	createInfo.PApplicationInfo = &appInfo
	exts := Window.GetRequiredInstanceExtensions()
	createInfo.EnabledExtensionCount = uint32(len(exts))
	createInfo.PpEnabledExtensionNames = exts
	if res := vk.CreateInstance(&createInfo, nil, &d.instance); res != vk.Success {
		return errors.New("unable to create vulkan instance")
	}
	if err := vk.InitInstance(d.instance); err != nil {
		return err
	}
	surfPtr, err := Window.CreateWindowSurface(d.instance, nil)
	d.surface = vk.SurfaceFromPointer(surfPtr)
	if err != nil {
		return err
	}
	var deviceCount uint32
	if res := vk.EnumeratePhysicalDevices(d.instance, &deviceCount, nil); res != vk.Success {
		return errors.New("unable to get physical devices")
	}
	devices := make([]vk.PhysicalDevice, deviceCount)
	if res := vk.EnumeratePhysicalDevices(d.instance, &deviceCount, devices); res != vk.Success {
		return errors.New("unable to get physical devices")
	}
	var deviceSelected bool
	var physicalDevice vk.PhysicalDevice
deviceLoop:
	for _, device := range devices {
		var queueFamilyPropertyCount uint32
		var graphicsSupport, presentSupport bool
		vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyPropertyCount, nil)
		if queueFamilyPropertyCount == 0 {
			continue
		}
		queueFamilyProperties := make([]vk.QueueFamilyProperties, queueFamilyPropertyCount)
		vk.GetPhysicalDeviceQueueFamilyProperties(device, &queueFamilyPropertyCount, queueFamilyProperties)
		for i, q := range queueFamilyProperties {
			q.Deref()
			if q.QueueFlags&vk.QueueFlags(vk.QueueGraphicsBit) != 0 {
				d.graphicsIdx = uint32(i)
				graphicsSupport = true
			}
			var b32PresentSupport vk.Bool32
			vk.GetPhysicalDeviceSurfaceSupport(device, uint32(i), d.surface, &b32PresentSupport)
			if b32PresentSupport.B() {
				presentSupport = true
				d.presentIdx = uint32(i)
			}
		}
		if !graphicsSupport || !presentSupport {
			continue
		}
		var extensionCount uint32
		vk.EnumerateDeviceExtensionProperties(device, "", &extensionCount, nil)
		if extensionCount == 0 {
			continue
		}
		availableExtensions := make([]vk.ExtensionProperties, extensionCount)
		vk.EnumerateDeviceExtensionProperties(device, "", &extensionCount, availableExtensions)
		for _, req := range wantedExtensions {
			extensionFound := false
			for _, ext := range availableExtensions {
				ext.Deref()
				if vk.ToString(ext.ExtensionName[:]) == req {
					extensionFound = true
					break
				}
			}
			if !extensionFound {
				continue deviceLoop
			}
		}
		if res := vk.GetPhysicalDeviceSurfaceCapabilities(device, d.surface, &details.capabilities); res != vk.Success {
			continue
		}
		var formatCount uint32
		vk.GetPhysicalDeviceSurfaceFormats(device, d.surface, &formatCount, nil)
		if formatCount == 0 {
			continue
		}
		details.formats = make([]vk.SurfaceFormat, formatCount)
		vk.GetPhysicalDeviceSurfaceFormats(device, d.surface, &formatCount, details.formats)
		var presentModeCount uint32
		vk.GetPhysicalDeviceSurfacePresentModes(device, d.surface, &presentModeCount, nil)
		if presentModeCount == 0 {
			continue
		}
		details.presentModes = make([]vk.PresentMode, presentModeCount)
		vk.GetPhysicalDeviceSurfacePresentModes(device, d.surface, &presentModeCount, details.presentModes)
		deviceSelected = true
		physicalDevice = device
		d.gpu = device
	}
	if !deviceSelected {
		return errors.New("failed to find a sutible GPU")
	}
	qi := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: d.graphicsIdx,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	if d.graphicsIdx != d.presentIdx {
		qi = append(qi, vk.DeviceQueueCreateInfo{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: d.presentIdx,
			QueueCount:       1,
			PQueuePriorities: []float32{1.0},
		})
	}
	ret := vk.CreateDevice(physicalDevice, &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(qi)),
		PQueueCreateInfos:       qi,
		EnabledExtensionCount:   uint32(len(wantedExtensions)),
		PpEnabledExtensionNames: safeStrings(wantedExtensions),
	}, nil, &d.device)
	if ret != vk.Success {
		return errors.New("unable to create logical device")
	}
	vk.GetDeviceQueue(d.device, d.graphicsIdx, 0, &d.graphicsQueue)
	vk.GetDeviceQueue(d.device, d.presentIdx, 0, &d.presentQueue)
	return nil
}

func (d *VkDevice) createSwapChain() error {
	surfaceFormat := d.chooseSwapSurfaceFormat()
	surfaceFormat.Deref()
	presentMode := d.chooseSwapPresentMode()
	extent := d.chooseSwapExtent()
	extent.Deref()
	details.capabilities.Deref()
	imageCount := details.capabilities.MinImageCount + 1
	if details.capabilities.MaxImageCount > 0 {
		if imageCount > details.capabilities.MaxImageCount {
			imageCount = details.capabilities.MaxImageCount
		}
	}
	createInfo := vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          d.surface,
		MinImageCount:    imageCount,
		ImageFormat:      surfaceFormat.Format,
		ImageColorSpace:  surfaceFormat.ColorSpace,
		ImageExtent:      extent,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
	}
	if d.graphicsIdx != d.presentIdx {
		createInfo.ImageSharingMode = vk.SharingModeConcurrent
		createInfo.QueueFamilyIndexCount = 2
		createInfo.PQueueFamilyIndices = []uint32{d.graphicsIdx, d.presentIdx}
	} else {
		createInfo.ImageSharingMode = vk.SharingModeExclusive
		createInfo.QueueFamilyIndexCount = 0
		createInfo.PQueueFamilyIndices = []uint32{}
	}
	createInfo.PreTransform = details.capabilities.CurrentTransform
	createInfo.CompositeAlpha = vk.CompositeAlphaOpaqueBit
	createInfo.PresentMode = presentMode
	createInfo.Clipped = vk.True
	createInfo.OldSwapchain = vk.Swapchain(vk.NullHandle)
	var swapchain vk.Swapchain
	if res := vk.CreateSwapchain(d.device, &createInfo, nil, &swapchain); res != vk.Success {
		return errors.New("failed to create swap chain")
	}
	d.swapChain = swapchain
	var numImgs uint32
	vk.GetSwapchainImages(d.device, d.swapChain, &numImgs, nil)
	d.images = make([]vk.Image, numImgs)
	if res := vk.GetSwapchainImages(d.device, d.swapChain, &numImgs, d.images); res != vk.Success {
		return errors.New("failed to get swap chain images")
	}
	d.swapChainImageFormat = surfaceFormat.Format
	d.swapChainExtent = extent
	return nil
}

func (d *VkDevice) chooseSwapSurfaceFormat() vk.SurfaceFormat {
	if len(details.formats) == 1 {
		details.formats[0].Deref()
		if details.formats[0].Format == vk.FormatUndefined {
			return vk.SurfaceFormat{
				Format:     vk.FormatB8g8r8Unorm,
				ColorSpace: vk.ColorSpaceSrgbNonlinear,
			}
		}
	}
	for _, f := range details.formats {
		f.Deref()
		if f.Format == vk.FormatB8g8r8Unorm && f.ColorSpace == vk.ColorSpaceSrgbNonlinear {
			return f
		}
	}
	return details.formats[0]
}

func (d *VkDevice) chooseSwapPresentMode() vk.PresentMode {
	bestMode := vk.PresentModeFifo
	for _, p := range details.presentModes {
		if p == vk.PresentModeMailbox {
			return p
		}
		if p == vk.PresentModeImmediate {
			bestMode = p
		}
	}
	return bestMode
}

func (d *VkDevice) chooseSwapExtent() vk.Extent2D {
	details.capabilities.Deref()
	if details.capabilities.CurrentExtent.Width != vk.MaxUint32 {
		return details.capabilities.CurrentExtent
	}
	w := uint32(CanvasWidth())
	h := uint32(CanvasHeight())
	actualExtent := vk.Extent2D{
		Width:  w,
		Height: h,
	}
	actualExtent.Width = clamp(details.capabilities.MaxImageExtent.Width,
		details.capabilities.MinImageExtent.Width, actualExtent.Width)
	actualExtent.Height = clamp(details.capabilities.MaxImageExtent.Height,
		details.capabilities.MinImageExtent.Height, actualExtent.Height)
	return actualExtent
}

func (d *VkDevice) createImageViews() error {
	d.swapChainImageViews = make([]vk.ImageView, len(d.images))
	for i, image := range d.images {
		createInfo := vk.ImageViewCreateInfo{
			SType:    vk.StructureTypeImageViewCreateInfo,
			Image:    image,
			ViewType: vk.ImageViewType2d,
			Format:   d.swapChainImageFormat,
		}
		createInfo.Components.R = vk.ComponentSwizzleIdentity
		createInfo.Components.G = vk.ComponentSwizzleIdentity
		createInfo.Components.B = vk.ComponentSwizzleIdentity
		createInfo.Components.A = vk.ComponentSwizzleIdentity
		createInfo.SubresourceRange.BaseMipLevel = 0
		createInfo.SubresourceRange.LevelCount = 1
		createInfo.SubresourceRange.BaseArrayLayer = 0
		createInfo.SubresourceRange.LayerCount = 1
		if res := vk.CreateImageView(d.device, &createInfo, nil, &d.swapChainImageViews[i]); res != vk.Success {
			return errors.New("unable to create image view from swap chain images")
		}
	}
	return nil
}
