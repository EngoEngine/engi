package common

import (
	"encoding/xml"
	"fmt"
	"io"
	"path"

	"github.com/EngoEngine/engo"
)

// TextureAtlas is a collection of small textures grouped into a big image
type TextureAtlas struct {
	XMLName xml.Name `xml:"TextureAtlas"`
	Text    string   `xml:",chardata"`
	// ImagePath is the path of the main image all the textures will be derived from
	ImagePath string `xml:"imagePath,attr"`
	// SubTextures is a slice of SubTextures
	SubTextures []SubTexture `xml:"SubTexture"`
}

// SubTexture represents a texture from a region in the TextureAtlas
type SubTexture struct {
	Text string `xml:",chardata"`
	// Name is the location of the subtexture before it was packed, Used as the url in the image loader
	Name string `xml:"name,attr"`
	// X coordinate of the subtexture in reference to the main image
	X float32 `xml:"x,attr"`
	// Y coordinate of the subtexture in reference to the main image
	Y float32 `xml:"y,attr"`
	// Width of the subtexture in reference to the main image
	Width float32 `xml:"width,attr"`
	// Height of the subtexture in reference to the main image
	Height float32 `xml:"height,attr"`
}

// TextureAtlasResource contains reference to a loaded TextureAtlas and the texture of the main image
type TextureAtlasResource struct {
	// texture is a the TextureID of the main image
	texture TextureID
	// url is the location of the xml file
	url string
	// Atlas is the TextureAtlas filled with data from the parsed XML file
	Atlas *TextureAtlas
}

// URL retrieves the url to the .xml file
func (r TextureAtlasResource) URL() string {
	return r.url
}

// textureAtlasLoader is reponsible for managing '.xml' files exported from TexturePacker (https://www.codeandweb.com/texturepacker)
type textureAtlasLoader struct {
	atlases map[string]*TextureAtlasResource
}

// Load will load the xml file and the main image as well as add references
// for sub textures/images in engo.Files, subtextures keep their path url (with appended extension from main image path if it does not exist),
// the main image is loaded in reference to the directory of the xml file
// For example this sub texture:
//  <SubTexture name="subimg" x="10" y="10" width="50" height="50"/>
// can be retrieved with this go code
//  texture, err := common.LoadedSprite("subimg.png")
func (t *textureAtlasLoader) Load(url string, data io.Reader) error {
	atlas, err := createAtlasFromXML(data, url)
	if err != nil {
		return err
	}

	t.atlases[url] = atlas
	return nil
}

// Unload removes the preloaded atlass from the cache and clears
// references to all SubTextures from the image loader
func (t *textureAtlasLoader) Unload(url string) error {
	imgURL := path.Join(path.Dir(url), t.atlases[url].Atlas.ImagePath)
	if err := imgLoader.Unload(imgURL); err != nil {
		return err
	}
	for _, subTexture := range t.atlases[url].Atlas.SubTextures {
		if err := imgLoader.Unload(subTexture.Name); err != nil {
			return err
		}
	}

	delete(t.atlases, url)
	return nil
}

// Resource retrieves and returns the texture atlas of type TextureAtlasResource
func (t *textureAtlasLoader) Resource(url string) (engo.Resource, error) {
	atlas, ok := t.atlases[url]
	if !ok {
		return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
	}

	return atlas, nil
}

// createAtlasFromXML unmarshals and unpacks the xml data into a TextureAtlas
// it also adds the main image and subtextures to the imageLoader
// if the subtexture doesn't have an extension in it's Name field,
// it will append the main image's extension to it
func createAtlasFromXML(r io.Reader, url string) (*TextureAtlasResource, error) {
	var atlas *TextureAtlas
	err := xml.NewDecoder(r).Decode(&atlas)
	if err != nil {
		return nil, err
	}

	imgURL := path.Join(path.Dir(url), atlas.ImagePath)
	if err := engo.Files.Load(imgURL); err != nil {
		return nil, fmt.Errorf("failed load texture atlas image: %v", err)
	}

	res, err := engo.Files.Resource(imgURL)
	if err != nil {
		return nil, err
	}

	img, ok := res.(TextureResource)
	if !ok {
		return nil, fmt.Errorf("resource not of type `TextureResource`: %v", url)
	}

	ext := path.Ext(atlas.ImagePath)
	for i, subTexture := range atlas.SubTextures {
		texture := &Texture{
			id:     img.Texture,
			width:  subTexture.Width,
			height: subTexture.Height,
		}

		viewport := engo.AABB{
			Min: engo.Point{
				X: subTexture.X / img.Width,
				Y: subTexture.Y / img.Height,
			},
			Max: engo.Point{
				X: (subTexture.X + subTexture.Width) / img.Width,
				Y: (subTexture.Y + subTexture.Height) / img.Height,
			},
		}

		subtextureURL := subTexture.Name
		if path.Ext(subTexture.Name) == "" {
			subtextureURL += ext
			atlas.SubTextures[i].Name = subtextureURL
		}

		imgLoader.images[subtextureURL] = TextureResource{Texture: texture.id, Width: texture.width, Height: texture.height, Viewport: &viewport}
	}

	return &TextureAtlasResource{
		Atlas:   atlas,
		url:     url,
		texture: img.Texture,
	}, nil

}

func init() {
	engo.Files.Register(".xml", &textureAtlasLoader{atlases: make(map[string]*TextureAtlasResource)})
}
