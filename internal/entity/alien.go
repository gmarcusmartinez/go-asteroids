package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

type Alien struct {
	Sprite        *ebiten.Image
	Obj           *resolv.Circle
	Position      engine.Vector
	angle         float64
	movement      engine.Vector
	IsIntelligent bool
}

func NewAlien(baseVelocity float64, playerPos engine.Vector) *Alien {
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
			Sprite:        sprite,
			Position:      pos,
			Obj:           engine.CircleFor(sprite, pos),
			movement:      movement,
			IsIntelligent: false,
		}

		alien.Obj.SetPosition(pos.X, pos.Y)
	case 1:
		/* comes in from left shoots random */
		x := float64(-100)
		y := float64(rand.Intn(engine.ScreenHeight-100) + 100)

		target := engine.Vector{X: 0, Y: y}
		pos := engine.Vector{X: x, Y: y}
		velocity := baseVelocity + rand.Float64()*2.5
		movement := engine.Vector{X: target.X + velocity, Y: 0}

		alien = Alien{
			Sprite:        sprite,
			Position:      pos,
			Obj:           engine.CircleFor(sprite, pos),
			movement:      movement,
			IsIntelligent: false,
		}

		alien.Obj.SetPosition(pos.X, pos.Y)
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
		target := playerPos

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
			Sprite:        sprite,
			Position:      pos,
			Obj:           engine.CircleFor(sprite, pos),
			angle:         angle,
			movement:      movement,
			IsIntelligent: true,
		}

		alien.Obj.SetPosition(pos.X, pos.Y)
	}

	alien.Obj.Tags().Set(engine.TagAlien)
	return &alien
}

func (a *Alien) Update() {
	dx := a.movement.X
	dy := a.movement.Y

	a.Position.X += dx
	a.Position.Y += dy

	a.Obj.SetPosition(a.Position.X, a.Position.Y)
}

func (a *Alien) Draw(screen *ebiten.Image) {
	bounds := a.Sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Translate(a.Position.X, a.Position.Y)
	screen.DrawImage(a.Sprite, op)
}
