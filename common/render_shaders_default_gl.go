//+build !vulkan

package common

import (
	"fmt"
	"runtime"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/gl"
)

const (
	defaultVertexShader = `
	attribute vec2 in_Position;
	attribute vec2 in_TexCoords;
	attribute vec4 in_Color;

	uniform mat3 matrixProjView;

	varying vec4 var_Color;
	varying vec2 var_TexCoords;

	void main() {
	  var_Color = in_Color;
	  var_TexCoords = in_TexCoords;

	  vec3 matr = matrixProjView * vec3(in_Position, 1.0);
	  gl_Position = vec4(matr.xy, 0, matr.z);
	}
`

	defaultFragmentShader = `
	#ifdef GL_ES
	#define LOWP lowp
	precision mediump float;
	#else
	#define LOWP
	#endif

	varying vec4 var_Color;
	varying vec2 var_TexCoords;

	uniform sampler2D uf_Texture;

	void main (void) {
	  gl_FragColor = var_Color * texture2D(uf_Texture, var_TexCoords);
	}
`
)

type basicShader struct {
	BatchSize int

	indices     []uint16
	indexBuffer *gl.Buffer
	program     *gl.Program

	vertices                     []float32
	vertexBuffer                 *gl.Buffer
	lastTexture                  *gl.Texture
	lastRepeating                TextureRepeating
	lastMagFilter, lastMinFilter ZoomFilter

	inPosition  int
	inTexCoords int
	inColor     int

	matrixProjView *gl.UniformLocation

	projectionMatrix *engo.Matrix
	viewMatrix       *engo.Matrix
	projViewMatrix   *engo.Matrix
	modelMatrix      *engo.Matrix
	cullingMatrix    *engo.Matrix

	projViewChange bool

	camera        *CameraSystem
	cameraEnabled bool

	idx int
}

func (s *basicShader) Setup(w *ecs.World) error {
	if s.BatchSize > MaxSprites {
		return fmt.Errorf("%d is greater than the maximum batch size of %d", s.BatchSize, MaxSprites)
	}

	if s.BatchSize <= 0 {
		if runtime.GOOS == "js" {
			s.BatchSize = 2048 // js can't seem to handle the whole buffer size
		} else {
			s.BatchSize = MaxSprites
		}
	}

	// Create the vertex buffer for batching.
	s.vertices = make([]float32, s.BatchSize*spriteSize)
	s.vertexBuffer = engo.Gl.CreateBuffer()
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
	var err error
	s.program, err = LoadShader(defaultVertexShader, defaultFragmentShader)
	if err != nil {
		return err
	}
	s.indexBuffer = engo.Gl.CreateBuffer()
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, s.indexBuffer)
	engo.Gl.BufferData(engo.Gl.ELEMENT_ARRAY_BUFFER, s.indices, engo.Gl.STATIC_DRAW)

	s.inPosition = engo.Gl.GetAttribLocation(s.program, "in_Position")
	s.inTexCoords = engo.Gl.GetAttribLocation(s.program, "in_TexCoords")
	s.inColor = engo.Gl.GetAttribLocation(s.program, "in_Color")

	s.matrixProjView = engo.Gl.GetUniformLocation(s.program, "matrixProjView")

	s.projectionMatrix = engo.IdentityMatrix()
	s.viewMatrix = engo.IdentityMatrix()
	s.projViewMatrix = engo.IdentityMatrix()
	s.modelMatrix = engo.IdentityMatrix()
	s.cullingMatrix = engo.IdentityMatrix()

	s.setTexture(nil)

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
