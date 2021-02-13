//+build vulkan

package common

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/gl"
	vk "github.com/vulkan-go/vulkan"
)

type basicShader struct {
	BatchSize int

	indices  []uint16
	vertices []float32

	inPosition  int
	inTexCoords int
	inColor     int

	projectionMatrix *engo.Matrix
	viewMatrix       *engo.Matrix
	projViewMatrix   *engo.Matrix
	modelMatrix      *engo.Matrix
	cullingMatrix    *engo.Matrix

	projViewChange bool

	camera        *CameraSystem
	cameraEnabled bool

	idx                      int
	renderPass               vk.RenderPass
	pipelineLayout           vk.PipelineLayout
	graphicsPipelines        []vk.Pipeline
	swapChainFramebuffers    []vk.Framebuffer
	commandPool              vk.CommandPool
	commandBuffers           []vk.CommandBuffer
	imageAvailableSemaphores []vk.Semaphore
	renderFinishedSemaphores []vk.Semaphore
	inFlightFences           []vk.Fence
	currentFrame             int
	framebufferResized       bool
	lock                     sync.Mutex
	vertexBuffer             vk.Buffer
	vertexBufferMemory       vk.DeviceMemory
	indexBuffer              vk.Buffer
	indexBufferMemory        vk.DeviceMemory
	descriptorSetLayouts     []vk.DescriptorSetLayout
	uniformBuffers           []vk.Buffer
	uniformBuffersMemory     []vk.DeviceMemory
	startTime                time.Time
	descriptorPool           vk.DescriptorPool
	descriptorSets           []vk.DescriptorSet
}

func (s *basicShader) Setup(w *ecs.World) error {
	if s.BatchSize > MaxSprites {
		return fmt.Errorf("%d is greater than the maximum batch size of %d", s.BatchSize, MaxSprites)
	}
	if s.BatchSize <= 0 {
		s.BatchSize = MaxSprites
	}
	// Create the vertex buffer for batching.
	s.vertices = make([]float32, s.BatchSize*spriteSize)
	// Create and populate indices buffer. The size of the buffer depends on the batch size.
	// These should never change, so we can just initialize them once here and be done with it.
	numIndicies := s.BatchSize * 6
	s.indices = make([]uint16, numIndicies)
	for i, j := 0, 0; i < numIndicies; i, j = i+6, j+4 {
		s.indices[i+0] = uint16(j + 0)
		s.indices[i+1] = uint16(j + 1)
		s.indices[i+2] = uint16(j + 2)
		s.indices[i+3] = uint16(j + 0)
		s.indices[i+4] = uint16(j + 2)
		s.indices[i+5] = uint16(j + 3)
	}

	engo.Mailbox.Listen("WindowResizeMessage", func(m engo.Message) {
		_, ok := m.(engo.WindowResizeMessage)
		if !ok {
			return
		}
		s.lock.Lock()
		s.framebufferResized = true
		s.lock.Unlock()
	})

	if err := s.createRenderPass(); err != nil {
		return err
	}
	if err := s.createDescriptorSetLayout(); err != nil {
		return err
	}
	if err := s.createGraphicsPipeline(); err != nil {
		return err
	}
	if err := s.createFrameBuffers(); err != nil {
		return err
	}
	if err := s.createCommandPool(); err != nil {
		return err
	}
	if err := r.createTextureImage(); err != nil {
		return err
	}
	if err := r.createTextureImageView(); err != nil {
		return err
	}
	if err := r.createTextureSampler(); err != nil {
		return err
	}
	if err := r.createVertexBuffer(); err != nil {
		return err
	}
	if err := r.createIndexBuffer(); err != nil {
		return err
	}
	if err := r.createUniformBuffers(); err != nil {
		return err
	}
	if err := r.createDescriptorPool(); err != nil {
		return err
	}
	if err := r.createDescriptorSets(); err != nil {
		return err
	}
	if err := r.createCommandBuffers(); err != nil {
		return err
	}
	if err := r.createSyncObjects(); err != nil {
		return err
	}

	s.projectionMatrix = engo.IdentityMatrix()
	s.viewMatrix = engo.IdentityMatrix()
	s.projViewMatrix = engo.IdentityMatrix()
	s.modelMatrix = engo.IdentityMatrix()
	s.cullingMatrix = engo.IdentityMatrix()

	s.setTexture(nil)

	return nil
}

