package goasteroids

import (
	"go-asteroids/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

type ShieldIndicator struct {
	position Vector
	rotation float64
	sprite   *ebiten.Image
}

func NewShieldIndicator(pos Vector) *ShieldIndicator {
	sprite := assets.ShieldIndicator

	return &ShieldIndicator{
		position: pos,
		rotation: 0,
		sprite:   sprite,
	}
}

func (s *ShieldIndicator) Update() {}

func (s *ShieldIndicator) Draw(screen *ebiten.Image) {
	bounds := s.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(halfW, halfH)
	cm := colorm.ColorM{}
	cm.Scale(1.0, 1.0, 1.0, 0.5)
	op.GeoM.Translate(s.position.X, s.position.Y)

	colorm.DrawImage(screen, s.sprite, cm, op)

}
