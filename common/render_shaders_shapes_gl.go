//+build !vulkan

package common

import (
	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/gl"
	"github.com/EngoEngine/math"
)

type legacyShader struct {
	program *gl.Program

	indicesRectangles    []uint16
	indicesRectanglesVBO *gl.Buffer

	inPosition int
	inColor    int

	matrixProjection *gl.UniformLocation
	matrixView       *gl.UniformLocation
	matrixModel      *gl.UniformLocation

	projectionMatrix []float32
	viewMatrix       []float32
	modelMatrix      []float32

	camera        *CameraSystem
	cameraEnabled bool

	lastBuffer *gl.Buffer
}

func (l *legacyShader) Setup(w *ecs.World) error {
	var err error
	l.program, err = LoadShader(`
attribute vec2 in_Position;
attribute vec4 in_Color;

uniform mat3 matrixProjection;
uniform mat3 matrixView;
uniform mat3 matrixModel;

varying vec4 var_Color;

void main() {
  var_Color = in_Color;

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

void main (void) {
  gl_FragColor = var_Color;
}`)

	if err != nil {
		return err
	}

	// Create and populate indices buffer
	l.indicesRectangles = []uint16{0, 1, 2, 0, 2, 3}
	l.indicesRectanglesVBO = engo.Gl.CreateBuffer()
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, l.indicesRectanglesVBO)
	engo.Gl.BufferData(engo.Gl.ELEMENT_ARRAY_BUFFER, l.indicesRectangles, engo.Gl.STATIC_DRAW)

	// Define things that should be read from the texture buffer
	l.inPosition = engo.Gl.GetAttribLocation(l.program, "in_Position")
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

func (l *legacyShader) Pre() {
	engo.Gl.Enable(engo.Gl.BLEND)
	engo.Gl.BlendFunc(engo.Gl.SRC_ALPHA, engo.Gl.ONE_MINUS_SRC_ALPHA)

	// Bind shader and buffer, enable attributes
	engo.Gl.UseProgram(l.program)
	engo.Gl.EnableVertexAttribArray(l.inPosition)
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

func (l *legacyShader) updateBuffer(ren *RenderComponent, space *SpaceComponent) {
	if len(ren.BufferData.BufferContent) == 0 {
		ren.BufferData.BufferContent = make([]float32, l.computeBufferSize(ren.Drawable)) // because we add at most this many elements to it
	}
	if changed := l.generateBufferContent(ren, space, ren.BufferData.BufferContent); !changed {
		return
	}

	if ren.BufferData.BufferContent == nil {
		ren.BufferData.Buffer = engo.Gl.CreateBuffer()
	}
	engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, ren.BufferData.Buffer)
	engo.Gl.BufferData(engo.Gl.ARRAY_BUFFER, ren.BufferData.BufferContent, engo.Gl.STATIC_DRAW)
}

func (l *legacyShader) computeBufferSize(draw Drawable) int {
	switch shape := draw.(type) {
	case Triangle:
		return 65
	case Rectangle:
		return 90
	case Circle:
		return 1800
	case ComplexTriangles:
		return len(shape.Points) * 6
	case Curve:
		return 1800
	default:
		return 0
	}
}

