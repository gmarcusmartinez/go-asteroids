package goasteroids

import (
	"go-asteroids/assets"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/solarlune/resolv"
)

const (
	baseMeteorVelocity   = 0.25
	meteorSpawnTime      = 100 * time.Millisecond
	meteorSpeedUpAmount  = 0.1
	meteorSpeedUpTime    = 1000 * time.Millisecond
	cleanupExplosionTime = 200 * time.Millisecond
)

type GameScene struct {
	player               *Player
	baseVelocity         float64
	meteors              map[int]*Meteor
	meteorCount          int
	meteorsPerLevel      int
	meteorSpawnTimer     *Timer
	velocityTimer        *Timer
	space                *resolv.Space
	lasers               map[int]*Laser
	laserCount           int
	score                int
	explosionSprite      *ebiten.Image
	explosionSmallSprite *ebiten.Image
	explosionFrames      []*ebiten.Image
	cleanupTimer         *Timer
	playerIsDead         bool
	audioContext         *audio.Context
	thrustPlayer         *audio.Player
}

func NewGameScene() *GameScene {
	g := &GameScene{
		baseVelocity:         baseMeteorVelocity,
		meteors:              make(map[int]*Meteor),
		meteorCount:          0,
		meteorsPerLevel:      2,
		meteorSpawnTimer:     NewTimer(meteorSpawnTime),
		velocityTimer:        NewTimer(meteorSpeedUpTime),
		space:                resolv.NewSpace(ScreenWidth, ScreenHeight, 16, 16),
		lasers:               make(map[int]*Laser),
		laserCount:           0,
		explosionSprite:      assets.ExplosionSprite,
		explosionSmallSprite: assets.ExplosionSmallSprite,
		cleanupTimer:         NewTimer(cleanupExplosionTime),
	}

	g.player = NewPlayer(g)

	g.space.Add(g.player.playerObj)

	g.explosionFrames = assets.Explosion

	/* load audio */
	g.audioContext = audio.NewContext(48000)
	thrustPlayer, _ := g.audioContext.NewPlayer(assets.ThrustSound)
	g.thrustPlayer = thrustPlayer

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()

	g.isPlayerDying()

	g.isPlayerDead(state)

	g.spawnMeteors()

	for _, m := range g.meteors {
		m.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	g.speedUpMeteors()

	g.isPlayerCollidingWithMeteor()

	g.isMeteorHitByPlayerLaser()

	g.cleanupMeteorsAndAliens()

	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
	}

	for _, l := range g.lasers {
		l.Draw(screen)
	}

}

func (g *GameScene) Layout(width, height int) (ScreenWidth, ScreenHeight int) {
	return width, height
}

func (g *GameScene) spawnMeteors() {
	g.meteorSpawnTimer.Update()
	if g.meteorSpawnTimer.IsReady() {
		g.meteorSpawnTimer.Reset()

		if len(g.meteors) < g.meteorsPerLevel && g.meteorCount < g.meteorsPerLevel {
			m := NewMeteor(g.baseVelocity, g, len(g.meteors)-1)
			/* add meteors to game space */
			g.space.Add(m.meteorObj)
			g.meteorCount++
			g.meteors[g.meteorCount] = m

		}
	}
}

func (g *GameScene) speedUpMeteors() {
	g.velocityTimer.Update()
	if g.velocityTimer.IsReady() {
		g.velocityTimer.Reset()
		g.baseVelocity += meteorSpeedUpAmount
	}
}

func (g *GameScene) isPlayerCollidingWithMeteor() {
	for _, m := range g.meteors {
		if m.meteorObj.IsIntersecting(g.player.playerObj) {
			if !g.player.isShielded {
				/* trigger dying animation */
				m.game.player.isDying = true
				break
			} else {
				/* bounce meteor*/
			}
		}
	}
}

func (g *GameScene) isMeteorHitByPlayerLaser() {
	for _, m := range g.meteors {
		for _, l := range g.lasers {
			if m.meteorObj.IsIntersecting(l.laserObj) {
				if m.meteorObj.Tags().Has(TagSmall) {
					/* hit small meteor */
					m.sprite = g.explosionSmallSprite
					g.score++
				} else {
					/* hit large meteor */
					oldPos := m.position
					m.sprite = g.explosionSprite
					g.score++

					numToSpawn := rand.Intn(numberOfSmallMeteorsFromLargeMetoer)
					for range numToSpawn {
						meteor := NewSmallMeteor(baseMeteorVelocity, g, len(m.game.meteors)-1)
						meteor.position = Vector{
							oldPos.X + float64(rand.Intn(100-50)) + 50,
							oldPos.Y + float64(rand.Intn(100-50)) + 50,
						}
						meteor.meteorObj.SetPosition(meteor.position.X, meteor.position.Y)
						g.space.Add(meteor.meteorObj)
						g.meteorCount++
						g.meteors[m.game.meteorCount] = meteor
					}

				}
			}
		}
	}
}

func (g *GameScene) cleanupMeteorsAndAliens() {
	g.cleanupTimer.Update()
	if g.cleanupTimer.IsReady() {
		for i, m := range g.meteors {
			if m.sprite == g.explosionSprite || m.sprite == g.explosionSmallSprite {
				delete(g.meteors, i)
				g.space.Remove(m.meteorObj)
			}
		}
		g.cleanupTimer.Reset()
	}
}

func (g *GameScene) isPlayerDying() {
	const maxDyingFrames = 12

	if !g.player.isDying {
		return
	}

	g.player.dyingTimer.Update()
	if !g.player.dyingTimer.IsReady() {
		return
	}

	g.player.dyingTimer.Reset()
	g.player.dyingCounter++

	if g.player.dyingCounter == maxDyingFrames {
		g.player.isDying = false
		g.player.isDead = true
		return
	}

	/* run animation */
	g.player.sprite = g.explosionFrames[g.player.dyingCounter]
}

func (g *GameScene) isPlayerDead(state *State) {
	if g.player.isDead {
		g.player.livesRemaining--
		if g.player.livesRemaining == 0 {
			state.SceneManager.GoToScene(NewGameScene())
		}
	}
}
