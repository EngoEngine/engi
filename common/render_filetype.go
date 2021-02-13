package common

import (
	"fmt"
	"image"
	"image/draw"

	// imported to decode jpegs and upload them to the GPU.
	_ "image/jpeg"
	// imported to decode .pngs and upload them to the GPU.
	_ "image/png"
	// imported to decode .gifs and uppload them to the GPU.
	_ "image/gif"
	"io"

	// these are for svg support

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"

	"github.com/EngoEngine/engo"
)

// imgLoader is the shared imageLoader for all image file formats
var imgLoader *imageLoader

// TextureResource is the resource used by the RenderSystem. It uses .jpg, .gif, and .png images
type TextureResource struct {
	Texture  TextureID
	Width    float32
	Height   float32
	Viewport *engo.AABB
	url      string
}

// URL is the file path of the TextureResource
func (t TextureResource) URL() string {
	return t.url
}

type imageLoader struct {
	images map[string]TextureResource
}

func (i *imageLoader) Load(url string, data io.Reader) error {
	var res TextureResource
	if getExt(url) == ".svg" {
		icon, err := oksvg.ReadIconStream(data, oksvg.WarnErrorMode)
		if err != nil {
			return err
		}
		w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		gv := rasterx.NewScannerGV(w, h, img, img.Bounds())
		r := rasterx.NewDasher(w, h, gv)
		icon.Draw(r, 1.0)
		b := img.Bounds()
		newm := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(newm, newm.Bounds(), img, b.Min, draw.Src)
		res = NewTextureResource(&ImageObject{newm})
	} else {
		img, _, err := image.Decode(data)
		if err != nil {
			return err
		}
		b := img.Bounds()
		newm := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(newm, newm.Bounds(), img, b.Min, draw.Src)
		res = NewTextureResource(&ImageObject{newm})
	}
	res.url = url
	i.images[url] = res

	return nil
}

func (i *imageLoader) Unload(url string) error {
	delete(i.images, url)
	return nil
}

func (i *imageLoader) Resource(url string) (engo.Resource, error) {
	texture, ok := i.images[url]
	if !ok {
		return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
	}

	return texture, nil
}

// Image holds data and properties of an .jpg, .gif, or .png file
type Image interface {
	Data() interface{}
	Width() int
	Height() int
}

// UploadTexture sends the image to the GPU, to be kept in GPU RAM
func UploadTexture(img Image) TextureID {
	var id TextureID
	if !engo.Headless() {
		id = createTextureID(img)
	}
	return id
}

// NewTextureResource sends the image to the GPU and returns a `TextureResource` for easy access
func NewTextureResource(img Image) TextureResource {
	id := UploadTexture(img)
	return TextureResource{Texture: id, Width: float32(img.Width()), Height: float32(img.Height())}
}

// NewTextureSingle sends the image to the GPU and returns a `Texture` with a viewport for single-sprite images
func NewTextureSingle(img Image) Texture {
	id := UploadTexture(img)
	return Texture{id, float32(img.Width()), float32(img.Height()), engo.AABB{Max: engo.Point{X: 1.0, Y: 1.0}}}
}

// ImageToNRGBA takes a given `image.Image` and converts it into an `image.NRGBA`. Especially useful when transforming
// image.Uniform to something usable by `engo`.
func ImageToNRGBA(img image.Image, width, height int) *image.NRGBA {
	newm := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(newm, newm.Bounds(), img, image.Point{0, 0}, draw.Src)

	return newm
}

// ImageObject is a pure Go implementation of a `Drawable`
type ImageObject struct {
	data *image.NRGBA
}

// NewImageObject creates a new ImageObject given the image.NRGBA reference
func NewImageObject(img *image.NRGBA) *ImageObject {
	return &ImageObject{img}
}

// Data returns the entire image.NRGBA object
func (i *ImageObject) Data() interface{} {
	return i.data
}

// Width returns the maximum X coordinate of the image
func (i *ImageObject) Width() int {
	return i.data.Rect.Max.X
}

// Height returns the maximum Y coordinate of the image
func (i *ImageObject) Height() int {
	return i.data.Rect.Max.Y
}

// LoadedSprite loads the texture-reference from `engo.Files`, and wraps it in a `*Texture`.
// This method is intended for image-files which represent entire sprites.
func LoadedSprite(url string) (*Texture, error) {
	res, err := engo.Files.Resource(url)
	if err != nil {
		return nil, err
	}

	img, ok := res.(TextureResource)
	if !ok {
		return nil, fmt.Errorf("resource not of type `TextureResource`: %s", url)
	}

	viewport := engo.AABB{Max: engo.Point{X: 1.0, Y: 1.0}}
	if img.Viewport != nil {
		viewport = *img.Viewport
	}
	return &Texture{img.Texture, img.Width, img.Height, viewport}, nil
}

// Texture represents a texture loaded in the GPU RAM (by using OpenGL), which defined dimensions and viewport
type Texture struct {
	id       TextureID
	width    float32
	height   float32
	viewport engo.AABB
}

// Width returns the width of the texture.
func (t Texture) Width() float32 {
	return t.width
}

// Height returns the height of the texture.
func (t Texture) Height() float32 {
	return t.height
}

// Texture returns the OpenGL ID of the Texture.
func (t Texture) Texture() TextureID {
	return t.id
}

// View returns the viewport properties of the Texture. The order is Min.X, Min.Y, Max.X, Max.Y.
func (t Texture) View() (float32, float32, float32, float32) {
	return t.viewport.Min.X, t.viewport.Min.Y, t.viewport.Max.X, t.viewport.Max.Y
}

// Close removes the Texture data from the GPU.
func (t Texture) Close() {
	if !engo.Headless() {
		t.close()
	}
}

func init() {
	imgLoader = &imageLoader{images: make(map[string]TextureResource)}
	engo.Files.Register(".jpg", imgLoader)
	engo.Files.Register(".png", imgLoader)
	engo.Files.Register(".gif", imgLoader)
	engo.Files.Register(".svg", imgLoader)
}