func (l *legacyShader) generateBufferContent(ren *RenderComponent, space *SpaceComponent, buffer []float32) bool {
	w := space.Width
	h := space.Height

	var changed bool

	tint := colorToFloat32(ren.Color)

	switch shape := ren.Drawable.(type) {
	case Triangle:
		switch shape.TriangleType {
		case TriangleIsosceles:
			setBufferValue(buffer, 0, w/2, &changed)
			//setBufferValue(buffer, 1, 0, &changed)
			setBufferValue(buffer, 2, tint, &changed)

			setBufferValue(buffer, 3, w, &changed)
			setBufferValue(buffer, 4, h, &changed)
			setBufferValue(buffer, 5, tint, &changed)

			//setBufferValue(buffer, 6, 0, &changed)
			setBufferValue(buffer, 7, h, &changed)
			setBufferValue(buffer, 8, tint, &changed)

			if shape.BorderWidth > 0 {
				borderTint := colorToFloat32(shape.BorderColor)
				b := shape.BorderWidth
				s, c := math.Sincos(math.Atan(2 * h / w))

				pts := [][]float32{
					//Left
					{w / 2, 0},
					{0, h},
					{b, h},
					{b, h},
					{(w / 2) + b*c, b * s},
					{w / 2, 0},
					//Right
					{w / 2, 0},
					{w, h},
					{w - b, h},
					{w - b, h},
					{(w / 2) - b*c, b * s},
					{w / 2, 0},
					//Bottom
					{0, h},
					{w, h},
					{b * c, h - b*s},
					{b * c, h - b*s},
					{w - b*c, h - b*s},
					{w, h},
				}

				for i, p := range pts {
					setBufferValue(buffer, 9+3*i, p[0], &changed)
					setBufferValue(buffer, 10+3*i, p[1], &changed)
					setBufferValue(buffer, 11+3*i, borderTint, &changed)
				}
			}
		case TriangleRight:
			//setBufferValue(buffer, 0, 0, &changed)
			//setBufferValue(buffer, 1, 0, &changed)
			setBufferValue(buffer, 2, tint, &changed)

			setBufferValue(buffer, 3, w, &changed)
			setBufferValue(buffer, 4, h, &changed)
			setBufferValue(buffer, 5, tint, &changed)

			//setBufferValue(buffer, 6, 0, &changed)
			setBufferValue(buffer, 7, h, &changed)
			setBufferValue(buffer, 8, tint, &changed)

			if shape.BorderWidth > 0 {
				borderTint := colorToFloat32(shape.BorderColor)
				b := shape.BorderWidth

				pts := [][]float32{
					//Left
					{0, 0},
					{0, h},
					{b, h},
					{b, h},
					{b, b * h / w},
					{0, 0},
					//Right
					{0, 0},
					{w, h},
					{w - b, h},
					{w - b, h},
					{0, b},
					{0, 0},
					//Bottom
					{0, h},
					{w, h},
					{w - b*w/h, h - b},
					{w - b*w/h, h - b},
					{0, h - b},
					{0, h},
				}

				for i, p := range pts {
					setBufferValue(buffer, 9+3*i, p[0], &changed)
					setBufferValue(buffer, 10+3*i, p[1], &changed)
					setBufferValue(buffer, 11+3*i, borderTint, &changed)
				}
			}
		}

	case Circle:
		if shape.Arc == 0 {
			shape.Arc = 360
		}
		theta := float32(2.0*math.Pi/298.0) * shape.Arc / 360
		cx := w / 2
		bx := shape.BorderWidth
		cy := h / 2
		by := shape.BorderWidth
		var borderTint float32
		hasBorder := shape.BorderWidth > 0
		if hasBorder {
			borderTint = colorToFloat32(shape.BorderColor)
		}
		setBufferValue(buffer, 0, w/2, &changed)
		setBufferValue(buffer, 1, h/2, &changed)
		setBufferValue(buffer, 2, tint, &changed)
		if hasBorder {
			setBufferValue(buffer, 900, w/2, &changed)
			setBufferValue(buffer, 901, h/2, &changed)
			setBufferValue(buffer, 902, borderTint, &changed)
		}
		for i := 1; i < 300; i++ {
			s, c := math.Sincos(float32(i) * theta)
			setBufferValue(buffer, i*3, cx+cx*c-bx, &changed)
			setBufferValue(buffer, i*3+1, cy+cy*s-by, &changed)
			setBufferValue(buffer, i*3+2, tint, &changed)
			if hasBorder {
				setBufferValue(buffer, i*3+900, cx+cx*c, &changed)
				setBufferValue(buffer, i*3+901, cy+cy*s, &changed)
				setBufferValue(buffer, i*3+902, borderTint, &changed)
			}
		}

	case Rectangle:
		//setBufferValue(buffer, 0, 0, &changed)
		//setBufferValue(buffer, 1, 0, &changed)
		setBufferValue(buffer, 2, tint, &changed)

		setBufferValue(buffer, 3, w, &changed)
		//setBufferValue(buffer, 4, 0, &changed)
		setBufferValue(buffer, 5, tint, &changed)

		setBufferValue(buffer, 6, w, &changed)
		setBufferValue(buffer, 7, h, &changed)
		setBufferValue(buffer, 8, tint, &changed)

		setBufferValue(buffer, 9, w, &changed)
		setBufferValue(buffer, 10, h, &changed)
		setBufferValue(buffer, 11, tint, &changed)

		//setBufferValue(buffer, 12, 0, &changed)
		setBufferValue(buffer, 13, h, &changed)
		setBufferValue(buffer, 14, tint, &changed)

		//setBufferValue(buffer, 15, 0, &changed)
		//setBufferValue(buffer, 16, 0, &changed)
		setBufferValue(buffer, 17, tint, &changed)

		if shape.BorderWidth > 0 {
			borderTint := colorToFloat32(shape.BorderColor)
			b := shape.BorderWidth
			pts := [][]float32{
				//Top
				{0, 0},
				{w, 0},
				{w, b},
				{w, b},
				{0, b},
				{0, 0},
				//Right
				{w - b, b},
				{w, b},
				{w, h - b},
				{w, h - b},
				{w - b, h - b},
				{w - b, b},
				//Bottom
				{w, h - b},
				{w, h},
				{0, h},
				{0, h},
				{0, h - b},
				{w, h - b},
				//Left
				{0, b},
				{b, b},
				{b, h - b},
				{b, h - b},
				{0, h - b},
				{0, b},
			}

			for i, p := range pts {
				setBufferValue(buffer, 18+3*i, p[0], &changed)
				setBufferValue(buffer, 19+3*i, p[1], &changed)
				setBufferValue(buffer, 20+3*i, borderTint, &changed)
			}
		}

	case ComplexTriangles:
		var index int
		for _, point := range shape.Points {
			setBufferValue(buffer, index, point.X*w, &changed)
			setBufferValue(buffer, index+1, point.Y*h, &changed)
			setBufferValue(buffer, index+2, tint, &changed)
			index += 3
		}

		if shape.BorderWidth > 0 {
			borderTint := colorToFloat32(shape.BorderColor)

			for _, point := range shape.Points {
				setBufferValue(buffer, index, point.X*w, &changed)
				setBufferValue(buffer, index+1, point.Y*h, &changed)
				setBufferValue(buffer, index+2, borderTint, &changed)
				index += 3
			}
		}
	case Curve:
		lw := shape.LineWidth
		pts := make([][]float32, 0)
		for i := 0; i < 100; i++ {
			pt := make([]float32, 2)
			t := float32(i) / 100
			switch len(shape.Points) {
			case 0:
				pt[0] = t * w
				pt[1] = t * h
			case 1:
				pt[0] = 2*(1-t)*t*shape.Points[0].X + t*t*w
				pt[1] = 2*(1-t)*t*shape.Points[0].Y + t*t*h
			case 2:
				pt[0] = 3*(1-t)*(1-t)*t*shape.Points[0].X + 3*(1-t)*t*t*shape.Points[1].X + t*t*t*w
				pt[1] = 3*(1-t)*(1-t)*t*shape.Points[0].Y + 3*(1-t)*t*t*shape.Points[1].Y + t*t*t*h
			default:
				unsupportedType(ren.Drawable)
			}
			pts = append(pts, pt)
		}
		for i := 0; i < len(pts)-1; i++ {
			num := pts[i+1][1] - pts[i][1]
			if engo.FloatEqual(num, 0) { //horizontal line
				setBufferValue(buffer, i*18, pts[i][0], &changed)
				setBufferValue(buffer, i*18+1, pts[i][1]-lw, &changed)
				setBufferValue(buffer, i*18+2, tint, &changed)
				setBufferValue(buffer, i*18+3, pts[i+1][0], &changed)
				setBufferValue(buffer, i*18+4, pts[i+1][1]-lw, &changed)
				setBufferValue(buffer, i*18+5, tint, &changed)
				setBufferValue(buffer, i*18+6, pts[i+1][0], &changed)
				setBufferValue(buffer, i*18+7, pts[i+1][1]+lw, &changed)
				setBufferValue(buffer, i*18+8, tint, &changed)
				setBufferValue(buffer, i*18+9, pts[i+1][0], &changed)
				setBufferValue(buffer, i*18+10, pts[i+1][1]+lw, &changed)
				setBufferValue(buffer, i*18+11, tint, &changed)
				setBufferValue(buffer, i*18+12, pts[i][0], &changed)
				setBufferValue(buffer, i*18+13, pts[i][1]+lw, &changed)
				setBufferValue(buffer, i*18+14, tint, &changed)
				setBufferValue(buffer, i*18+15, pts[i][0], &changed)
				setBufferValue(buffer, i*18+16, pts[i][1]-lw, &changed)
				setBufferValue(buffer, i*18+17, tint, &changed)
				continue
			}
			denom := pts[i+1][0] - pts[i+1][0]
			if engo.FloatEqual(denom, 0) { //vertical line
				setBufferValue(buffer, i*18, pts[i+1][0]-lw, &changed)
				setBufferValue(buffer, i*18+1, pts[i+1][1], &changed)
				setBufferValue(buffer, i*18+2, tint, &changed)
				setBufferValue(buffer, i*18+3, pts[i+1][0]+lw, &changed)
				setBufferValue(buffer, i*18+4, pts[i+1][1], &changed)
				setBufferValue(buffer, i*18+5, tint, &changed)
				setBufferValue(buffer, i*18+6, pts[i][0]+lw, &changed)
				setBufferValue(buffer, i*18+7, pts[i][1], &changed)
				setBufferValue(buffer, i*18+8, tint, &changed)
				setBufferValue(buffer, i*18+9, pts[i][0]+lw, &changed)
				setBufferValue(buffer, i*18+10, pts[i][1], &changed)
				setBufferValue(buffer, i*18+11, tint, &changed)
				setBufferValue(buffer, i*18+12, pts[i][0]-lw, &changed)
				setBufferValue(buffer, i*18+13, pts[i][1], &changed)
				setBufferValue(buffer, i*18+14, tint, &changed)
				setBufferValue(buffer, i*18+15, pts[i+1][0]-lw, &changed)
				setBufferValue(buffer, i*18+16, pts[i+1][1], &changed)
				setBufferValue(buffer, i*18+17, tint, &changed)
				continue
			}
			m1 := num / denom
			m2 := -1 / m1
			dx := math.Sqrt(lw*lw/(1+m2*m2)) / 2
			dy := m2 * dx
			setBufferValue(buffer, i*18, pts[i][0]-dx, &changed)
			setBufferValue(buffer, i*18+1, pts[i][1]-dy, &changed)
			setBufferValue(buffer, i*18+2, tint, &changed)
			setBufferValue(buffer, i*18+3, pts[i+1][0]-dx, &changed)
			setBufferValue(buffer, i*18+4, pts[i+1][1]-dy, &changed)
			setBufferValue(buffer, i*18+5, tint, &changed)
			setBufferValue(buffer, i*18+6, pts[i+1][0]+dx, &changed)
			setBufferValue(buffer, i*18+7, pts[i+1][1]+dy, &changed)
			setBufferValue(buffer, i*18+8, tint, &changed)
			setBufferValue(buffer, i*18+9, pts[i+1][0]+dx, &changed)
			setBufferValue(buffer, i*18+10, pts[i+1][1]+dy, &changed)
			setBufferValue(buffer, i*18+11, tint, &changed)
			setBufferValue(buffer, i*18+12, pts[i][0]+dx, &changed)
			setBufferValue(buffer, i*18+13, pts[i][1]+dy, &changed)
			setBufferValue(buffer, i*18+14, tint, &changed)
			setBufferValue(buffer, i*18+15, pts[i][0]-dx, &changed)
			setBufferValue(buffer, i*18+16, pts[i][1]-dy, &changed)
			setBufferValue(buffer, i*18+17, tint, &changed)
		}
	default:
		unsupportedType(ren.Drawable)
	}

	return changed
}

