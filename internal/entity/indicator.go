package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// Indicator is a static, half-transparent HUD icon. Lives, shields, and
// hyperspace all render identically — only the sprite differs.
type Indicator struct {
	position engine.Vector
	sprite   *ebiten.Image
}

func NewLifeIndicator(pos engine.Vector) *Indicator {
	return &Indicator{position: pos, sprite: assets.LifeIndicator}
}

func NewShieldIndicator(pos engine.Vector) *Indicator {
	return &Indicator{position: pos, sprite: assets.ShieldIndicator}
}

func NewHyperspaceIndicator(pos engine.Vector) *Indicator {
	return &Indicator{position: pos, sprite: assets.HyperspaceIndicator}
}

func (i *Indicator) Update() {}

func (i *Indicator) Draw(screen *ebiten.Image) {
	bounds := i.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(halfW, halfH)
	cm := colorm.ColorM{}
	cm.Scale(1.0, 1.0, 1.0, 0.5)
	op.GeoM.Translate(i.position.X, i.position.Y)

	colorm.DrawImage(screen, i.sprite, cm, op)
}
