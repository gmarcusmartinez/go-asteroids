package goasteroids

import (
	"go-asteroids/assets"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
)

const (
	rotationPerSecond    = math.Pi
	maxAcceleration      = 8.0
	ScreenWidth          = 1280
	ScreenHeight         = 720
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

var currentAcceleration float64
var shotsFired = 0

type Player struct {
	game                *GameScene
	sprite              *ebiten.Image
	rotation            float64
	position            Vector
	playerVelocity      float64
	playerObj           *resolv.Circle
	shootCooldown       *Timer
	burstCooldown       *Timer
	isShielded          bool
	isDying             bool
	isDead              bool
	dyingTimer          *Timer
	dyingCounter        int
	livesRemaining      int
	lifeIndicators      []*LifeIndicator
	shieldTimer         *Timer
	shieldsRemaining    int
	shieldIndicators    []*ShieldIndicator
	hyperspaceIndicator *HyperspaceIndicator
	hyperspaceTimer     *Timer
	driftAngle          float64
	driftTimer          *Timer
}

func NewPlayer(game *GameScene) *Player {
	sprite := assets.PlayerSprite

	/* center player on screen */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos := Vector{
		X: ScreenWidth/2 - halfW,
		Y: ScreenHeight/2 - halfH,
	}

	/* create collision object */
	playerObj := resolv.NewCircle(pos.X, pos.Y, float64(sprite.Bounds().Dx()/2))

	/* setup life indicators*/
	var lifeIndicators []*LifeIndicator
	var xPosition = 20.0

	for range numberOfLives {
		li := NewLifeIndicator(Vector{
			X: xPosition,
			Y: 20,
		})
		lifeIndicators = append(lifeIndicators, li)
		xPosition += 50.0
	}

	/* setup shield indicators*/
	var shieldIndicators []*ShieldIndicator
	xPosition = 45.0

	for range numberOfShields {
		si := NewShieldIndicator(Vector{
			X: xPosition,
			Y: 60,
		})
		shieldIndicators = append(shieldIndicators, si)
		xPosition += 50.0
	}

	p := &Player{
		sprite:              sprite,
		game:                game,
		position:            pos,
		playerObj:           playerObj,
		shootCooldown:       NewTimer(shootCooldown),
		burstCooldown:       NewTimer(burstCooldown),
		isShielded:          false,
		isDying:             false,
		isDead:              false,
		dyingTimer:          NewTimer(dyingAnimationAmount),
		dyingCounter:        0,
		livesRemaining:      numberOfLives,
		lifeIndicators:      lifeIndicators,
		shieldsRemaining:    numberOfShields,
		shieldIndicators:    shieldIndicators,
		hyperspaceIndicator: NewHyperspaceIndicator(Vector{X: 37.0, Y: 95.0}),
		hyperspaceTimer:     nil,
		driftTimer:          nil,
	}

	p.playerObj.SetPosition(pos.X, pos.Y)
	p.playerObj.Tags().Set(TagPlayer)

	return p
}

func (p *Player) Draw(screen *ebiten.Image) {
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}

	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(p.rotation)
	op.GeoM.Translate(halfW, halfH)

	op.GeoM.Translate(p.position.X, p.position.Y)

	screen.DrawImage(p.sprite, op)
}

func (p *Player) Update() {
	speed := rotationPerSecond / float64(ebiten.TPS())

	p.isPlayerDead()

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}

	p.accelerate()

	p.useShield()

	p.isDoneAccelerating()

	p.isDrifting()

	p.isDriftFinished()

	p.reverse()

	p.isDoneReversing()

	p.updateExhaustSprite()

	p.playerObj.SetPosition(p.position.X, p.position.Y)

	p.burstCooldown.Update()

	p.shootCooldown.Update()

	p.fireLasers()

	p.hyperspace()

	if p.hyperspaceTimer != nil {
		p.hyperspaceTimer.Update()
	}
}

func (p *Player) accelerate() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) {
		return
	}

	p.driftTimer = nil

	p.keepOnScreen()

	if currentAcceleration < maxAcceleration {
		currentAcceleration = p.playerVelocity + 4
	}

	if currentAcceleration >= 8 {
		currentAcceleration = 8
	}

	p.playerVelocity = currentAcceleration

	/* move in the direction we are pointing */
	dx := math.Sin(p.rotation) * currentAcceleration
	dy := math.Cos(p.rotation) * -currentAcceleration

	p.showExhaust()

	/* move player */
	p.position.X += dx
	p.position.Y += dy

	/* play thrust sound */
	if !p.game.thrustPlayer.IsPlaying() {
		_ = p.game.thrustPlayer.Rewind()
		p.game.thrustPlayer.Play()
	}
}

func (p *Player) isDoneAccelerating() {
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		if p.game.thrustPlayer.IsPlaying() {
			p.game.thrustPlayer.Pause()
		}

		/* figure out velocity */
		if p.playerVelocity < currentAcceleration*10 {
			p.playerVelocity = currentAcceleration*10 - 5.0
		}

		if p.playerVelocity < 0 {
			p.playerVelocity = 0
		}

		currentAcceleration = 0

		/* create a drift timer */
		p.driftTimer = NewTimer(driftTime)

		/* save angle of rotation */
		p.driftAngle = p.rotation

	}
}

