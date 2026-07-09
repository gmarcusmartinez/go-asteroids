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
	sprite := assets.AlienSprites[rand.Intn(len(assets.AlienSprites))]

	var pos, movement engine.Vector
	var angle float64
	var intelligent bool

	fromRight := float64(engine.ScreenWidth + 100)
	fromLeft := float64(-100)

	switch rand.Intn(3) {
	case 0:
		pos, movement = edgeSpawn(fromRight, baseVelocity, -1)
	case 1:
		pos, movement = edgeSpawn(fromLeft, baseVelocity, +1)
	case 2:
		pos, angle, movement = intelligentSpawn(baseVelocity, playerPos)
		intelligent = true
	}

	alien := Alien{
		Sprite:        sprite,
		Position:      pos,
		Obj:           engine.CircleFor(sprite, pos),
		angle:         angle,
		movement:      movement,
		IsIntelligent: intelligent,
	}

	alien.Obj.SetPosition(pos.X, pos.Y)
	alien.Obj.Tags().Set(engine.TagAlien)

	return &alien
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

func (a *Alien) Update() {
	dx := a.movement.X
	dy := a.movement.Y

	a.Position.X += dx
	a.Position.Y += dy

	a.Obj.SetPosition(a.Position.X, a.Position.Y)
}

func edgeSpawn(x, baseVelocity, dir float64) (pos, movement engine.Vector) {
	y := float64(rand.Intn(engine.ScreenHeight-100) + 100)

	velocity := baseVelocity + rand.Float64()*2.5
	pos = engine.Vector{X: x, Y: y}
	movement = engine.Vector{X: dir * velocity}

	return pos, movement
}

// spawns an alien on a circle around screen center and aims its movement toward the player.
func intelligentSpawn(baseVelocity float64, playerPos engine.Vector) (pos engine.Vector, angle float64, movement engine.Vector) {
	middle := engine.Vector{X: engine.ScreenWidth / 2, Y: engine.ScreenHeight / 2}
	angle = rand.Float64() * 2 * math.Pi
	r := engine.ScreenHeight / 2.0

	pos = engine.Vector{
		X: middle.X + math.Cos(angle)*r,
		Y: middle.Y + math.Sin(angle)*r,
	}

	velocity := baseVelocity + rand.Float64()*1.5
	direction := engine.Vector{
		X: playerPos.X - pos.X,
		Y: playerPos.Y / -pos.Y,
	}.Normalize()

	movement = engine.Vector{X: direction.X * velocity, Y: direction.Y * velocity}
	return pos, angle, movement
}
