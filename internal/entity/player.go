package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

const (
	dyingAnimationAmount = 50 * time.Millisecond
	numberOfLives        = 3
)

type Player struct {
	scene     Scene
	Sprite    *ebiten.Image
	Rotation  float64
	Position  engine.Vector
	PlayerObj *resolv.Circle

	motion motion
	weapon weapon

	IsShielded       bool
	shieldTimer      *engine.Timer
	ShieldsRemaining int

	IsDying        bool
	IsDead         bool
	DyingTimer     *engine.Timer
	DyingCounter   int
	LivesRemaining int

	hyperspaceTimer *engine.Timer
}

func NewPlayer(scene Scene) *Player {
	sprite := assets.PlayerSprite

	/* center player on screen */
	pos := engine.CenterSprite(engine.Vector{
		X: engine.ScreenWidth / 2,
		Y: engine.ScreenHeight / 2,
	}, sprite)

	p := &Player{
		scene:            scene,
		Sprite:           sprite,
		Position:         pos,
		PlayerObj:        engine.CircleFor(sprite, pos),
		weapon:           newWeapon(),
		DyingTimer:       engine.NewTimer(dyingAnimationAmount),
		LivesRemaining:   numberOfLives,
		ShieldsRemaining: numberOfShields,
	}

	p.PlayerObj.Tags().Set(engine.TagPlayer)

	return p
}

func (p *Player) Draw(screen *ebiten.Image) {
	engine.DrawSprite(screen, p.Sprite, p.Position, p.Rotation)
}

func (p *Player) Update() {
	p.isPlayerDead()

	p.rotate()
	p.move()

	p.useShield()
	p.fireLasers()
	p.hyperspace()
}

func (p *Player) isPlayerDead() {
	if p.IsDead {
		p.scene.SetPlayerDead()
	}
}

func (p *Player) spawnPoint(distance float64) engine.Vector {
	bounds := p.Sprite.Bounds()

	return engine.Vector{
		X: p.Position.X + float64(bounds.Dx())/2 + math.Sin(p.Rotation)*distance,
		Y: p.Position.Y + float64(bounds.Dy())/2 - math.Cos(p.Rotation)*distance,
	}
}
