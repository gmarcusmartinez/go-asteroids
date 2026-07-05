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

	/* position x and y coords from center of sprite */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH

	/* create a laser obj */
	l := &Laser{
		Position: pos,
		rotation: rotation,
		sprite:   sprite,
		Obj:      resolv.NewRectangle(pos.X, pos.Y, float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy())),
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
	bounds := l.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(l.rotation)
	op.GeoM.Translate(halfW, halfH)
	op.GeoM.Translate(l.Position.X, l.Position.Y)

	screen.DrawImage(l.sprite, op)

}