func (l *legacyShader) Draw(ren *RenderComponent, space *SpaceComponent) {
	if l.lastBuffer != ren.BufferData.Buffer || ren.BufferData.Buffer == nil {
		l.updateBuffer(ren, space)

		engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, ren.BufferData.Buffer)
		engo.Gl.VertexAttribPointer(l.inPosition, 2, engo.Gl.FLOAT, false, 12, 0)
		engo.Gl.VertexAttribPointer(l.inColor, 4, engo.Gl.UNSIGNED_BYTE, true, 12, 8)

		l.lastBuffer = ren.BufferData.Buffer
	}

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

	switch shape := ren.Drawable.(type) {
	case Triangle:
		num := 3
		if shape.BorderWidth > 0 {
			num = 21
		}
		engo.Gl.DrawArrays(engo.Gl.TRIANGLES, 0, num)
	case Rectangle:
		num := 6
		if shape.BorderWidth > 0 {
			num = 30
		}
		engo.Gl.DrawArrays(engo.Gl.TRIANGLES, 0, num)
	case Circle:
		if shape.BorderWidth > 0 {
			if engo.FloatEqual(shape.Arc, 360) || engo.FloatEqual(shape.Arc, 0) {
				engo.Gl.DrawArrays(engo.Gl.TRIANGLE_FAN, 300, 300)
			} else {
				engo.Gl.DrawArrays(engo.Gl.TRIANGLE_FAN, 300, 290)
			}
		}
		engo.Gl.DrawArrays(engo.Gl.TRIANGLE_FAN, 0, 300)
	case ComplexTriangles:
		engo.Gl.DrawArrays(engo.Gl.TRIANGLES, 0, len(shape.Points))

		if shape.BorderWidth > 0 {
			borderWidth := shape.BorderWidth
			if l.cameraEnabled {
				borderWidth /= l.camera.z
			}
			engo.Gl.LineWidth(borderWidth)
			engo.Gl.DrawArrays(engo.Gl.LINE_LOOP, len(shape.Points), len(shape.Points))
		}
	case Curve:
		engo.Gl.DrawArrays(engo.Gl.TRIANGLES, 0, 600)
	default:
		unsupportedType(ren.Drawable)
	}
}

func (l *legacyShader) Post() {
	l.lastBuffer = nil

	// Cleanup
	engo.Gl.DisableVertexAttribArray(l.inPosition)
	engo.Gl.DisableVertexAttribArray(l.inColor)

	engo.Gl.BindBuffer(engo.Gl.ARRAY_BUFFER, nil)
	engo.Gl.BindBuffer(engo.Gl.ELEMENT_ARRAY_BUFFER, nil)

	engo.Gl.Disable(engo.Gl.BLEND)
}

func (l *legacyShader) SetCamera(c *CameraSystem) {
	if l.cameraEnabled {
		l.camera = c
	}
}
