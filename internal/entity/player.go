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
	driftTime            = time.Second * 30
)

type Player struct {
	scene               Scene
	currentAcceleration float64
	shotsFired          int
	Sprite              *ebiten.Image
	Rotation            float64
	Position            engine.Vector
	playerVelocity      float64
	PlayerObj           *resolv.Circle
	shootCooldown       *engine.Timer
	burstCooldown       *engine.Timer
	IsShielded          bool
	IsDying             bool
	IsDead              bool
	DyingTimer          *engine.Timer
	DyingCounter        int
	LivesRemaining      int
	LifeIndicators      []*Indicator
	shieldTimer         *engine.Timer
	ShieldsRemaining    int
	ShieldIndicators    []*Indicator
	HyperspaceIndicator *Indicator
	HyperspaceTimer     *engine.Timer
	driftAngle          float64
	driftTimer          *engine.Timer
}

func NewPlayer(scene Scene) *Player {
	sprite := assets.PlayerSprite

	/* center player on screen */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos := engine.Vector{
		X: engine.ScreenWidth/2 - halfW,
		Y: engine.ScreenHeight/2 - halfH,
	}

	/* create collision object */
	playerObj := engine.CircleFor(sprite, pos)

	/* setup life indicators*/
	var lifeIndicators []*Indicator
	var xPosition = 20.0

	for range numberOfLives {
		li := NewLifeIndicator(engine.Vector{
			X: xPosition,
			Y: 20,
		})
		lifeIndicators = append(lifeIndicators, li)
		xPosition += 50.0
	}

	/* setup shield indicators*/
	var shieldIndicators []*Indicator
	xPosition = 45.0

	for range numberOfShields {
		si := NewShieldIndicator(engine.Vector{
			X: xPosition,
			Y: 60,
		})
		shieldIndicators = append(shieldIndicators, si)
		xPosition += 50.0
	}

	p := &Player{
		Sprite:              sprite,
		scene:               scene,
		Position:            pos,
		PlayerObj:           playerObj,
		shootCooldown:       engine.NewTimer(shootCooldown),
		burstCooldown:       engine.NewTimer(burstCooldown),
		IsShielded:          false,
		IsDying:             false,
		IsDead:              false,
		DyingTimer:          engine.NewTimer(dyingAnimationAmount),
		DyingCounter:        0,
		LivesRemaining:      numberOfLives,
		LifeIndicators:      lifeIndicators,
		ShieldsRemaining:    numberOfShields,
		ShieldIndicators:    shieldIndicators,
		HyperspaceIndicator: NewHyperspaceIndicator(engine.Vector{X: 37.0, Y: 95.0}),
		HyperspaceTimer:     nil,
		driftTimer:          nil,
	}

	p.PlayerObj.SetPosition(pos.X, pos.Y)
	p.PlayerObj.Tags().Set(engine.TagPlayer)

	return p
}

func (p *Player) Draw(screen *ebiten.Image) {
	engine.DrawSprite(screen, p.Sprite, p.Position, p.Rotation)
}

func (p *Player) Update() {
	speed := rotationPerSecond / float64(ebiten.TPS())

	p.isPlayerDead()

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.Rotation -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.Rotation += speed
	}

	p.accelerate()

	p.useShield()

	p.isDoneAccelerating()

	p.isDrifting()

	p.isDriftFinished()

	p.reverse()

	p.isDoneReversing()

	p.updateExhaustSprite()

	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)

	p.burstCooldown.Update()

	p.shootCooldown.Update()

	p.fireLasers()

	p.hyperspace()

	if p.HyperspaceTimer != nil {
		p.HyperspaceTimer.Update()
	}
}

func (p *Player) accelerate() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) {
		return
	}

	p.driftTimer = nil

	p.keepOnScreen()

	if p.currentAcceleration < engine.MaxAcceleration {
		p.currentAcceleration = p.playerVelocity + 4
	}

	if p.currentAcceleration >= 8 {
		p.currentAcceleration = 8
	}

	p.playerVelocity = p.currentAcceleration

	/* move in the direction we are pointing */
	dx := math.Sin(p.Rotation) * p.currentAcceleration
	dy := math.Cos(p.Rotation) * -p.currentAcceleration

	p.showExhaust()

	/* move player */
	p.Position.X += dx
	p.Position.Y += dy

	/* play thrust sound */
	p.scene.PlayThrust()
}

