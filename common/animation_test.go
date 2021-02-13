package common

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/EngoEngine/ecs"
	"github.com/EngoEngine/engo"
)

type TestDrawable struct {
	ID int
}

func (*TestDrawable) Texture() TextureID { return nil }

func (*TestDrawable) Width() float32 { return 0 }

func (*TestDrawable) Height() float32 { return 0 }

func (*TestDrawable) View() (float32, float32, float32, float32) { return 0, 0, 0, 0 }

func (*TestDrawable) Close() {}

func TestNewAnimationComponent(t *testing.T) {
	drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	ac := NewAnimationComponent(drawables, 0.1)
	if len(ac.Animations) != 0 {
		t.Errorf("NewAnimationComponent returned an already populated map of animations")
	}
	if len(ac.Drawables) != len(drawables) {
		t.Errorf("NewAnimationComponent drawables length did not match specified drawables")
	}
	if ac.Rate != 0.1 {
		t.Errorf("NewAnimationComponent Rate did not match passed in value")
	}
	if ac.CurrentAnimation != nil {
		t.Errorf("NewAnimationComponent initalized CurrentAnimation with a value")
	}
}

func TestAnimationComponentSelectAnimationByName(t *testing.T) {
	drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	ac := NewAnimationComponent(drawables, 0.1)
	firstFrames := []int{0, 1, 2}
	secondFrames := []int{1, 2, 1, 0}
	actions := []*Animation{
		{
			Name:   "ZeroOneTwo",
			Frames: firstFrames,
		},
		{
			Name:   "OneTwoOneZero",
			Frames: secondFrames,
		},
	}
	ac.AddAnimations(actions)
	ac.SelectAnimationByName("ZeroOneTwo")
	for i, frame := range ac.CurrentAnimation.Frames {
		if frame != firstFrames[i] {
			t.Errorf("Animation ZeroOneTwo was not set to current animation after SelectAnimationByName")
		}
	}
	ac.SelectAnimationByName("OneTwoOneZero")
	for i, frame := range ac.CurrentAnimation.Frames {
		if frame != secondFrames[i] {
			t.Errorf("Animation OneTwoOneZero was not set to current animation after SelectAnimationByName")
		}
	}
}

func TestAnimationComponentSelectAnimationByAction(t *testing.T) {
	drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	ac := NewAnimationComponent(drawables, 0.1)
	firstFrames := []int{0, 1, 2}
	secondFrames := []int{1, 2, 1, 0}
	actions := []*Animation{
		{
			Name:   "ZeroOneTwo",
			Frames: firstFrames,
		},
		{
			Name:   "OneTwoOneZero",
			Frames: secondFrames,
		},
	}
	ac.AddAnimations(actions)
	ac.SelectAnimationByAction(actions[0])
	for i, frame := range ac.CurrentAnimation.Frames {
		if frame != firstFrames[i] {
			t.Errorf("Animation ZeroOneTwo was not set to current animation after SelectAnimationByAction")
		}
	}
	ac.SelectAnimationByAction(actions[1])
	for i, frame := range ac.CurrentAnimation.Frames {
		if frame != secondFrames[i] {
			t.Errorf("Animation OneTwoOneZero was not set to current animation after SelectAnimationByAction")
		}
	}
}

func TestAnimationComponentAddDefaultAnimation(t *testing.T) {
	drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	ac := NewAnimationComponent(drawables, 0.1)
	def := &Animation{
		Name:   "default",
		Frames: []int{0, 1, 2},
	}
	ac.AddDefaultAnimation(def)
	if ac.def.Name != def.Name {
		t.Error("Default animation was not set by AddDefaultAnimation")
	}
}

