package common

import (
	"github.com/EngoEngine/engo"
	"github.com/EngoEngine/engo/math"
)

const (
	orth = "orthogonal"
	iso  = "isometric"
)

// Level is a parsed TMX level containing all layers and default Tiled attributes
type Level struct {
	// Orientation is the parsed level orientation from the TMX XML, like orthogonal, isometric, etc.
	Orientation string
	// RenderOrder is the in Tiled specified TileMap render order, like right-down, right-up, etc.
	RenderOrder string
	width       int
	height      int
	// TileWidth defines the width of each tile in the level
	TileWidth int
	// TileHeight defines the height of each tile in the level
	TileHeight int
	// NextObjectId is the next free Object ID defined by Tiled
	NextObjectID int
	// TileLayers contains all TileLayer of the level
	TileLayers []*TileLayer
	// ImageLayers contains all ImageLayer of the level
	ImageLayers []*ImageLayer
	// ObjectLayers contains all ObjectLayer of the level
	ObjectLayers []*ObjectLayer
	// Properties are custom properties of the level
	Properties  []Property
	resourceMap map[uint32]Texture
	pointMap    map[mapPoint]*Tile
	framesMap   map[uint32][]uint32
}

// Property is any custom property. The Type corresponds to the type (int,
// float, etc) stored in the Value as a string
type Property struct {
	Name, Type, Value string
}

// TileLayer contains a list of its tiles plus all default Tiled attributes
type TileLayer struct {
	// Name defines the name of the tile layer given in the TMX XML / Tiled
	Name string
	// Width is the integer width of each tile in this layer
	Width int
	// Height is the integer height of each tile in this layer
	Height int
	// Tiles contains the list of tiles
	Tiles []*Tile
	// Opacity is the opacity of the layer from [0,1]
	Opacity float32
	// Visible is if the layer is visible
	Visible bool
	// X is the x position of the tile layer
	X float32
	// Y is the y position of the tile layer
	Y float32
	// XOffset is the x-offset of the tile layer
	OffSetX float32
	// YOffset is the y-offset of the tile layer
	OffSetY float32
	// Properties are the custom properties of the layer
	Properties []Property
}

// ImageLayer contains a list of its images plus all default Tiled attributes
type ImageLayer struct {
	// Name defines the name of the image layer given in the TMX XML / Tiled
	Name string
	// Source contains the original image filename
	Source string
	// Images contains the list of all image tiles
	Images []*Tile
	// Opacity is the opacity of the layer from [0,1]
	Opacity float32
	// Visible is if the layer is visible
	Visible bool
	// XOffset is the x-offset of the layer
	OffSetX float32
	// YOffset is the y-offset of the layer
	OffSetY float32
	// Properties are the custom properties of the layer
	Properties []Property
}

// ObjectLayer contains a list of its standard objects as well as a list of all its polyline objects
type ObjectLayer struct {
	// Name defines the name of the object layer given in the TMX XML / Tiled
	Name string
	// Color is the color of the object
	Color string
	// OffSetX is the parsed X offset for the object layer
	OffSetX float32
	// OffSetY is the parsed Y offset for the object layer
	OffSetY float32
	// Opacity is the opacity of the layer from [0,1]
	Opacity float32
	// Visible is if the layer is visible
	Visible bool
	// Properties are the custom properties of the layer
	Properties []Property
	// Objects contains the list of (regular) Object objects
	Objects []*Object
	// DrawOrder is whether the objects are drawn according to the order of
	// appearance (“index”) or sorted by their y-coordinate (“topdown”).
	// Defaults to “topdown”.
	DrawOrder string
}

// Object is a standard TMX object with all its default Tiled attributes
type Object struct {
	// ID is the unique ID of each object defined by Tiled
	ID uint32
	// Name defines the name of the object given in Tiled
	Name string
	// Type contains the string type which was given in Tiled
	Type string
	// X holds the X float64 coordinate of the object in the map
	X float32
	// X holds the X float64 coordinate of the object in the map
	Y float32
	// Width is the width of the object in pixels
	Width float32
	// Height is the height of the object in pixels
	Height float32
	// Properties are the custom properties of the object
	Properties []Property
	// Tiles are the tiles, if any, associated with the object
	Tiles []*Tile
	// Lines are the lines, if any, associated with the object
	Lines []TMXLine
	// Ellipses are the ellipses, if any, associated with the object
	Ellipses []TMXCircle
	// Text is the text, if any, associated with the object
	Text []TMXText
}