func (p *Player) isDoneAccelerating() {
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		p.scene.PauseThrust()

		/* figure out velocity */
		if p.playerVelocity < p.currentAcceleration*10 {
			p.playerVelocity = p.currentAcceleration*10 - 5.0
		}

		if p.playerVelocity < 0 {
			p.playerVelocity = 0
		}

		p.currentAcceleration = 0

		/* create a drift timer */
		p.driftTimer = engine.NewTimer(driftTime)

		/* save angle of rotation */
		p.driftAngle = p.Rotation

	}
}

func (p *Player) isDrifting() {
	if p.driftTimer == nil {
		return
	}

	p.keepOnScreen()
	p.driftTimer.Update()

	decelerationSpeed := p.playerVelocity / float64(ebiten.TPS()) * 4

	p.Position.X += math.Sin(p.driftAngle) * decelerationSpeed
	p.Position.Y += math.Cos(p.driftAngle) * -decelerationSpeed
	p.PlayerObj.SetPosition(p.Position.X, p.Position.Y)

}

func (p *Player) isDriftFinished() {
	if p.driftTimer != nil && p.driftTimer.IsReady() {
		p.driftTimer = nil
		p.playerVelocity = 0
	}
}

func (p *Player) reverse() {
	if !ebiten.IsKeyPressed(ebiten.KeyDown) {
		return
	}

	p.driftTimer = nil

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
	if !p.burstCooldown.IsReady() {
		return
	}

	if p.shootCooldown.IsReady() && ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.shootCooldown.Reset()
		p.shotsFired++
		if p.shotsFired <= maxShotsPerBurst {
			bounds := p.Sprite.Bounds()
			halfW := float64(bounds.Dx()) / 2
			halfH := float64(bounds.Dy()) / 2

			spawnPos := engine.Vector{
				X: p.Position.X + halfW + math.Sin(p.Rotation)*laserSpawnOffset,
				Y: p.Position.Y + halfH + math.Cos(p.Rotation)*-laserSpawnOffset,
			}

			p.scene.SpawnLaser(spawnPos, p.Rotation)
			p.scene.PlayLaserSound(p.shotsFired)
		} else {
			p.burstCooldown.Reset()
			p.shotsFired = 0
		}
	}
}

func (p *Player) hyperspace() {
	if ebiten.IsKeyPressed(ebiten.KeyH) && (p.HyperspaceTimer == nil || p.HyperspaceTimer.IsReady()) {
		var randX, randY int

		for {
			randX = rand.Intn(engine.ScreenWidth)
			randY = rand.Intn(engine.ScreenHeight)
			collision := engine.CheckCollision(p.PlayerObj)
			if !collision {
				break
			}
		}

		p.Position.X = float64(randX)
		p.Position.Y = float64(randY)

		if p.HyperspaceTimer == nil {
			p.HyperspaceTimer = engine.NewTimer(hyperspaceCooldown)
		}

		p.HyperspaceTimer.Reset()
	}
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

func (p *Player) showExhaust() {
	/* show exhaust */
	bounds := p.Sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	/* where to spawn exhaust */
	exhaustSpawnPosition := engine.Vector{
		X: p.Position.X + halfW + math.Sin(p.Rotation)*exhaustSpawnOffset,
		Y: p.Position.Y + halfH + math.Cos(p.Rotation)*-exhaustSpawnOffset,
	}

	p.scene.SetExhaust(NewExhaust(exhaustSpawnPosition, p.Rotation+180.0*math.Pi/180.0))
}

func (p *Player) useShield() {
	if ebiten.IsKeyPressed(ebiten.KeyS) && !p.IsShielded && p.ShieldsRemaining > 0 {
		p.scene.PlayShieldSound()

		p.IsShielded = true
		p.shieldTimer = engine.NewTimer(shieldDuration)
		p.scene.SetShield(NewShield(p))
		p.ShieldsRemaining--
		p.ShieldIndicators = p.ShieldIndicators[:len(p.ShieldIndicators)-1]
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
