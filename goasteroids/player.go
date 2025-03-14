package goasteroids

import (
	"go-asteroids/assets"
	"math"
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
)

var currentAcceleration float64
var shotsFired = 0

type Player struct {
	game           *GameScene
	sprite         *ebiten.Image
	rotation       float64
	position       Vector
	playerVelocity float64
	playerObj      *resolv.Circle
	shootCooldown  *Timer
	burstCooldown  *Timer
	isShielded     bool
	isDying        bool
	isDead         bool
	dyingTimer     *Timer
	dyingCounter   int
	livesRemaining int
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

	p := &Player{
		sprite:         sprite,
		game:           game,
		position:       pos,
		playerObj:      playerObj,
		shootCooldown:  NewTimer(shootCooldown),
		burstCooldown:  NewTimer(burstCooldown),
		isShielded:     false,
		isDying:        false,
		isDead:         false,
		dyingTimer:     NewTimer(dyingAnimationAmount),
		dyingCounter:   0,
		livesRemaining: 1,
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
	p.isDoneAccelerating()
	p.updateExhaustSprite()

	p.playerObj.SetPosition(p.position.X, p.position.Y)
	p.burstCooldown.Update()
	p.shootCooldown.Update()

	p.fireLasers()
}

func (p *Player) accelerate() {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
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

		/* show exhaust */
		bounds := p.sprite.Bounds()
		halfW := float64(bounds.Dx()) / 2
		halfH := float64(bounds.Dy()) / 2

		/* where to spawn exhaust */
		exhauseSpawnPosition := Vector{
			p.position.X + halfW + math.Sin(p.rotation)*exhaustSpawnOffset,
			p.position.Y + halfH + math.Cos(p.rotation)*-exhaustSpawnOffset,
		}

		p.game.exhaust = NewExhaust(exhauseSpawnPosition, p.rotation+180.0*math.Pi/180.0)

		/* move player */
		p.position.X += dx
		p.position.Y += dy

		/* play thrust sound */
		if !p.game.thrustPlayer.IsPlaying() {
			_ = p.game.thrustPlayer.Rewind()
			p.game.thrustPlayer.Play()
		}
	}
}

func (p *Player) isDoneAccelerating() {
	if inpututil.IsKeyJustReleased(ebiten.KeyUp) {
		if p.game.thrustPlayer.IsPlaying() {
			p.game.thrustPlayer.Pause()
		}
	}
}

func (p *Player) updateExhaustSprite() {
	if !ebiten.IsKeyPressed(ebiten.KeyUp) && p.game.exhaust != nil {
		p.game.exhaust = nil
	}
}

func (p *Player) fireLasers() {
	if p.burstCooldown.IsReady() {
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
			} else {
				p.burstCooldown.Reset()
				shotsFired = 0
			}
		}
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