func (s *basicShader) createRenderPass() error {
	colorAttachment := vk.AttachmentDescription{
		Format:         s.swapChainImageFormat,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutPresentSrc,
	}

	colorAttachmentRef := vk.AttachmentReference{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}

	subpass := vk.SubpassDescription{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    []vk.AttachmentReference{colorAttachmentRef},
	}

	dependency := vk.SubpassDependency{
		SrcSubpass:    vk.SubpassExternal,
		DstSubpass:    0,
		SrcStageMask:  vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		SrcAccessMask: 0,
		DstStageMask:  vk.PipelineStageFlags(vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit),
	}

	renderPassInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    []vk.AttachmentDescription{colorAttachment},
		SubpassCount:    1,
		PSubpasses:      []vk.SubpassDescription{subpass},
		DependencyCount: 1,
		PDependencies:   []vk.SubpassDependency{dependency},
	}

	var renderPass vk.RenderPass
	if res := vk.CreateRenderPass(engo.Device.Device(), &renderPassInfo, nil, &renderPass); res != vk.Success {
		return errors.New("failed to create render pass")
	}
	s.renderPass = renderPass

	return nil
}

func (s *basicShader) createDescriptorSetLayout() error {
	uboLayoutBinding := vk.DescriptorSetLayoutBinding{
		Binding:            0,
		DescriptorType:     vk.DescriptorTypeUniformBuffer,
		DescriptorCount:    1,
		StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		PImmutableSamplers: []vk.Sampler{vk.NullSampler},
	}
	samplerLayoutBinding := vk.DescriptorSetLayoutBinding{
		Binding:         1,
		DescriptorCount: 1,
		DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
		StageFlags:      vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
	}
	bindings := []vk.DescriptorSetLayoutBinding{uboLayoutBinding, samplerLayoutBinding}

	layoutInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: uint32(len(bindings)),
		PBindings:    bindings,
	}

	var descriptorSetLayout vk.DescriptorSetLayout
	if res := vk.CreateDescriptorSetLayout(engo.Device.Device(), &layoutInfo, nil, &descriptorSetLayout); res != vk.Success {
		return errors.New("unable to create descriptor set layout")
	}
	s.descriptorSetLayouts = append(s.descriptorSetLayouts, descriptorSetLayout)

	return nil
}

