package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	laserSpeedPerSecond = 1000.0
)

type Laser struct {
	Position engine.Vector
	rotation float64
	sprite   *ebiten.Image
	Obj      *resolv.ConvexPolygon
}

func NewLaser(pos engine.Vector, rotation float64, index int) *Laser {
	/* set the sprite */
	sprite := assets.LaserSprite

	/* shift to top-left so the sprite is centered on pos */
	pos = engine.CenterSprite(pos, sprite)

	/* create a laser obj */
	l := &Laser{
		Position: pos,
		rotation: rotation,
		sprite:   sprite,
		Obj:      engine.RectangleFor(sprite, pos),
	}

	/* set the position of the collision obj */
	l.Obj.SetPosition(pos.X, pos.Y)
	l.Obj.SetData(&engine.ObjectData{Index: index})
	l.Obj.Tags().Set(engine.TagLaser)

	return l

}

func (l *Laser) Update() {
	speed := laserSpeedPerSecond / float64(ebiten.TPS())

	dx := math.Sin(l.rotation) * speed
	dy := math.Cos(l.rotation) * -speed

	l.Position.X += dx
	l.Position.Y += dy

	l.Obj.SetPosition(l.Position.X, l.Position.Y)
}

func (l *Laser) Draw(screen *ebiten.Image) {
	engine.DrawSprite(screen, l.sprite, l.Position, l.rotation)
}
