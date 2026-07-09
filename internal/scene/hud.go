package scene

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"go-asteroids/internal/entity"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

/* HUD layout: rows of half-transparent icons in the top-left corner */
const indicatorSpacing = 50.0

var (
	lifeRowOrigin      = engine.Vector{X: 20, Y: 20}
	shieldRowOrigin    = engine.Vector{X: 45, Y: 60}
	hyperspacePosition = engine.Vector{X: 37, Y: 95}
)

func drawHUD(screen *ebiten.Image, p *entity.Player) {
	for i := range p.LivesRemaining {
		drawIndicator(screen, assets.LifeIndicator, engine.Vector{
			X: lifeRowOrigin.X + float64(i)*indicatorSpacing,
			Y: lifeRowOrigin.Y,
		})
	}

	for i := range p.ShieldsRemaining {
		drawIndicator(screen, assets.ShieldIndicator, engine.Vector{
			X: shieldRowOrigin.X + float64(i)*indicatorSpacing,
			Y: shieldRowOrigin.Y,
		})
	}

	if p.HyperspaceReady() {
		drawIndicator(screen, assets.HyperspaceIndicator, hyperspacePosition)
	}
}

func drawIndicator(screen, sprite *ebiten.Image, pos engine.Vector) {
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(halfW, halfH)
	cm := colorm.ColorM{}
	cm.Scale(1.0, 1.0, 1.0, 0.5)
	op.GeoM.Translate(pos.X, pos.Y)

	colorm.DrawImage(screen, sprite, cm, op)
}
