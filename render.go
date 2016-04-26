package engo

import (
	"fmt"
	"image/color"
	"sort"

	"engo.io/ecs"
	"engo.io/gl"
)

const (
	RenderSystemPriority = -1000
)

type renderChangeMessage struct{}

func (renderChangeMessage) Type() string {
	return "renderChangeMessage"
}

type Drawable interface {
	Texture() *gl.Texture
	Width() float32
	Height() float32
	View() (float32, float32, float32, float32)
}

type RenderComponent struct {
	// Hidden is used to prevent drawing by OpenGL
	Hidden bool
	// Scale is the scale at which to render, in the X and Y axis
	Scale Point
	// Color is not tested at the moment - TODO: make sure it works, and document what it does
	Color color.Color
	// Drawable refers to the Texture that should be drawn
	Drawable Drawable

	shader Shader
	zIndex float32

	buffer        *gl.Buffer
	bufferContent []float32
}

func NewRenderComponent(d Drawable, scale Point) RenderComponent {
	rc := RenderComponent{
		Color:    color.White,
		Scale:    scale,
		Drawable: d,
	}

	return rc
}

func (r *RenderComponent) SetShader(s Shader) {
	r.shader = s
	Mailbox.Dispatch(&renderChangeMessage{})
}

func (r *RenderComponent) SetZIndex(index float32) {
	r.zIndex = index
	Mailbox.Dispatch(&renderChangeMessage{})
}

type renderEntity struct {
	*ecs.BasicEntity
	*RenderComponent
	*SpaceComponent
}

type renderEntityList []renderEntity

func (r renderEntityList) Len() int {
	return len(r)
}

func (r renderEntityList) Less(i, j int) bool {
	// Sort by shader-pointer if they have the same zIndex
	if r[i].RenderComponent.zIndex == r[j].RenderComponent.zIndex {
		// TODO: optimize this for performance
		return fmt.Sprintf("%p", r[i].RenderComponent.shader) < fmt.Sprintf("%p", r[j].RenderComponent.shader)
	}

	return r[i].RenderComponent.zIndex < r[j].RenderComponent.zIndex
}

func (r renderEntityList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

type RenderSystem struct {
	entities renderEntityList
	world    *ecs.World

	sortingNeeded bool
	currentShader Shader
}

func (*RenderSystem) Priority() int { return RenderSystemPriority }

func (rs *RenderSystem) New(w *ecs.World) {
	rs.world = w

	if !headless {
		initShaders()
	}

	Mailbox.Listen("renderChangeMessage", func(Message) {
		rs.sortingNeeded = true
	})
}

func (rs *RenderSystem) Add(basic *ecs.BasicEntity, render *RenderComponent, space *SpaceComponent) {
	rs.entities = append(rs.entities, renderEntity{basic, render, space})
	rs.sortingNeeded = true
}

func (rs *RenderSystem) Remove(basic ecs.BasicEntity) {
	var delete int = -1
	for index, entity := range rs.entities {
		if entity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		rs.entities = append(rs.entities[:delete], rs.entities[delete+1:]...)
		rs.sortingNeeded = true
	}
}

func (rs *RenderSystem) Update(dt float32) {
	if headless {
		return
	}

	if rs.sortingNeeded {
		sort.Sort(rs.entities)
		rs.sortingNeeded = false
	}

	Gl.Clear(Gl.COLOR_BUFFER_BIT)

	// TODO: it's linear for now, but that might very well be a bad idea
	for _, e := range rs.entities {
		if e.RenderComponent.Hidden {
			continue // with other entities
		}

		// Retrieve a shader, may be the default one -- then use it if we aren't already using it
		shader := e.RenderComponent.shader
		if shader == nil {
			shader = DefaultShader
		}

		// Change Shader if we have to
		if shader != rs.currentShader {
			if rs.currentShader != nil {
				rs.currentShader.Post()
			}
			shader.Pre()
			rs.currentShader = shader
		}

		rs.currentShader.UpdateBuffer(e.RenderComponent)

		rs.currentShader.Draw(e.RenderComponent.Drawable.Texture(), e.RenderComponent.buffer,
			e.SpaceComponent.Position.X, e.SpaceComponent.Position.Y,
			e.RenderComponent.Scale.X, e.RenderComponent.Scale.Y,
			e.SpaceComponent.Rotation)
	}

	if rs.currentShader != nil {
		rs.currentShader.Post()
		rs.currentShader = nil
	}
}
