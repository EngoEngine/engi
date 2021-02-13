//+build !vulkan

package common

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/math"
	"github.com/EngoEngine/gl"
)

type textShader struct {
	program *gl.Program

	indicesRectangles    []uint16
	indicesRectanglesVBO *gl.Buffer

	inPosition  int
	inTexCoords int
	inColor     int

	matrixProjection *gl.UniformLocation
	matrixView       *gl.UniformLocation
	matrixModel      *gl.UniformLocation

	projectionMatrix []float32
	viewMatrix       []float32
	modelMatrix      []float32

	camera        *CameraSystem
	cameraEnabled bool

	lastBuffer  *gl.Buffer
	lastTexture *gl.Texture
}

func (l *textShader) Setup(w *ecs.World) error {
	var err error
	l.program, err = LoadShader(`
attribute vec2 in_Position;
attribute vec2 in_TexCoords;
attribute vec4 in_Color;

uniform mat3 matrixProjection;
uniform mat3 matrixView;
uniform mat3 matrixModel;

varying vec4 var_Color;
varying vec2 var_TexCoords;

void main() {
  var_Color = in_Color;
  var_TexCoords = in_TexCoords;

  vec3 matr = matrixProjection * matrixView * matrixModel * vec3(in_Position, 1.0);
  gl_Position = vec4(matr.xy, 0, matr.z);
}
`, `
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
}`)

	if err != nil {
		return err
	}

	// Create and populate indices buffer
	l.indicesRectangles = make([]uint16, 6*bufferSize)
	for i, j := 0, 0; i < bufferSize*6; i, j = i+6, j+4 {
		l.indicesRectangles[i+0] = uint16(j + 0)
		l.indicesRectangles[i+1] = uint16(j + 1)
		l.indicesRectangles[i+2] = uint16(j + 2)
		l.indicesRectangles[i+3] = uint16(j + 0)
		l.indicesRectangles[i+4] = uint16(j + 2)
		l.indicesRectangles[i+5] = uint16(j + 3)
	}
	l.indicesRectanglesVBO = engo.Gl.CreateBuffer()
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, l.indicesRectanglesVBO)
	engo.Gl.BufferData(engo.Gl.ELEMENT_ARRAY_BUFFER, l.indicesRectangles, engo.Gl.STATIC_DRAW)

	// Define things that should be read from the texture buffer
	l.inPosition = engo.Gl.GetAttribLocation(l.program, "in_Position")
	l.inTexCoords = engo.Gl.GetAttribLocation(l.program, "in_TexCoords")
	l.inColor = engo.Gl.GetAttribLocation(l.program, "in_Color")

	// Define things that should be set per draw
	l.matrixProjection = engo.Gl.GetUniformLocation(l.program, "matrixProjection")
	l.matrixView = engo.Gl.GetUniformLocation(l.program, "matrixView")
	l.matrixModel = engo.Gl.GetUniformLocation(l.program, "matrixModel")

	l.projectionMatrix = make([]float32, 9)
	l.projectionMatrix[8] = 1

	l.viewMatrix = make([]float32, 9)
	l.viewMatrix[0] = 1
	l.viewMatrix[4] = 1
	l.viewMatrix[8] = 1

	l.modelMatrix = make([]float32, 9)
	l.modelMatrix[0] = 1
	l.modelMatrix[4] = 1
	l.modelMatrix[8] = 1

	return nil
}

func (l *textShader) Pre() {
	engo.Gl.Enable(engo.Gl.BLEND)
	engo.Gl.BlendFunc(engo.Gl.SRC_ALPHA, engo.Gl.ONE_MINUS_SRC_ALPHA)

	// Bind shader and buffer, enable attributes
	engo.Gl.UseProgram(l.program)
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, l.indicesRectanglesVBO)
	engo.Gl.EnableVertexAttribArray(l.inPosition)
	engo.Gl.EnableVertexAttribArray(l.inTexCoords)
	engo.Gl.EnableVertexAttribArray(l.inColor)

	if engo.ScaleOnResize() {
		l.projectionMatrix[0] = 1 / (engo.GameWidth() / 2)
		l.projectionMatrix[4] = 1 / (-engo.GameHeight() / 2)
	} else {
		l.projectionMatrix[0] = 1 / (engo.CanvasWidth() / (2 * engo.CanvasScale()))
		l.projectionMatrix[4] = 1 / (-engo.CanvasHeight() / (2 * engo.CanvasScale()))
	}

	if l.cameraEnabled {
		l.viewMatrix[1], l.viewMatrix[0] = math.Sincos(l.camera.angle * math.Pi / 180)
		l.viewMatrix[3] = -l.viewMatrix[1]
		l.viewMatrix[4] = l.viewMatrix[0]
		l.viewMatrix[6] = -l.camera.x
		l.viewMatrix[7] = -l.camera.y
		l.viewMatrix[8] = l.camera.z
	} else {
		l.viewMatrix[6] = -1 / l.projectionMatrix[0]
		l.viewMatrix[7] = 1 / l.projectionMatrix[4]
	}

	engo.Gl.UniformMatrix3fv(l.matrixProjection, false, l.projectionMatrix)
	engo.Gl.UniformMatrix3fv(l.matrixView, false, l.viewMatrix)
}