func (p *Player) isDrifting() {
	if p.driftTimer == nil {
		return
	}

	p.keepOnScreen()
	p.driftTimer.Update()

	decelerationSpeed := p.playerVelocity / float64(ebiten.TPS()) * 4

	p.position.X += math.Sin(p.driftAngle) * decelerationSpeed
	p.position.Y += math.Cos(p.driftAngle) * -decelerationSpeed
	p.playerObj.SetPosition(p.position.X, p.position.Y)

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
	dx := math.Sin(p.rotation) * -3
	dy := math.Cos(p.rotation) * 3

	p.showExhaust()

	/* move player */
	p.position.X += dx
	p.position.Y += dy

	p.playerObj.SetPosition(p.position.X, p.position.Y)

	if !p.game.thrustPlayer.IsPlaying() {
		_ = p.game.thrustPlayer.Rewind()
		p.game.thrustPlayer.Play()
	}
}

func (p *Player) isDoneReversing() {
	if inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		if p.game.thrustPlayer.IsPlaying() {
			p.game.thrustPlayer.Pause()
		}
	}
}

func (p *Player) updateExhaustSprite() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) && !ebiten.IsKeyPressed(ebiten.KeyDown) && p.game.exhaust != nil {
		p.game.exhaust = nil
	}
}

func (p *Player) fireLasers() {
	if !p.burstCooldown.IsReady() {
		return
	}

	if p.shootCooldown.IsReady() && ebiten.IsKeyPressed(ebiten.KeySpace) {
		p.shootCooldown.Reset()
		shotsFired++
		if shotsFired <= maxShotsPerBurst {
			bounds := p.sprite.Bounds()
			halfW := float64(bounds.Dx()) / 2
			halfH := float64(bounds.Dy()) / 2

			spawnPos := Vector{
				p.position.X + halfW + math.Sin(p.rotation)*laserSpawnOffset,
				p.position.Y + halfH + math.Cos(p.rotation)*-laserSpawnOffset,
			}

			p.game.laserCount++
			laser := NewLaser(spawnPos, p.rotation, p.game.laserCount, p.game)
			p.game.lasers[p.game.laserCount] = laser

			p.game.space.Add(laser.laserObj)

			switch shotsFired {
			case 1:
				if !p.game.laserOnePlayer.IsPlaying() {
					_ = p.game.laserOnePlayer.Rewind()
					p.game.laserOnePlayer.Play()
				}
			case 2:
				if !p.game.laserTwoPlayer.IsPlaying() {
					_ = p.game.laserTwoPlayer.Rewind()
					p.game.laserTwoPlayer.Play()
				}
			case 3:
				if !p.game.laserThreePlayer.IsPlaying() {
					_ = p.game.laserThreePlayer.Rewind()
					p.game.laserThreePlayer.Play()
				}
			}
		} else {
			p.burstCooldown.Reset()
			shotsFired = 0
		}
	}
}

func (p *Player) hyperspace() {
	if ebiten.IsKeyPressed(ebiten.KeyH) && (p.hyperspaceTimer == nil || p.hyperspaceTimer.IsReady()) {
		var randX, randY int

		for {
			randX = rand.Intn(ScreenWidth)
			randY = rand.Intn(ScreenHeight)
			collision := p.game.checkCollision(p.playerObj, nil)
			if !collision {
				break
			}
		}

		p.position.X = float64(randX)
		p.position.Y = float64(randY)

		if p.hyperspaceTimer == nil {
			p.hyperspaceTimer = NewTimer(hyperspaceCooldown)
		}

		p.hyperspaceTimer.Reset()
	}
}

func (p *Player) isPlayerDead() {
	if p.isDead {
		p.game.playerIsDead = true
	}
}

func (p *Player) keepOnScreen() {
	if p.position.X >= float64(ScreenWidth) {
		p.position.X = 0
		p.playerObj.SetPosition(0, p.position.Y)
	}

	if p.position.X < 0 {
		p.position.X = ScreenWidth
		p.playerObj.SetPosition(ScreenWidth, p.position.Y)

	}

	if p.position.Y >= float64(ScreenHeight) {
		p.position.Y = 0
		p.playerObj.SetPosition(p.position.X, 0)

	}

	if p.position.Y < 0 {
		p.position.Y = ScreenHeight
		p.playerObj.SetPosition(p.position.X, ScreenHeight)
	}
}

func (p *Player) showExhaust() {
	/* show exhaust */
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	/* where to spawn exhaust */
	exhaustSpawnPosition := Vector{
		p.position.X + halfW + math.Sin(p.rotation)*exhaustSpawnOffset,
		p.position.Y + halfH + math.Cos(p.rotation)*-exhaustSpawnOffset,
	}

	p.game.exhaust = NewExhaust(exhaustSpawnPosition, p.rotation+180.0*math.Pi/180.0)
}

func (p *Player) useShield() {
	if ebiten.IsKeyPressed(ebiten.KeyS) && !p.isShielded && p.shieldsRemaining > 0 {
		if !p.game.shieldsUpPlayer.IsPlaying() {
			_ = p.game.shieldsUpPlayer.Rewind()
			p.game.shieldsUpPlayer.Play()
		}

		p.isShielded = true
		p.shieldTimer = NewTimer(shieldDuration)
		p.game.shield = NewShield(Vector{}, p.rotation, p.game)
		p.shieldsRemaining--
		p.shieldIndicators = p.shieldIndicators[:len(p.shieldIndicators)-1]
	}

	if p.shieldTimer != nil && p.isShielded {
		p.shieldTimer.Update()
	}

	if p.shieldTimer != nil && p.shieldTimer.IsReady() {
		p.shieldTimer = nil
		p.isShielded = false
		p.game.space.Remove(p.game.shield.shieldObj)
		p.game.shield = nil
	}

}
