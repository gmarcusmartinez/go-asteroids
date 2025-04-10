package goasteroids

import (
	"go-asteroids/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type ShieldIndicator struct {
	position Vector
	sprite   *ebiten.Image
}

func NewShieldIndicator(pos Vector) *ShieldIndicator {
	return &ShieldIndicator{
		position: pos,
		sprite:   assets.ShieldIndicator,
	}
}

func (si *ShieldIndicator) Update() {}

func (si *ShieldIndicator) Draw(screen *ebiten.Image) {
	bounds := si.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(halfW, halfH)
	cm := colorm.ColorM{}
	cm.Scale(1.0, 1.0, 1.0, 0.5)
	op.GeoM.Translate(si.position.X, si.position.Y)

	colorm.DrawImage(screen, si.sprite, cm, op)

}