func (s *basicShader) createGraphicsPipeline() error {
	vertShaderData, err := shaders.Asset("default/vert.spv")
	if err != nil {
		return err
	}
	fragShaderData, err := shaders.Asset("default/frag.spv")
	if err != nil {
		return err
	}
	vertShaderModule, err := LoadShaderModule(vertShaderData)
	if err != nil {
		return err
	}
	fragShaderModule, err := LoadShaderModule(fragShaderData)
	if err != nil {
		return err
	}

	vertShaderStageInfo := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageVertexBit,
		Module: vertShaderModule,
		PName:  SafeString("main"),
	}

	fragShaderStageInfo := vk.PipelineShaderStageCreateInfo{
		SType:  vk.StructureTypePipelineShaderStageCreateInfo,
		Stage:  vk.ShaderStageFragmentBit,
		Module: fragShaderModule,
		PName:  SafeString("main"),
	}

	shaderStages := []vk.PipelineShaderStageCreateInfo{
		vertShaderStageInfo,
		fragShaderStageInfo,
	}

	var a []vk.VertexInputAttributeDescription
	a = append(a, vk.VertexInputAttributeDescription{
		Binding:  0,
		Location: 0,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   0,
	})
	a = append(a, vk.VertexInputAttributeDescription{
		Binding:  0,
		Location: 1,
		Format:   vk.FormatR32g32b32Sfloat,
		Offset:   2 * 4,
	})
	a = append(a, vk.VertexInputAttributeDescription{
		Binding:  0,
		Location: 2,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   5 * 4,
	})
	b := vk.VertexInputBindingDescription{
		Binding:   0,
		Stride:    7 * 4,
		InputRate: vk.VertexInputRateVertex,
	}

	vertexInputInfo := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   1,
		VertexAttributeDescriptionCount: uint32(len(a)),
		PVertexBindingDescriptions:      []vk.VertexInputBindingDescription{b},
		PVertexAttributeDescriptions:    a,
	}

	inputAssembly := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyTriangleList,
		PrimitiveRestartEnable: vk.False,
	}

	viewport := vk.Viewport{
		X:        0,
		Y:        0,
		Width:    float32(engo.Device.SwapChainExtent().Width),
		Height:   float32(engo.Device.SwapChainExtent().Height),
		MinDepth: 0,
		MaxDepth: 1,
	}

	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: engo.Device.SwapChainExtent(),
	}

	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		PViewports:    []vk.Viewport{viewport},
		ScissorCount:  1,
		PScissors:     []vk.Rect2D{scissor},
	}

	rasterizer := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		PolygonMode:             vk.PolygonModeFill,
		LineWidth:               1,
		CullMode:                vk.CullModeFlags(vk.CullModeBackBit),
		FrontFace:               vk.FrontFaceCounterClockwise,
		DepthBiasEnable:         vk.False,
	}

	multisampling := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		SampleShadingEnable:   vk.False,
		RasterizationSamples:  vk.SampleCount1Bit,
		MinSampleShading:      1,
		AlphaToCoverageEnable: vk.False,
		AlphaToOneEnable:      vk.False,
	}

	colorBlendAttachment := vk.PipelineColorBlendAttachmentState{
		ColorWriteMask:      vk.ColorComponentFlags(vk.ColorComponentRBit | vk.ColorComponentGBit | vk.ColorComponentBBit | vk.ColorComponentABit),
		BlendEnable:         vk.False,
		SrcColorBlendFactor: vk.BlendFactorOne,
		DstColorBlendFactor: vk.BlendFactorZero,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	}

	colorBlending := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{colorBlendAttachment},
	}

	pipelineLayoutInfo := vk.PipelineLayoutCreateInfo{
		SType:          vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount: 1,
		PSetLayouts:    s.descriptorSetLayouts,
	}
	var pipelineLayout vk.PipelineLayout
	if res := vk.CreatePipelineLayout(engo.Device.Device(), &pipelineLayoutInfo, nil, &pipelineLayout); res != vk.Success {
		return errors.New("failed to create pipeline layout")
	}
	r.pipelineLayout = pipelineLayout

	pipelineInfo := vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          2,
		PStages:             shaderStages,
		PVertexInputState:   &vertexInputInfo,
		PInputAssemblyState: &inputAssembly,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterizer,
		PMultisampleState:   &multisampling,
		PColorBlendState:    &colorBlending,
		Layout:              s.pipelineLayout,
		RenderPass:          s.renderPass,
		Subpass:             0,
	}

	s.graphicsPipelines = make([]vk.Pipeline, 1)
	if res := vk.CreateGraphicsPipelines(engo.Device.Device(), nil, 1, []vk.GraphicsPipelineCreateInfo{pipelineInfo}, nil, s.graphicsPipelines); res != vk.Success {
		errors.New("failed to create graphics pipeline")
	}

	vk.DestroyShaderModule(engo.Device.Device(), vertShaderModule, nil)
	vk.DestroyShaderModule(engo.Device.Device(), fragShaderModule, nil)

	return nil
}

func (s *basicShader) createFrameBuffers() error {
	s.swapChainFramebuffers = make([]vk.Framebuffer, len(s.swapChainImageViews))

	for idx, view := range r.swapChainImageViews {
		attachments := []vk.ImageView{view}

		framebufferInfo := vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      s.renderPass,
			AttachmentCount: 1,
			PAttachments:    attachments,
			Width:           engo.Device.SwapChainExtent().Width,
			Height:          engo.Device.SwapChainExtent().Height,
			Layers:          1,
		}

		if res := vk.CreateFramebuffer(engo.Device.Device(), &framebufferInfo, nil, &s.swapChainFramebuffers[idx]); res != vk.Success {
			return errors.New("failed to create framebuffer")
		}
	}

	return nil
}

