package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
)

const (
	rotationPerSecond    = math.Pi
	exhaustSpawnOffset   = -50.0
	shootCooldown        = time.Millisecond * 150
	burstCooldown        = time.Millisecond * 500
	laserSpawnOffset     = 50.0
	maxShotsPerBurst     = 3
	dyingAnimationAmount = 50 * time.Millisecond
	numberOfLives        = 3
	numberOfShields      = 3
	shieldDuration       = time.Second * 6
	hyperspaceCooldown   = time.Second * 10
	hyperspaceMaxTries   = 32
	driftTime            = time.Second * 30
)

/* motion is the player's thrust and drift state */
type motion struct {
	acceleration float64
	velocity     float64
	driftAngle   float64
	driftTimer   *engine.Timer
}

// weapon is the burst-fire state: up to maxShotsPerBurst quick shots,
// then a longer pause.
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

// fire consumes one shot, returning its number within the burst, or false
// while cooling down or when the burst is spent.
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

// jumpToSafeSpot teleports to a random collision-free position, giving up
// after hyperspaceMaxTries so a crowded screen cannot hang the game loop.
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

/* HyperspaceReady reports whether hyperspace is off cooldown. */
func (p *Player) HyperspaceReady() bool {
	return p.hyperspaceTimer == nil || p.hyperspaceTimer.IsReady()
}

func (p *Player) isPlayerDead() {
	if p.IsDead {
		p.scene.SetPlayerDead()
	}
}

func (p *Player) keepOnScreen() {
	p.Position = engine.WrapPosition(p.Position)
	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)
}

// spawnPoint returns the point `distance` ahead of the sprite center along
// the player's heading; negative distances land behind the ship.
func (p *Player) spawnPoint(distance float64) engine.Vector {
	bounds := p.Sprite.Bounds()

	return engine.Vector{
		X: p.Position.X + float64(bounds.Dx())/2 + math.Sin(p.Rotation)*distance,
		Y: p.Position.Y + float64(bounds.Dy())/2 - math.Cos(p.Rotation)*distance,
	}
}

func (p *Player) showExhaust() {
	p.scene.SetExhaust(NewExhaust(p.spawnPoint(exhaustSpawnOffset), p.Rotation+math.Pi))
}

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