func TestAnimationComponentIntegration(t *testing.T) {
	drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	ac := NewAnimationComponent(drawables, 0.1)
	firstFrames := []int{0, 1, 2}
	secondFrames := []int{1, 2, 1, 0}
	actions := []*Animation{
		{
			Name:   "ZeroOneTwo",
			Frames: firstFrames,
		},
		{
			Name:   "OneTwoOneZero",
			Frames: secondFrames,
		},
	}
	def := &Animation{
		Name:   "default",
		Frames: []int{2, 1, 2},
	}
	ac.AddAnimations(actions)
	ac.AddDefaultAnimation(def)
	ac.SelectAnimationByName("ZeroOneTwo")
	exp := []int{0, 1, 2, 2, 1, 2, 2, 1, 2}
	for _, e := range exp {
		if ac.CurrentAnimation == nil {
			ac.SelectAnimationByAction(def)
		}
		d := ac.Cell()
		td := d.(*TestDrawable)
		if td.ID != e {
			t.Errorf("Wrong frame from AnimationComponent.Cell()\nWanted: %v\nGot: %v", e, td.ID)
			return
		}
		ac.NextFrame()
	}
	ac.SelectAnimationByName("OneTwoOneZero")
	exp = []int{1, 2, 1, 0, 2, 1, 2}
	for _, e := range exp {
		if ac.CurrentAnimation == nil {
			ac.SelectAnimationByAction(def)
		}
		d := ac.Cell()
		td := d.(*TestDrawable)
		if td.ID != e {
			t.Errorf("Wrong frame from AnimationComponent.Cell()\nWanted: %v\nGot: %v", e, td.ID)
			return
		}
		ac.NextFrame()
	}
}

func TestAnimationComponentNextFrameNoData(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	ac := NewAnimationComponent(drawables, 0.1)
	action := &Animation{
		Name:   "action",
		Frames: []int{},
	}
	ac.AddAnimation(action)
	ac.SelectAnimationByAction(action)
	ac.NextFrame()

	if strings.HasSuffix(buf.String(), "No data for this animation") {
		t.Errorf("Wrong message recieved when NextFrame was called with no data. Got: %v", buf.String())
	}
}

type TestAnimationEntity struct {
	*ecs.BasicEntity
	*AnimationComponent
	*RenderComponent
}

type TestAnimationSystem struct {
	entities []TestAnimationEntity
	w        *ecs.World
	updates  int
}

func (t *TestAnimationSystem) New(w *ecs.World) {
	t.w = w
}

func (t *TestAnimationSystem) Add(b *ecs.BasicEntity, a *AnimationComponent, r *RenderComponent) {
	t.entities = append(t.entities, TestAnimationEntity{b, a, r})
}

func (t *TestAnimationSystem) Remove(b ecs.BasicEntity) {}

func (t *TestAnimationSystem) Update(dt float32) {
	switch t.updates {
	case 2:
		t.entities[2].SelectAnimationByName("OneZero")
	case 3:
		drawables := []Drawable{
			&TestDrawable{11},
			&TestDrawable{12},
			&TestDrawable{13},
		}
		*t.entities[3].AnimationComponent = NewAnimationComponent(drawables, 0.1)
		t.entities[3].AddDefaultAnimation(&Animation{
			Name:   "Default",
			Frames: []int{0, 1, 2},
		})
	case 7:
		t.entities[1].SelectAnimationByName("Nothing")
	}
	t.updates++
}

func (t *TestAnimationSystem) GetCurrentFrameDrawables() []int {
	ret := []int{}
	for _, e := range t.entities {
		d := e.Drawable.(*TestDrawable)
		ret = append(ret, d.ID)
	}
	return ret
}

type TestAnimation struct {
	ecs.BasicEntity
	AnimationComponent
	RenderComponent
}

type TestAnimationScene struct {
	w *ecs.World
}

func (*TestAnimationScene) Type() string { return "TestAnimationScene" }

func (*TestAnimationScene) Preload() {}