func (l *textShader) updateBuffer(ren *RenderComponent, space *SpaceComponent) {
	txt, ok := ren.Drawable.(Text)
	if !ok {
		unsupportedType(ren.Drawable)
		return
	}

	if len(ren.BufferData.BufferContent) < 20*len(txt.Text) {
		ren.BufferData.BufferContent = make([]float32, 20*len(txt.Text)) // TODO: update this to actual value?
	}

	// Reset buffer so artifacts don't occur when txt.Text changes
	for i := 0; i < len(ren.BufferContent); i++ {
		ren.BufferContent[i] = 0
	}
	if changed := l.generateBufferContent(ren, space, ren.BufferContent); !changed {
		return
	}

	if ren.BufferData.Buffer == nil {
		ren.BufferData.Buffer = engo.Gl.CreateBuffer()
	}
	engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, ren.BufferData.Buffer)
	engo.Gl.BufferData(engo.Gl.ARRAY_BUFFER, ren.BufferData.BufferContent, engo.Gl.STATIC_DRAW)
}

func (l *textShader) generateBufferContent(ren *RenderComponent, space *SpaceComponent, buffer []float32) bool {
	var changed bool

	tint := colorToFloat32(ren.Color)
	txt, ok := ren.Drawable.(Text)
	if !ok {
		unsupportedType(ren.Drawable)
		return false
	}

	atlas, ok := atlasCache[*txt.Font]
	if !ok {
		// Generate texture first
		atlas = txt.Font.generateFontAtlas(UnicodeCap)
		atlasCache[*txt.Font] = atlas
	}

	var currentX float32
	var currentY float32

	var modifier float32 = 1
	if txt.RightToLeft {
		modifier = -1
	}

	letterSpace := float32(txt.Font.Size) * txt.LetterSpacing
	lineSpace := txt.LineSpacing * atlas.Height['X']

	for index, char := range txt.Text {
		// TODO: this might not work for all characters
		switch {
		case char == '\n':
			currentX = 0
			currentY += atlas.Height['X'] + lineSpace
			continue
		case char < 32: // all system stuff should be ignored
			continue
		}

		offset := 20 * index

		// These five are at 0, 0:
		setBufferValue(buffer, 0+offset, currentX, &changed)
		setBufferValue(buffer, 1+offset, currentY, &changed)
		setBufferValue(buffer, 2+offset, atlas.XLocation[char]/atlas.TotalWidth, &changed)
		setBufferValue(buffer, 3+offset, atlas.YLocation[char]/atlas.TotalHeight, &changed)
		setBufferValue(buffer, 4+offset, tint, &changed)

		// These five are at 1, 0:
		setBufferValue(buffer, 5+offset, currentX+atlas.Width[char]+letterSpace, &changed)
		setBufferValue(buffer, 6+offset, currentY, &changed)
		setBufferValue(buffer, 7+offset, (atlas.XLocation[char]+atlas.Width[char])/atlas.TotalWidth, &changed)
		setBufferValue(buffer, 8+offset, atlas.YLocation[char]/atlas.TotalHeight, &changed)
		setBufferValue(buffer, 9+offset, tint, &changed)

		// These five are at 1, 1:
		setBufferValue(buffer, 10+offset, currentX+atlas.Width[char]+letterSpace, &changed)
		setBufferValue(buffer, 11+offset, currentY+atlas.Height[char]+lineSpace, &changed)
		setBufferValue(buffer, 12+offset, (atlas.XLocation[char]+atlas.Width[char])/atlas.TotalWidth, &changed)
		setBufferValue(buffer, 13+offset, (atlas.YLocation[char]+atlas.Height[char])/atlas.TotalHeight, &changed)
		setBufferValue(buffer, 14+offset, tint, &changed)

		// These five are at 0, 1:
		setBufferValue(buffer, 15+offset, currentX, &changed)
		setBufferValue(buffer, 16+offset, currentY+atlas.Height[char]+lineSpace, &changed)
		setBufferValue(buffer, 17+offset, atlas.XLocation[char]/atlas.TotalWidth, &changed)
		setBufferValue(buffer, 18+offset, (atlas.YLocation[char]+atlas.Height[char])/atlas.TotalHeight, &changed)
		setBufferValue(buffer, 19+offset, tint, &changed)

		currentX += modifier * (atlas.Width[char] + letterSpace)
	}

	return changed
}

