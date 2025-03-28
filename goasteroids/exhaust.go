package goasteroids

import (
	"go-asteroids/assets"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	exhaustSpawnOffset = -50.0
)

type Exhaust struct {
	position Vector
	rotation float64
	sprite   *ebiten.Image
}

func NewExhaust(pos Vector, rotation float64) *Exhaust {
	/* set the sprite */
	sprite := assets.ExhaustSprite

	/* position x and y coords from center of sprite */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH

	/* create a exhaust obj */
	return &Exhaust{
		position: pos,
		rotation: rotation,
		sprite:   sprite,
	}
}

func (e *Exhaust) Update() {
	speed := maxAcceleration / float64(ebiten.TPS())
	e.position.X += math.Sin(e.rotation) * speed
	e.position.Y += math.Cos(e.rotation) * -speed
}

func (e *Exhaust) Draw(screen *ebiten.Image) {
	bounds := e.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(e.rotation)
	op.GeoM.Translate(halfW, halfH)
	op.GeoM.Translate(e.position.X, e.position.Y)

	screen.DrawImage(e.sprite, op)

}
