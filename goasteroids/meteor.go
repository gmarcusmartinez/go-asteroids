package goasteroids

import (
	"go-asteroids/assets"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	roataionSpeedMin                    = -0.02
	roataionSpeedMax                    = 0.02
	numberOfSmallMeteorsFromLargeMetoer = 4
)

type Meteor struct {
	game          *GameScene
	position      Vector
	rotation      float64
	movement      Vector
	angle         float64
	rotationSpeed float64
	sprite        *ebiten.Image
}

func NewMeteor(baseVelocity float64, g *GameScene, index int) *Meteor {
	/* target the center of the screen */
	target := Vector{
		X: ScreenWidth / 2,
		Y: ScreenHeight / 2,
	}

	/* pick a random angle */
	angle := rand.Float64() * 2 * math.Pi

	/* spawn distance from center */
	r := ScreenWidth/2.0 + 500

	/* create the position vector */
	pos := Vector{
		X: target.X + math.Cos(angle)*r,
		Y: target.Y + math.Sin(angle)*r,
	}

	/* give meteor random velocity */
	velocity := baseVelocity + rand.Float64()*1.5

	/* create and normalize direction vector */
	direction := Vector{
		X: target.X - pos.X,
		Y: target.Y - pos.Y,
	}
	normalizedDirection := direction.Normalize()

	/* create movement vector */
	movement := Vector{
		X: normalizedDirection.X * velocity,
		Y: normalizedDirection.Y * velocity,
	}

	/* assign a sprite to the meteor */
	sprite := assets.MeteorSprites[rand.Intn(len(assets.MeteorSprites))]

	/* create a meteor object and return */
	m := &Meteor{
		game:          g,
		position:      pos,
		angle:         angle,
		movement:      movement,
		rotationSpeed: roataionSpeedMin + rand.Float64()*(roataionSpeedMax-roataionSpeedMin),
		sprite:        sprite,
	}

	return m
}

func (m *Meteor) Draw(screen *ebiten.Image) {
	bounds := m.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(m.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(m.position.X, m.position.Y)
	screen.DrawImage(m.sprite, op)

}

func (m *Meteor) Update() {
	dx := m.movement.X
	dy := m.movement.Y

	m.position.X += dx
	m.position.Y += dy
	m.rotation += m.rotationSpeed

	m.keepOnScreen()
}

func (m *Meteor) keepOnScreen() {
	if m.position.X >= float64(ScreenWidth) {
		m.position.X = 0
	}

	if m.position.X < 0 {
		m.position.X = ScreenWidth
	}

	if m.position.Y >= float64(ScreenHeight) {
		m.position.Y = 0
	}

	if m.position.Y < 0 {
		m.position.Y = ScreenHeight
	}
}