func (t *TestAnimationScene) Setup(u engo.Updater) {
	t.w = u.(*ecs.World)

	t.w.AddSystem(&AnimationSystem{})
	t.w.AddSystem(&TestAnimationSystem{})

	anim0 := TestAnimation{BasicEntity: ecs.NewBasic()}
	anim0Drawables := []Drawable{
		&TestDrawable{0},
		&TestDrawable{1},
		&TestDrawable{2},
	}
	anim0Actions := []*Animation{
		{
			Name:   "ZeroOneTwo",
			Frames: []int{0, 1, 2},
		},
		{
			Name:   "OneTwoOneTwo",
			Frames: []int{1, 2, 1, 2},
		},
	}
	anim0DefaultAnimation := &Animation{
		Name:   "Default",
		Frames: []int{0, 1, 0},
	}
	anim0.RenderComponent.Drawable = anim0Drawables[0]
	anim0.AnimationComponent = NewAnimationComponent(anim0Drawables, 2)
	anim0.AnimationComponent.AddAnimations(anim0Actions)
	anim0.AnimationComponent.AddDefaultAnimation(anim0DefaultAnimation)
	for _, system := range t.w.Systems() {
		switch sys := system.(type) {
		case *AnimationSystem:
			sys.Add(&anim0.BasicEntity, &anim0.AnimationComponent, &anim0.RenderComponent)
		case *TestAnimationSystem:
			sys.Add(&anim0.BasicEntity, &anim0.AnimationComponent, &anim0.RenderComponent)
		}
	}

	anim1 := TestAnimation{BasicEntity: ecs.NewBasic()}
	anim1Drawables := []Drawable{
		&TestDrawable{3},
		&TestDrawable{4},
		&TestDrawable{5},
		&TestDrawable{6},
	}
	anim1Actions := []*Animation{
		{
			Name:   "ZeroOneTwoThree",
			Frames: []int{0, 1, 2, 3},
		},
		{
			Name:   "Nothing",
			Frames: []int{},
		},
	}
	anim1DefaultAnimation := &Animation{
		Name:   "Default",
		Frames: []int{0, 3, 1, 3},
	}
	anim1.RenderComponent.Drawable = anim1Drawables[0]
	anim1.AnimationComponent = NewAnimationComponent(anim1Drawables, 1)
	anim1.AnimationComponent.AddAnimations(anim1Actions)
	anim1.AnimationComponent.AddDefaultAnimation(anim1DefaultAnimation)
	for _, system := range t.w.Systems() {
		switch sys := system.(type) {
		case *AnimationSystem:
			sys.Add(&anim1.BasicEntity, &anim1.AnimationComponent, &anim1.RenderComponent)
		case *TestAnimationSystem:
			sys.Add(&anim1.BasicEntity, &anim1.AnimationComponent, &anim1.RenderComponent)
		}
	}

	anim2 := TestAnimation{BasicEntity: ecs.NewBasic()}
	anim2Drawables := []Drawable{
		&TestDrawable{7},
		&TestDrawable{8},
	}
	anim2Actions := []*Animation{
		{
			Name:   "OneZero",
			Frames: []int{1, 0},
		},
	}
	anim2.RenderComponent.Drawable = anim2Drawables[0]
	anim2.AnimationComponent = NewAnimationComponent(anim2Drawables, 1)
	anim2.AnimationComponent.AddAnimations(anim2Actions)
	for _, system := range t.w.Systems() {
		switch sys := system.(type) {
		case *AnimationSystem:
			sys.Add(&anim2.BasicEntity, &anim2.AnimationComponent, &anim2.RenderComponent)
		case *TestAnimationSystem:
			sys.Add(&anim2.BasicEntity, &anim2.AnimationComponent, &anim2.RenderComponent)
		}
	}

	anim3 := TestAnimation{BasicEntity: ecs.NewBasic()}
	anim3Drawables := []Drawable{
		&TestDrawable{10},
	}
	anim3.RenderComponent.Drawable = anim3Drawables[0]
	for _, system := range t.w.Systems() {
		switch sys := system.(type) {
		case *AnimationSystem:
			sys.Add(&anim3.BasicEntity, &anim3.AnimationComponent, &anim3.RenderComponent)
		case *TestAnimationSystem:
			sys.Add(&anim3.BasicEntity, &anim3.AnimationComponent, &anim3.RenderComponent)
		}
	}
}

func TestAnimationSystemIntegration(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	s := TestAnimationScene{}
	engo.Run(engo.RunOptions{
		HeadlessMode: true,
		NoRun:        true,
	}, &s)
	testSystem := &TestAnimationSystem{}
	for _, sys := range s.w.Systems() {
		if system, ok := sys.(*TestAnimationSystem); ok {
			testSystem = system
		}
	}
	// initial state
	exp := []int{0, 3, 7, 10}
	res := testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect initial frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 3, 7, 10}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect first frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 6, 7, 10}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect second frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 4, 7, 10}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect third frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{1, 6, 8, 10}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect fourth frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{1, 3, 7, 11}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect fifth frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 6, 7, 12}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect sixth frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 4, 7, 13}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect seventh frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 6, 7, 11}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect eighth frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	s.w.Update(1)
	exp = []int{0, 3, 7, 12}
	res = testSystem.GetCurrentFrameDrawables()
	for i := 0; i < 4; i++ {
		if exp[i] != res[i] {
			t.Errorf("Incorrect eighth frame!\nWanted: %v\nGot: %v\nIndex: %v", exp[i], res[i], i)
			return
		}
	}
	if !strings.HasSuffix(buf.String(), "No frame data for this animation\n") {
		t.Errorf("No notice printed to log when animation has no frame data\nLog was: %v", buf.String())
		return
	}
}
