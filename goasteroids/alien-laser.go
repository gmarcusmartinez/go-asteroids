package goasteroids

import (
	"go-asteroids/assets"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	alienLaserSpeedPerSecond = 1000.0
)

type AlienLaser struct {
	position Vector
	rotation float64
	sprite   *ebiten.Image
	laserObj *resolv.ConvexPolygon
}

func NewAlienLaser(pos Vector, rotation float64) *AlienLaser {
	/* set the sprite */
	sprite := assets.AlienLaserSprite

	/* position x and y coords from center of sprite */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH

	/* create an alien laser obj */
	al := &AlienLaser{
		position: pos,
		rotation: rotation,
		sprite:   sprite,
		laserObj: resolv.NewRectangle(pos.X, pos.Y, float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy())),
	}

	/* set the position of the collision obj */
	al.laserObj.SetPosition(pos.X, pos.Y)
	al.laserObj.Tags().Set(TagLaser)

	return al

}

func (al *AlienLaser) Update() {
	speed := alienLaserSpeedPerSecond / float64(ebiten.TPS())

	al.position.X += math.Sin(al.rotation) * speed
	al.position.Y += math.Cos(al.rotation) * -speed

	al.laserObj.SetPosition(al.position.X, al.position.Y)
}

func (al *AlienLaser) Draw(screen *ebiten.Image) {
	bounds := al.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(al.rotation)
	op.GeoM.Translate(al.position.X, al.position.Y)

	screen.DrawImage(al.sprite, op)

}
