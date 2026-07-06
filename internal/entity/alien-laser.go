package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	alienLaserSpeedPerSecond = 1000.0
)

type AlienLaser struct {
	Position engine.Vector
	rotation float64
	sprite   *ebiten.Image
	LaserObj *resolv.ConvexPolygon
}

func NewAlienLaser(pos engine.Vector, rotation float64) *AlienLaser {
	/* set the sprite */
	sprite := assets.AlienLaserSprite

	/* shift to top-left so the sprite is centered on pos */
	pos = engine.CenterSprite(pos, sprite)

	/* create an alien laser obj */
	al := &AlienLaser{
		Position: pos,
		rotation: rotation,
		sprite:   sprite,
		LaserObj: resolv.NewRectangle(pos.X, pos.Y, float64(sprite.Bounds().Dx()), float64(sprite.Bounds().Dy())),
	}

	/* set the position of the collision obj */
	al.LaserObj.SetPosition(pos.X, pos.Y)
	al.LaserObj.Tags().Set(engine.TagLaser)

	return al

}

func (al *AlienLaser) Update() {
	speed := alienLaserSpeedPerSecond / float64(ebiten.TPS())

	al.Position.X += math.Sin(al.rotation) * speed
	al.Position.Y += math.Cos(al.rotation) * -speed

	al.LaserObj.SetPosition(al.Position.X, al.Position.Y)
}

func (al *AlienLaser) Draw(screen *ebiten.Image) {
	bounds := al.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(al.rotation)
	op.GeoM.Translate(al.Position.X, al.Position.Y)

	screen.DrawImage(al.sprite, op)

}