func (s *basicShader) createCommandPool() error {
	poolInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		QueueFamilyIndex: engo.Device.GraphicsQueueIndex(),
	}

	var commandPool vk.CommandPool
	if res := vk.CreateCommandPool(engo.Device.Device(), &poolInfo, nil, &commandPool); res != vk.Success {
		return errors.New("failed to create command pool")
	}
	s.commandPool = commandPool

	return nil
}

func (s *basicShader) Pre() {
	engo.Gl.Enable(engo.Gl.BLEND)
	engo.Gl.BlendFunc(engo.Gl.SRC_ALPHA, engo.Gl.ONE_MINUS_SRC_ALPHA)
	// Enable shader and buffer, enable attributes in shader
	engo.Gl.UseProgram(s.program)
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, s.indexBuffer)
	engo.Gl.EnableVertexAttribArray(s.inPosition)
	engo.Gl.EnableVertexAttribArray(s.inTexCoords)
	engo.Gl.EnableVertexAttribArray(s.inColor)

	// The matrixProjView shader uniform is projection * view.
	// We do the multiplication on the CPU instead of sending each matrix to the shader and letting the GPU do the multiplication,
	// because it's likely faster to do the multiplication client side and send the result over the shader bus than to send two separate
	// buffers over the bus and then do the multiplication on the GPU.
	if s.projViewChange {
		s.projViewMatrix = s.projectionMatrix.Multiply(s.viewMatrix)
		s.projViewChange = false
	}
	engo.Gl.UniformMatrix3fv(s.matrixProjView, false, s.projViewMatrix.Val[:])

	// Since we are batching client side, we only have one VBO, so we can just bind it now and use it for the entire frame.
	engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, s.vertexBuffer)
	engo.Gl.VertexAttribPointer(s.inPosition, 2, engo.Gl.FLOAT, false, 20, 0)
	engo.Gl.VertexAttribPointer(s.inTexCoords, 2, engo.Gl.FLOAT, false, 20, 8)
	engo.Gl.VertexAttribPointer(s.inColor, 4, engo.Gl.UNSIGNED_BYTE, true, 20, 16)
}

func (s *basicShader) PrepareCulling() {
	s.projViewChange = true
	// (Re)initialize the projection matrix.
	s.projectionMatrix.Identity()
	if engo.ScaleOnResize() {
		s.projectionMatrix.Scale(1/(engo.GameWidth()/2), 1/(-engo.GameHeight()/2))
	} else {
		s.projectionMatrix.Scale(1/(engo.CanvasWidth()/(2*engo.CanvasScale())), 1/(-engo.CanvasHeight()/(2*engo.CanvasScale())))
	}
	// (Re)initialize the view matrix
	s.viewMatrix.Identity()
	if s.cameraEnabled {
		s.viewMatrix.Scale(1/s.camera.z, 1/s.camera.z)
		s.viewMatrix.Translate(-s.camera.x, -s.camera.y).Rotate(s.camera.angle)
	} else {
		scaleX, scaleY := s.projectionMatrix.ScaleComponent()
		s.viewMatrix.Translate(-1/scaleX, 1/scaleY)
	}
	s.cullingMatrix.Identity()
	s.cullingMatrix.Multiply(s.projectionMatrix).Multiply(s.viewMatrix)
	s.cullingMatrix.Scale(engo.GetGlobalScale().X, engo.GetGlobalScale().Y)
}

func (s *basicShader) ShouldDraw(rc *RenderComponent, sc *SpaceComponent) bool {
	tsc := SpaceComponent{
		Position: sc.Position,
		Width:    rc.Drawable.Width() * rc.Scale.X,
		Height:   rc.Drawable.Height() * rc.Scale.Y,
		Rotation: sc.Rotation,
	}

	c := tsc.Corners()
	c[0].MultiplyMatrixVector(s.cullingMatrix)
	c[1].MultiplyMatrixVector(s.cullingMatrix)
	c[2].MultiplyMatrixVector(s.cullingMatrix)
	c[3].MultiplyMatrixVector(s.cullingMatrix)

	return !((c[0].X < -1 && c[1].X < -1 && c[2].X < -1 && c[3].X < -1) || // All points left of the "viewport"
		(c[0].X > 1 && c[1].X > 1 && c[2].X > 1 && c[3].X > 1) || // All points right of the "viewport"
		(c[0].Y < -1 && c[1].Y < -1 && c[2].Y < -1 && c[3].Y < -1) || // All points above of the "viewport"
		(c[0].Y > 1 && c[1].Y > 1 && c[2].Y > 1 && c[3].Y > 1)) // All points below of the "viewport"
}