func (l *textShader) Draw(ren *RenderComponent, space *SpaceComponent) {
<<<<<<< HEAD:common/render_shaders_text_gl.go
	if l.lastBuffer != ren.BufferData.Buffer || ren.BufferData.Buffer == nil {
=======
	txt, ok := ren.Drawable.(Text)
	if !ok {
		unsupportedType(ren.Drawable)
	}

	if l.lastBuffer != ren.Buffer || ren.Buffer == nil {
>>>>>>> master:common/render_shaders_text.go
		l.updateBuffer(ren, space)

		engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, ren.BufferData.Buffer)
		engo.Gl.VertexAttribPointer(l.inPosition, 2, engo.Gl.FLOAT, false, 20, 0)
		engo.Gl.VertexAttribPointer(l.inTexCoords, 2, engo.Gl.FLOAT, false, 20, 8)
		engo.Gl.VertexAttribPointer(l.inColor, 4, engo.Gl.UNSIGNED_BYTE, true, 20, 16)

		l.lastBuffer = ren.BufferData.Buffer
	}

	atlas, ok := atlasCache[*txt.Font]
	if !ok {
		// Generate texture first
		atlas = txt.Font.generateFontAtlas(UnicodeCap)
		atlasCache[*txt.Font] = atlas
	}

	if atlas.Texture != l.lastTexture {
		engo.Gl.BindTexture(engo.Gl.TEXTURE_2D, atlas.Texture)
		l.lastTexture = atlas.Texture
	}

	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_WRAP_S, engo.Gl.CLAMP_TO_EDGE)
	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_WRAP_T, engo.Gl.CLAMP_TO_EDGE)

	if space.Rotation != 0 {
		sin, cos := math.Sincos(space.Rotation * math.Pi / 180)

		l.modelMatrix[0] = ren.Scale.X * engo.GetGlobalScale().X * cos
		l.modelMatrix[1] = ren.Scale.X * engo.GetGlobalScale().X * sin
		l.modelMatrix[3] = ren.Scale.Y * engo.GetGlobalScale().Y * -sin
		l.modelMatrix[4] = ren.Scale.Y * engo.GetGlobalScale().Y * cos
	} else {
		l.modelMatrix[0] = ren.Scale.X * engo.GetGlobalScale().X
		l.modelMatrix[1] = 0
		l.modelMatrix[3] = 0
		l.modelMatrix[4] = ren.Scale.Y * engo.GetGlobalScale().Y
	}

	l.modelMatrix[6] = space.Position.X * engo.GetGlobalScale().X
	l.modelMatrix[7] = space.Position.Y * engo.GetGlobalScale().Y

	engo.Gl.UniformMatrix3fv(l.matrixModel, false, l.modelMatrix)

	engo.Gl.DrawElements(engo.Gl.TRIANGLES, 6*len(txt.Text), engo.Gl.UNSIGNED_SHORT, 0)
}

func (l *textShader) Post() {
	l.lastBuffer = nil
	l.lastTexture = nil

	// Cleanup
	engo.Gl.DisableVertexAttribArray(l.inPosition)
	engo.Gl.DisableVertexAttribArray(l.inTexCoords)
	engo.Gl.DisableVertexAttribArray(l.inColor)

	engo.Gl.BindTexture(engo.Gl.TEXTURE_2D, nil)
	engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, nil)
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, nil)

	engo.Gl.Disable(engo.Gl.BLEND)
}

func (l *textShader) SetCamera(c *CameraSystem) {
	if l.cameraEnabled {
		l.camera = c
	}
}
