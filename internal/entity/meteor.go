package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	rotationSpeedMin = -0.02
	rotationSpeedMax = 0.02
)

type Meteor struct {
	Position      engine.Vector
	rotation      float64
	Movement      engine.Vector
	angle         float64
	rotationSpeed float64
	Sprite        *ebiten.Image
	Obj           *resolv.Circle
}

func NewMeteor(baseVelocity float64, index int) *Meteor {
	/* target the center of the screen */
	target := engine.Vector{
		X: engine.ScreenWidth / 2,
		Y: engine.ScreenHeight / 2,
	}

	/* pick a random angle */
	angle := rand.Float64() * 2 * math.Pi

	/* spawn distance from center */
	r := engine.ScreenWidth/2.0 + 500

	/* create the position vector */
	pos := engine.Vector{
		X: target.X + math.Cos(angle)*r,
		Y: target.Y + math.Sin(angle)*r,
	}

	/* give meteor random velocity */
	velocity := baseVelocity + rand.Float64()*1.5

	/* create and normalize direction vector */
	direction := engine.Vector{
		X: target.X - pos.X,
		Y: target.Y - pos.Y,
	}
	normalizedDirection := direction.Normalize()

	/* create movement vector */
	movement := engine.Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	/* assign a sprite to the meteor */
	sprite := assets.MeteorSprites[rand.Intn(len(assets.MeteorSprites))]

	/* create the collision object */
	meteorObj := resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2))

	/* create a meteor object and return */
	m := &Meteor{
		Position:      pos,
		angle:         angle,
		Movement:      movement,
		rotationSpeed: rotationSpeedMin + rand.Float64()*(rotationSpeedMax-rotationSpeedMin),
		Sprite:        sprite,
		Obj:           meteorObj,
	}

	m.Obj.SetPosition(pos.X, pos.Y)
	m.Obj.Tags().Set(engine.TagMeteor | engine.TagLarge)
	m.Obj.SetData(&engine.ObjectData{Index: index})

	return m
}

func NewSmallMeteor(baseVelocity float64, index int) *Meteor {
	/* target the center of the screen */
	target := engine.Vector{
		X: engine.ScreenWidth / 2,
		Y: engine.ScreenHeight / 2,
	}

	/* pick a random angle */
	angle := rand.Float64() * 2 * math.Pi

	/* spawn distance from center */
	r := engine.ScreenWidth/2.0 + 500

	/* create the position vector */
	pos := engine.Vector{
		X: target.X + math.Cos(angle)*r,
		Y: target.Y + math.Sin(angle)*r,
	}

	/* give meteor random velocity */
	velocity := baseVelocity + rand.Float64()*1.5

	/* create and normalize direction vector */
	direction := engine.Vector{
		X: target.X - pos.X,
		Y: target.Y - pos.Y,
	}
	normalizedDirection := direction.Normalize()

	/* create movement vector */
	movement := engine.Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	/* assign a sprite to the meteor */
	sprite := assets.MeteorSpritesSmall[rand.Intn(len(assets.MeteorSpritesSmall))]

	/* create the collision object */
	meteorObj := resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2))

	/* create a meteor object and return */
	m := &Meteor{
		Position:      pos,
		angle:         angle,
		Movement:      movement,
		rotationSpeed: rotationSpeedMin + rand.Float64()*(rotationSpeedMax-rotationSpeedMin),
		Sprite:        sprite,
		Obj:           meteorObj,
	}

	m.Obj.SetPosition(pos.X, pos.Y)
	m.Obj.Tags().Set(engine.TagMeteor | engine.TagSmall)
	m.Obj.SetData(&engine.ObjectData{Index: index})

	return m
}

func (m *Meteor) Draw(screen *ebiten.Image) {
	engine.DrawSprite(screen, m.Sprite, m.Position, m.rotation)
}

func (m *Meteor) Update() {
	dx := m.Movement.X
	dy := m.Movement.Y

	m.Position.X += dx
	m.Position.Y += dy
	m.rotation += m.rotationSpeed

	m.keepOnScreen()

	/* update the collision object */
	m.Obj.SetPosition(m.Position.X, m.Position.Y)

}

func (m *Meteor) keepOnScreen() {
	if m.Position.X >= float64(engine.ScreenWidth) {
		m.Position.X = 0
		m.Obj.SetPosition(0, m.Position.Y)
	}

	if m.Position.X < 0 {
		m.Position.X = engine.ScreenWidth
		m.Obj.SetPosition(engine.ScreenWidth, m.Position.Y)
	}

	if m.Position.Y >= float64(engine.ScreenHeight) {
		m.Position.Y = 0
		m.Obj.SetPosition(m.Position.X, 0)
	}

	if m.Position.Y < 0 {
		m.Position.Y = engine.ScreenHeight
		m.Obj.SetPosition(m.Position.X, engine.ScreenHeight)

	}
}
