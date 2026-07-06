package entity

import "go-asteroids/internal/engine"

// Scene is the narrow view of the game scene that entities depend on.
type Scene interface {
	SpawnLaser(pos engine.Vector, rotation float64)
	SetExhaust(*Exhaust)
	SetShield(*Shield)
	ClearShield()
	SetPlayerDead()
	PlayThrust()
	PauseThrust()
	PlayLaserSound(shot int)
	PlayShieldSound()
}
