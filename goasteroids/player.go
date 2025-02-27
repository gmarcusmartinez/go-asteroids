package goasteroids

import (
	"go-asteroids/assets"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	rotationPerSecond = math.Pi
	maxAcceleration   = 8.0
	ScreenWidth       = 1280
	ScreenHeight      = 720
)

var currentAcceleration float64

type Player struct {
	game           *Game
	sprite         *ebiten.Image
	rotation       float64
	position       Vector
	playerVelocity float64
}

func NewPlayer(game *Game) *Player {
	sprite := assets.PlayerSprite

	/* center player on screen */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - halfW,
		Y: ScreenHeight/2 - halfH,
	}

	p := &Player{
		sprite:   sprite,
		game:     game,
		position: pos,
	}

	return p
}

func (p *Player) Draw(screen *ebiten.Image) {
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(p.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

func (p *Player) Update() {
	speed := rotationPerSecond / float64(ebiten.TPS())

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}

	p.accelerate()

}

func (p *Player) accelerate() {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.keepOnScreen()

		if currentAcceleration < maxAcceleration {
			currentAcceleration = p.playerVelocity + 4
		}

		if currentAcceleration >= 8 {
			currentAcceleration = 8
		}

		p.playerVelocity = currentAcceleration

		/* move in the direction we are pointing */
		dx := math.Sin(p.rotation) * currentAcceleration
		dy := math.Cos(p.rotation) * -currentAcceleration

		/* move player */
		p.position.X += dx
		p.position.Y += dy

	}
}

func (p *Player) keepOnScreen() {
	if p.position.X >= float64(ScreenWidth) {
		p.position.X = 0
	}

	if p.position.X < 0 {
		p.position.X = ScreenWidth
	}

	if p.position.Y >= float64(ScreenHeight) {
		p.position.Y = 0
	}

	if p.position.Y < 0 {
		p.position.Y = ScreenHeight
	}
}