func (s *basicShader) Draw(ren *RenderComponent, space *SpaceComponent) {
	// If our texture (or any of its properties) has changed or we've reached the end of our buffer, flush before moving on.
	if s.lastTexture != ren.Drawable.Texture() {
		s.flush()
		engo.Gl.BindTexture(engo.Gl.TEXTURE_2D, ren.Drawable.Texture())
		s.setTexture(ren.Drawable.Texture())
	} else if s.idx == len(s.vertices) {
		s.flush()
	}

	if s.lastRepeating != ren.Repeat {
		s.flush()
		var val int
		switch ren.Repeat {
		case NoRepeat:
			val = engo.Gl.CLAMP_TO_EDGE
		case ClampToEdge:
			val = engo.Gl.CLAMP_TO_EDGE
		case ClampToBorder:
			val = engo.Gl.CLAMP_TO_EDGE
		case Repeat:
			val = engo.Gl.REPEAT
		case MirroredRepeat:
			val = engo.Gl.MIRRORED_REPEAT
		}
		engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_WRAP_S, val)
		engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_WRAP_T, val)

		s.lastRepeating = ren.Repeat
	}

	if s.lastMagFilter != ren.magFilter {
		s.flush()
		var val int
		switch ren.magFilter {
		case FilterNearest:
			val = engo.Gl.NEAREST
		case FilterLinear:
			val = engo.Gl.LINEAR
		}
		engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_MAG_FILTER, val)

		s.lastMagFilter = ren.magFilter
	}

	if s.lastMinFilter != ren.minFilter {
		s.flush()
		var val int
		switch ren.minFilter {
		case FilterNearest:
			val = engo.Gl.NEAREST
		case FilterLinear:
			val = engo.Gl.LINEAR
		}
		engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_MIN_FILTER, val)

		s.lastMinFilter = ren.minFilter
	}

	// Update the vertex buffer data.
	s.updateBuffer(ren, space)
	s.idx += 20
}

func (s *basicShader) Post() {
	s.flush()
	s.setTexture(nil)

	// Cleanup
	engo.Gl.DisableVertexAttribArray(s.inPosition)
	engo.Gl.DisableVertexAttribArray(s.inTexCoords)
	engo.Gl.DisableVertexAttribArray(s.inColor)

	engo.Gl.BindTexture(engo.Gl.TEXTURE_2D, nil)
	engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, nil)
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, nil)

	engo.Gl.Disable(engo.Gl.BLEND)
}

// setTexture resets all last* values from basicShader to a new default value (255)
func (s *basicShader) setTexture(texture *gl.Texture) {
	s.lastTexture = texture
	s.lastMinFilter = 255
	s.lastMagFilter = 255
	s.lastRepeating = 255
}

func (s *basicShader) flush() {
	// If we haven't rendered anything yet, no point in flushing.
	if s.idx == 0 {
		return
	}
	engo.Gl.BufferData(engo.Gl.ARRAY_BUFFER, s.vertices, engo.Gl.STATIC_DRAW)
	// We only want to draw the indicies up to the number of sprites in the current batch.
	count := s.idx / 20 * 6
	engo.Gl.DrawElements(engo.Gl.TRIANGLES, count, engo.Gl.UNSIGNED_SHORT, 0)
	s.idx = 0
	// We need to reset the vertex buffer so that when we start drawing again, we don't accidentally use junk data.
	// The "simpler" way to do this would be to just create a new slice with make(), however that would cause the
	// previous slice to be marked for garbage collection and we'd prefer to keep the GC activity to a minimum.
	for i := range s.vertices {
		s.vertices[i] = 0
	}
}