// TMXCircle is a circle from the tmx map
// TODO: create a tile instead using the Shape (maybe a render component?)
type TMXCircle struct {
	X, Y, Width, Height float32
}

// TMXLine is a line from the tmx map
// TODO: create a tile or render coponent instead?
type TMXLine struct {
	Lines []*engo.Line
	Type  string
}

// TMXText is text associated with a Tiled Map. It should contain all the
// information needed to render text.
// TODO: create a tile instead and have the text rendered as a texture
type TMXText struct {
	Bold       bool
	Color      string
	FontFamily string
	Halign     string
	Italic     bool
	Kerning    bool
	Size       float32
	Strikeout  bool
	Underline  bool
	Valign     string
	WordWrap   bool
	CharData   string
}

// Bounds returns the level boundaries as an engo.AABB object
func (l *Level) Bounds() engo.AABB {
	switch l.Orientation {
	case orth:
		return engo.AABB{
			Min: l.screenPoint(engo.Point{X: 0, Y: 0}),
			Max: l.screenPoint(engo.Point{X: float32(l.width), Y: float32(l.height)}),
		}
	case iso:
		xMin := l.screenPoint(engo.Point{X: 0, Y: float32(l.height)}).X + float32(l.TileWidth)/2
		xMax := l.screenPoint(engo.Point{X: float32(l.width), Y: 0}).X + float32(l.TileWidth)/2
		yMin := l.screenPoint(engo.Point{X: 0, Y: 0}).Y
		yMax := l.screenPoint(engo.Point{X: float32(l.width), Y: float32(l.height)}).Y + float32(l.TileHeight)/2
		return engo.AABB{
			Min: engo.Point{X: xMin, Y: yMin},
			Max: engo.Point{X: xMax, Y: yMax},
		}
	}
	return engo.AABB{}
}

// mapPoint returns the map point of the passed in screen point
func (l *Level) mapPoint(screenPt engo.Point) engo.Point {
	switch l.Orientation {
	case orth:
		screenPt.Multiply(engo.Point{X: 1 / float32(l.TileWidth), Y: 1 / float32(l.TileHeight)})
		return screenPt
	case iso:
		return engo.Point{
			X: (screenPt.X / float32(l.TileWidth)) + (screenPt.Y / float32(l.TileHeight)),
			Y: (screenPt.Y / float32(l.TileHeight)) - (screenPt.X / float32(l.TileWidth)),
		}
	}
	return engo.Point{X: 0, Y: 0}
}

// screenPoint returns the screen point of the passed in map point
func (l *Level) screenPoint(mapPt engo.Point) engo.Point {
	switch l.Orientation {
	case orth:
		mapPt.Multiply(engo.Point{X: float32(l.TileWidth), Y: float32(l.TileHeight)})
		return mapPt
	case iso:
		return engo.Point{
			X: (mapPt.X - mapPt.Y) * float32(l.TileWidth) / 2,
			Y: (mapPt.X + mapPt.Y) * float32(l.TileHeight) / 2,
		}
	}
	return engo.Point{X: 0, Y: 0}
}

type mapPoint struct {
	X, Y int
}

// GetTile returns a *Tile at the given point (in space / render coordinates).
func (l *Level) GetTile(pt engo.Point) *Tile {
	mp := l.mapPoint(pt)
	x := int(math.Floor(mp.X))
	y := int(math.Floor(mp.Y))
	t, ok := l.pointMap[mapPoint{X: x, Y: y}]
	if !ok {
		return nil
	}
	return t
}

// Width returns the integer width of the level
func (l *Level) Width() int {
	return l.width
}

// Height returns the integer height of the level
func (l *Level) Height() int {
	return l.height
}

// Height returns the integer height of the tile
func (t *Tile) Height() float32 {
	return t.Image.Height()
}

// Width returns the integer width of the tile
func (t *Tile) Width() float32 {
	return t.Image.Width()
}

// Texture returns the tile's Image texture
func (t *Tile) Texture() TextureID {
	return t.Image.id
}

// Close deletes the stored texture of a tile
func (t *Tile) Close() {
	t.Image.Close()
}

// View returns the tile's viewport's min and max X & Y
func (t *Tile) View() (float32, float32, float32, float32) {
	return t.Image.View()
}

// Tile represents a tile in the TMX map.
type Tile struct {
	engo.Point
	Image     *Texture
	Drawables []Drawable
	Animation *Animation
}
