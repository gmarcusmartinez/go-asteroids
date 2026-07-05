package goasteroids

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

type Alien struct {
	game          *GameScene
	sprite        *ebiten.Image
	alienObj      *resolv.Circle
	position      engine.Vector
	angle         float64
	movement      engine.Vector
	isIntelligent bool
}

func NewAlien(baseVelocity float64, g *GameScene) *Alien {
	var alien Alien

	alienType := rand.Intn(3)

	sprite := assets.AlienSprites[rand.Intn(len(assets.AlienSprites))]

	switch alienType {
	case 0:
		/* comes in from right shoots random */
		x := float64(engine.ScreenWidth + 100)
		y := float64(rand.Intn(engine.ScreenHeight-100) + 100)

		target := engine.Vector{X: 0, Y: y}
		pos := engine.Vector{X: x, Y: y}
		velocity := baseVelocity + rand.Float64()*2.5
		movement := engine.Vector{X: target.X - velocity, Y: 0}

		alien = Alien{
			game:          g,
			sprite:        sprite,
			position:      pos,
			alienObj:      resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2)),
			movement:      movement,
			isIntelligent: false,
		}

		alien.alienObj.SetPosition(pos.X, pos.Y)
	case 1:
		/* comes in from left shoots random */
		x := float64(-100)
		y := float64(rand.Intn(engine.ScreenHeight-100) + 100)

		target := engine.Vector{X: 0, Y: y}
		pos := engine.Vector{X: x, Y: y}
		velocity := baseVelocity + rand.Float64()*2.5
		movement := engine.Vector{X: target.X + velocity, Y: 0}

		alien = Alien{
			game:          g,
			sprite:        sprite,
			position:      pos,
			alienObj:      resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2)),
			movement:      movement,
			isIntelligent: false,
		}

		alien.alienObj.SetPosition(pos.X, pos.Y)
	case 2:
		/* Intelligent Alien */
		middle := engine.Vector{
			X: engine.ScreenWidth / 2,
			Y: engine.ScreenHeight / 2,
		}

		angle := rand.Float64() * 2 * math.Pi
		r := engine.ScreenHeight / 2.0

		pos := engine.Vector{
			X: middle.X + math.Cos(angle)*r,
			Y: middle.Y + math.Sin(angle)*r,
		}

		velocity := baseVelocity + rand.Float64()*1.5
		target := g.player.position

		direction := engine.Vector{
			X: target.X - pos.X,
			Y: target.Y / -pos.Y,
		}

		normalizedDirection := direction.Normalize()
		movement := engine.Vector{
			X: normalizedDirection.X * velocity,
			Y: normalizedDirection.Y * velocity,
		}

		alien = Alien{
			game:          g,
			sprite:        sprite,
			position:      pos,
			alienObj:      resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2)),
			angle:         angle,
			movement:      movement,
			isIntelligent: true,
		}

		alien.alienObj.SetPosition(pos.X, pos.Y)
	}

	alien.alienObj.Tags().Set(engine.TagAlien)
	return &alien
}

func (a *Alien) Update() {
	dx := a.movement.X
	dy := a.movement.Y

	a.position.X += dx
	a.position.Y += dy

	a.alienObj.SetPosition(a.position.X, a.position.Y)
}

func (a *Alien) Draw(screen *ebiten.Image) {
	bounds := a.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Translate(a.position.X, a.position.Y)
	screen.DrawImage(a.sprite, op)
}