func (s *basicShader) updateBuffer(ren *RenderComponent, space *SpaceComponent) {
	// For backwards compatibility, ren.Buffer is set to the VBO and ren.BufferContent
	// is set to the slice of the vertex buffer for the current sprite. This same slice is
	// populated with vertex data via generateBufferContent.
	ren.BufferData.Buffer = s.vertexBuffer
	ren.BufferData.BufferContent = s.vertices[s.idx : s.idx+20]
	s.generateBufferContent(ren, space, ren.BufferData.BufferContent)
}

func (s *basicShader) makeModelMatrix(ren *RenderComponent, space *SpaceComponent) *engo.Matrix {
	// Instead of creating a new model matrix every time, we instead store a global one as a struct member
	// and just reset it for every sprite. This prevents us from allocating a bunch of new Matrix instances in memory
	// ultimately saving on GC activity.
	s.modelMatrix.Identity().Scale(engo.GetGlobalScale().X, engo.GetGlobalScale().Y).Translate(space.Position.X, space.Position.Y)
	if space.Rotation != 0 {
		s.modelMatrix.Rotate(space.Rotation)
	}
	s.modelMatrix.Scale(ren.Scale.X, ren.Scale.Y)
	return s.modelMatrix
}

func (s *basicShader) generateBufferContent(ren *RenderComponent, space *SpaceComponent, buffer []float32) bool {
	// We shouldn't use SpaceComponent to get width/height, because this usually already contains the Scale (which
	// is being added elsewhere, so we don't want to over-do it)
	w := ren.Drawable.Width()
	h := ren.Drawable.Height()

	tint := colorToFloat32(ren.Color)

	u, v, u2, v2 := ren.Drawable.View()

	if ren.Repeat != NoRepeat {
		u2 = space.Width / (ren.Drawable.Width() * ren.Scale.X)
		w *= u2
		v2 = space.Width / (ren.Drawable.Height() * ren.Scale.Y)
		h *= v2
	}

	var changed bool

	//setBufferValue(buffer, 0, 0, &changed)
	//setBufferValue(buffer, 1, 0, &changed)
	setBufferValue(buffer, 2, u, &changed)
	setBufferValue(buffer, 3, v, &changed)
	setBufferValue(buffer, 4, tint, &changed)

	setBufferValue(buffer, 5, w, &changed)
	//setBufferValue(buffer, 6, 0, &changed)
	setBufferValue(buffer, 7, u2, &changed)
	setBufferValue(buffer, 8, v, &changed)
	setBufferValue(buffer, 9, tint, &changed)

	setBufferValue(buffer, 10, w, &changed)
	setBufferValue(buffer, 11, h, &changed)
	setBufferValue(buffer, 12, u2, &changed)
	setBufferValue(buffer, 13, v2, &changed)
	setBufferValue(buffer, 14, tint, &changed)

	//setBufferValue(buffer, 15, 0, &changed)
	setBufferValue(buffer, 16, h, &changed)
	setBufferValue(buffer, 17, u, &changed)
	setBufferValue(buffer, 18, v2, &changed)
	setBufferValue(buffer, 19, tint, &changed)

	// Since each sprite in the batch has a different transform, we can't just send the model matrix into
	// the shader and let the GPU take care of it. Instead, we need to multiply the current sprite's model matrix
	// with the position component for each vertex of the current sprite on the CPU, and send the transformed
	// positions to the shader directly.
	modelMatrix := s.makeModelMatrix(ren, space)
	s.multModel(modelMatrix, buffer[:2])
	s.multModel(modelMatrix, buffer[5:7])
	s.multModel(modelMatrix, buffer[10:12])
	s.multModel(modelMatrix, buffer[15:17])
	return changed
}

func (s *basicShader) multModel(m *engo.Matrix, v []float32) {
	tmp := engo.MultiplyMatrixVector(m, v)
	v[0] = tmp[0]
	v[1] = tmp[1]
}

func (s *basicShader) SetCamera(c *CameraSystem) {
	s.projViewChange = true
	if s.cameraEnabled {
		s.camera = c
		s.viewMatrix.Identity().Translate(-s.camera.x, -s.camera.y).Rotate(s.camera.angle)
	} else {
		scaleX, scaleY := s.projectionMatrix.ScaleComponent()
		s.viewMatrix.Translate(-1/scaleX, 1/scaleY)
	}
}
