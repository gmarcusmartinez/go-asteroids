package entity

import (
	"go-asteroids/internal/engine"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	shootCooldown    = time.Millisecond * 150
	burstCooldown    = time.Millisecond * 500
	laserSpawnOffset = 50.0
	maxShotsPerBurst = 3
)

type weapon struct {
	shootCooldown *engine.Timer
	burstCooldown *engine.Timer
	shotsFired    int
}

func newWeapon() weapon {
	return weapon{
		shootCooldown: engine.NewTimer(shootCooldown),
		burstCooldown: engine.NewTimer(burstCooldown),
	}
}

func (w *weapon) update() {
	w.burstCooldown.Update()
	w.shootCooldown.Update()
}

func (w *weapon) fire() (int, bool) {
	if !w.burstCooldown.IsReady() || !w.shootCooldown.IsReady() {
		return 0, false
	}

	w.shootCooldown.Reset()
	w.shotsFired++

	if w.shotsFired > maxShotsPerBurst {
		w.burstCooldown.Reset()
		w.shotsFired = 0
		return 0, false
	}

	return w.shotsFired, true
}

func (p *Player) fireLasers() {
	p.weapon.update()

	if !ebiten.IsKeyPressed(ebiten.KeySpace) {
		return
	}

	shot, ok := p.weapon.fire()
	if !ok {
		return
	}

	p.scene.SpawnLaser(p.spawnPoint(laserSpawnOffset), p.Rotation)
	p.scene.PlayLaserSound(shot)
}
