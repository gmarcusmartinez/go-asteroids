package entity

import (
	"go-asteroids/internal/engine"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	numberOfShields    = 3
	shieldDuration     = time.Second * 6
	hyperspaceCooldown = time.Second * 10
	hyperspaceMaxTries = 32
)

func (p *Player) useShield() {
	if ebiten.IsKeyPressed(ebiten.KeyS) && !p.IsShielded && p.ShieldsRemaining > 0 {
		p.scene.PlayShieldSound()

		p.IsShielded = true
		p.shieldTimer = engine.NewTimer(shieldDuration)
		p.scene.SetShield(NewShield(p))
		p.ShieldsRemaining--
	}

	if p.shieldTimer != nil && p.IsShielded {
		p.shieldTimer.Update()
	}

	if p.shieldTimer != nil && p.shieldTimer.IsReady() {
		p.shieldTimer = nil
		p.IsShielded = false
		p.scene.ClearShield()
	}
}

func (p *Player) hyperspace() {
	if p.hyperspaceTimer != nil {
		p.hyperspaceTimer.Update()
	}

	if !ebiten.IsKeyPressed(ebiten.KeyH) || !p.HyperspaceReady() {
		return
	}

	if !p.jumpToSafeSpot() {
		return
	}

	if p.hyperspaceTimer == nil {
		p.hyperspaceTimer = engine.NewTimer(hyperspaceCooldown)
	}

	p.hyperspaceTimer.Reset()
}

// teleports to a random collision-free position, giving up after a maximum number of tries
func (p *Player) jumpToSafeSpot() bool {
	for range hyperspaceMaxTries {
		x := float64(rand.Intn(engine.ScreenWidth))
		y := float64(rand.Intn(engine.ScreenHeight))

		p.PlayerObj.SetPosition(x, y)

		if !engine.CheckCollision(p.PlayerObj) {
			p.Position = engine.Vector{X: x, Y: y}
			return true
		}
	}

	/* no safe spot found; stay put */
	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)
	return false
}

func (p *Player) HyperspaceReady() bool {
	return p.hyperspaceTimer == nil || p.hyperspaceTimer.IsReady()
}
