package entity

import (
	"go-asteroids/internal/engine"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	rotationPerSecond  = math.Pi
	exhaustSpawnOffset = -50.0
	driftTime          = time.Second * 30
)

type motion struct {
	acceleration float64
	velocity     float64
	driftAngle   float64
	driftTimer   *engine.Timer
}

func (p *Player) rotate() {
	speed := rotationPerSecond / float64(ebiten.TPS())

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.Rotation -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.Rotation += speed
	}
}

/* move applies thrust, drift, and reverse, then syncs the collision object */
func (p *Player) move() {
	p.accelerate()
	p.isDoneAccelerating()
	p.drift()
	p.reverse()
	p.isDoneReversing()
	p.updateExhaustSprite()

	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)
}

func (p *Player) accelerate() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) {
		return
	}

	p.motion.driftTimer = nil

	p.keepOnScreen()

	if p.motion.acceleration < engine.MaxAcceleration {
		p.motion.acceleration = p.motion.velocity + 4
	}

	if p.motion.acceleration >= engine.MaxAcceleration {
		p.motion.acceleration = engine.MaxAcceleration
	}

	p.motion.velocity = p.motion.acceleration

	/* move in the direction we are pointing */
	dx := math.Sin(p.Rotation) * p.motion.acceleration
	dy := math.Cos(p.Rotation) * -p.motion.acceleration

	p.showExhaust()

	/* move player */
	p.Position.X += dx
	p.Position.Y += dy

	/* play thrust sound */
	p.scene.PlayThrust()
}

func (p *Player) isDoneAccelerating() {
	if !inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		return
	}

	p.scene.PauseThrust()

	/* figure out velocity */
	if p.motion.velocity < p.motion.acceleration*10 {
		p.motion.velocity = p.motion.acceleration*10 - 5.0
	}

	if p.motion.velocity < 0 {
		p.motion.velocity = 0
	}

	p.motion.acceleration = 0

	/* drift along the current heading until the timer runs out */
	p.motion.driftTimer = engine.NewTimer(driftTime)
	p.motion.driftAngle = p.Rotation
}

func (p *Player) drift() {
	if p.motion.driftTimer == nil {
		return
	}

	p.keepOnScreen()
	p.motion.driftTimer.Update()

	decelerationSpeed := p.motion.velocity / float64(ebiten.TPS()) * 4

	p.Position.X += math.Sin(p.motion.driftAngle) * decelerationSpeed
	p.Position.Y += math.Cos(p.motion.driftAngle) * -decelerationSpeed
	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)

	if p.motion.driftTimer.IsReady() {
		p.motion.driftTimer = nil
		p.motion.velocity = 0
	}
}

func (p *Player) reverse() {
	if !ebiten.IsKeyPressed(ebiten.KeyDown) {
		return
	}

	p.motion.driftTimer = nil

	p.keepOnScreen()
	dx := math.Sin(p.Rotation) * -3
	dy := math.Cos(p.Rotation) * 3

	p.showExhaust()

	/* move player */
	p.Position.X += dx
	p.Position.Y += dy

	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)

	p.scene.PlayThrust()
}

func (p *Player) isDoneReversing() {
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		p.scene.PauseThrust()
	}
}

func (p *Player) updateExhaustSprite() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) && !ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.scene.SetExhaust(nil)
	}
}

func (p *Player) keepOnScreen() {
	p.Position = engine.WrapPosition(p.Position)
	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)
}

func (p *Player) showExhaust() {
	p.scene.SetExhaust(NewExhaust(p.spawnPoint(exhaustSpawnOffset), p.Rotation+math.Pi))
}
