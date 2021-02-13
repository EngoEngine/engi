//+build !vulkan

package common

import (
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/gl"
)

type RenderBuffer struct {
	rbo           *gl.RenderBuffer
	width, height int
}

type Framebuffer struct {
	fbo    *gl.FrameBuffer
	oldVP  [4]int32
	isOpen bool
}

type RenderTexture struct {
	tex           *gl.Texture
	width, height float32
	depth         bool
}

func CreateRenderBuffer(width, height int) *RenderBuffer {
	rbuf := &RenderBuffer{
		rbo:    engo.Gl.CreateRenderBuffer(),
		width:  width,
		height: height,
	}
	engo.Gl.BindRenderBuffer(rbuf.rbo)
	engo.Gl.RenderBufferStorage(engo.Gl.RGBA8, width, height)
	engo.Gl.BindRenderBuffer(nil)
	return rbuf
}

func CreateRenderTexture(width, height int, depthBuffer bool) *RenderTexture {
	texBuf := &RenderTexture{
		width:  float32(width),
		height: float32(height),
		tex:    engo.Gl.CreateTexture(),
		depth:  depthBuffer,
	}

	engo.Gl.BindTexture(engo.Gl.TEXTURE_2D, texBuf.tex)

	if depthBuffer {
		engo.Gl.TexImage2DEmpty(engo.Gl.TEXTURE_2D, 0, engo.Gl.DEPTH_COMPONENT, width, height, engo.Gl.DEPTH_COMPONENT, engo.Gl.UNSIGNED_BYTE)
	} else {
		engo.Gl.TexImage2DEmpty(engo.Gl.TEXTURE_2D, 0, engo.Gl.RGBA, width, height, engo.Gl.RGBA, engo.Gl.UNSIGNED_BYTE)
	}
	if err := engo.Gl.GetError(); err != 0 {
		panic(err)
	}
	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_MAG_FILTER, engo.Gl.NEAREST)
	engo.Gl.TexParameteri(engo.Gl.TEXTURE_2D, engo.Gl.TEXTURE_MIN_FILTER, engo.Gl.NEAREST)

	return texBuf
}

func CreateFramebuffer() *Framebuffer {
	return &Framebuffer{
		fbo: engo.Gl.CreateFrameBuffer(),
	}
}

func (t *RenderTexture) Bind() {
	if t.depth {
		engo.Gl.FrameBufferTexture2D(engo.Gl.FRAMEBUFFER, engo.Gl.DEPTH_ATTACHMENT, engo.Gl.TEXTURE_2D, t.tex, 0)
	} else {
		engo.Gl.FrameBufferTexture2D(engo.Gl.FRAMEBUFFER, engo.Gl.COLOR_ATTACHMENT0, engo.Gl.TEXTURE_2D, t.tex, 0)
	}
}

func (t *RenderTexture) Close() {
	engo.Gl.DeleteTexture(t.tex)
}

// Width returns the width of the texture.
func (t *RenderTexture) Width() float32 {
	return t.width
}

// Height returns the height of the texture.
func (t *RenderTexture) Height() float32 {
	return t.height
}

// Texture returns the OpenGL ID of the Texture.
func (t *RenderTexture) Texture() *gl.Texture {
	return t.tex
}

// View returns the viewport properties of the Texture. The order is Min.X, Min.Y, Max.X, Max.Y.
func (t *RenderTexture) View() (float32, float32, float32, float32) {
	return 0, 0, 1, 1
}

func (rb *RenderBuffer) Bind(attachment int) {
	engo.Gl.FrameBufferRenderBuffer(engo.Gl.FRAMEBUFFER, attachment, rb.rbo)
}

func (rb *RenderBuffer) Destroy() {
	engo.Gl.DeleteRenderBuffer(rb.rbo)
}

func (fb *Framebuffer) Open(width, height int) {
	if fb.isOpen {
		return
	}
	engo.Gl.BindFrameBuffer(fb.fbo)
	fb.oldVP = engo.Gl.GetViewport()
	engo.Gl.Viewport(0, 0, width, height)
	fb.isOpen = true
}

func (fb *Framebuffer) Close() {
	if !fb.isOpen {
		return
	}
	engo.Gl.BindFrameBuffer(nil)
	engo.Gl.Viewport(int(fb.oldVP[0]), int(fb.oldVP[1]), int(fb.oldVP[2]), int(fb.oldVP[3]))
	fb.isOpen = false
}

func (fb *Framebuffer) Destroy() {
	engo.Gl.DeleteFrameBuffer(fb.fbo)
}
