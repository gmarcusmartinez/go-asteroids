package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Exhaust struct {
	position engine.Vector
	rotation float64
	sprite   *ebiten.Image
}

func NewExhaust(pos engine.Vector, rotation float64) *Exhaust {
	/* set the sprite */
	sprite := assets.ExhaustSprite

	/* shift to top-left so the sprite is centered on pos */
	pos = engine.CenterSprite(pos, sprite)

	/* create a exhaust obj */
	return &Exhaust{
		position: pos,
		rotation: rotation,
		sprite:   sprite,
	}
}

func (e *Exhaust) Update() {
	speed := engine.MaxAcceleration / float64(ebiten.TPS())
	e.position.X += math.Sin(e.rotation) * speed
	e.position.Y += math.Cos(e.rotation) * -speed
}

func (e *Exhaust) Draw(screen *ebiten.Image) {
	engine.DrawSprite(screen, e.sprite, e.position, e.rotation)
}
