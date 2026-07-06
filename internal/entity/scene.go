package entity

import "go-asteroids/internal/engine"

// Scene is the narrow view of the game scene that entities depend on. It is
// declared here (in entity) so that scene imports entity, never the reverse.
// GameScene satisfies it structurally.
type Scene interface {
	// SpawnLaser fires a player laser from pos in the given rotation.
	SpawnLaser(pos engine.Vector, rotation float64)

	// SetExhaust stores the current exhaust plume (nil clears it).
	SetExhaust(*Exhaust)

	// SetShield registers a new shield (adding it to the space); ClearShield
	// removes the active shield.
	SetShield(*Shield)
	ClearShield()

	// SetPlayerDead marks the scene's player-dead flag.
	SetPlayerDead()

	// Semantic audio — replaces raw *audio.Player reach-in from entities.
	PlayThrust()
	PauseThrust()
	PlayLaserSound(shot int)
	PlayShieldSound()
}
