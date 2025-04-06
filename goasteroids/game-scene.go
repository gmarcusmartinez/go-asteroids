package goasteroids

import (
	"fmt"
	"go-asteroids/assets"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/solarlune/resolv"
)

const (
	baseMeteorVelocity   = 0.25
	meteorSpawnTime      = 100 * time.Millisecond
	meteorSpeedUpAmount  = 0.1
	meteorSpeedUpTime    = 1000 * time.Millisecond
	cleanupExplosionTime = 200 * time.Millisecond
	baseBeatWaitTime     = 1600
	numberOfStars        = 1000
	alienAttackTime      = 3 * time.Second
	alienSpawnTime       = 12 * time.Second
	baseAlienVelocity    = 0.5
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
	exhaust              *Exhaust
	laserOnePlayer       *audio.Player
	laserTwoPlayer       *audio.Player
	laserThreePlayer     *audio.Player
	explosionPlayer      *audio.Player
	beatOnePlayer        *audio.Player
	beatTwoPlayer        *audio.Player
	beatTimer            *Timer
	beatWaitTime         int
	playBeatOne          bool
	stars                []*Star
	currentLevel         int
	shield               *Shield
	shieldsUpPlayer      *audio.Player
	alienAttackTimer     *Timer
	alienCount           int
	alienLaserCount      int
	alienLaserPlayer     *audio.Player
	alienLasers          map[int]*AlienLaser
	alienSoundPlayer     *audio.Player
	alienSpawnTimer      *Timer
	aliens               map[int]*Alien
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
		beatTimer:            NewTimer(2 * time.Second),
		beatWaitTime:         baseBeatWaitTime,
		stars:                GenerateStars(numberOfStars),
		currentLevel:         1,
		aliens:               make(map[int]*Alien),
		alienCount:           0,
		alienLasers:          make(map[int]*AlienLaser),
		alienLaserCount:      0,
		alienSpawnTimer:      NewTimer(alienSpawnTime),
		alienAttackTimer:     NewTimer(alienAttackTime),
	}

	g.player = NewPlayer(g)

	g.space.Add(g.player.playerObj)

	g.explosionFrames = assets.Explosion

	/* load audio */
	g.audioContext = audio.NewContext(48000)
	thrustPlayer, _ := g.audioContext.NewPlayer(assets.ThrustSound)
	g.thrustPlayer = thrustPlayer

	laserOnePlayer, _ := g.audioContext.NewPlayer(assets.LaserOneSound)
	g.laserOnePlayer = laserOnePlayer

	laserTwoPlayer, _ := g.audioContext.NewPlayer(assets.LaserTwoSound)
	g.laserTwoPlayer = laserTwoPlayer

	laserThreePlayer, _ := g.audioContext.NewPlayer(assets.LaserThreeSound)
	g.laserThreePlayer = laserThreePlayer

	explosionPlayer, _ := g.audioContext.NewPlayer(assets.ExplosionSound)
	g.explosionPlayer = explosionPlayer

	beatOnePlayer, _ := g.audioContext.NewPlayer(assets.BeatOneSound)
	beatOnePlayer.SetVolume(0.5)
	g.beatOnePlayer = beatOnePlayer

	beatTwoPlayer, _ := g.audioContext.NewPlayer(assets.BeatTwoSound)
	beatTwoPlayer.SetVolume(0.5)
	g.beatTwoPlayer = beatTwoPlayer

	shieldsUpPlayer, _ := g.audioContext.NewPlayer(assets.ShieldSound)
	g.shieldsUpPlayer = shieldsUpPlayer

	alienLaserPlayer, _ := g.audioContext.NewPlayer(assets.AlienLaserSound)
	g.alienLaserPlayer = alienLaserPlayer

	alienSoundPlayer, _ := g.audioContext.NewPlayer(assets.AlienSound)
	alienSoundPlayer.SetVolume(0.5)
	g.alienSoundPlayer = alienSoundPlayer

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()

	g.updateExhaust()

	g.updateShield()

	g.isPlayerDying()

	g.isPlayerDead(state)

	g.spawnMeteors()

	g.spawnAliens()

	for _, a := range g.aliens {
		a.Update()
	}

	g.letAliensAttack()

	for _, al := range g.alienLasers {
		al.Update()
	}

	for _, m := range g.meteors {
		m.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	g.speedUpMeteors()

	g.isPlayerCollidingWithMeteor()

	g.isPlayerCollidingWithAlien()

	g.isPlayerHitByAlienLaser()

	g.isAlienHitByPlayerLaser()

	g.isMeteorHitByPlayerLaser()

	g.cleanupMeteorsAndAliens()

	g.beatSound()

	g.isLevelComplete(state)

	g.removeOffscreenAliens()

	g.removeOffscreenLasers()

	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	/* draw stars */
	for _, s := range g.stars {
		s.Draw(screen)
	}

	/* draw exhaust */
	if g.exhaust != nil {
		g.exhaust.Draw(screen)
	}

	/* draw shield */
	if g.shield != nil {
		g.shield.Draw(screen)
	}

	/* draw meteors */
	for _, m := range g.meteors {
		m.Draw(screen)
	}

	/* draw lasers */
	for _, l := range g.lasers {
		l.Draw(screen)
	}

	/* draw life indicators */
	if len(g.player.lifeIndicators) > 0 {
		for _, li := range g.player.lifeIndicators {
			li.Draw(screen)
		}
	}

	/* draw shield indicators */
	if len(g.player.shieldIndicators) > 0 {
		for _, si := range g.player.shieldIndicators {
			si.Draw(screen)
		}
	}

	/* draw aliens  */
	for _, a := range g.aliens {
		a.Draw(screen)
	}

	/* draw aliens lasers  */
	for _, al := range g.alienLasers {
		al.Draw(screen)
	}

	/* draw hyperspace indicator */
	if g.player.hyperspaceTimer == nil || g.player.hyperspaceTimer.IsReady() {
		g.player.hyperspaceIndicator.Draw(screen)
	}

	/* draw score */
	textToDraw := fmt.Sprintf("%06d", g.score)
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, 40)

	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   24,
	}, op)

	/* draw high score */
	if g.score >= highScore {
		highScore = g.score
	}

	textToDraw = fmt.Sprintf("HIGH SCORE %06d", highScore)
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}

	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, 80)

	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   16,
	}, op)

	/* draw current level */
	textToDraw = fmt.Sprintf("LEVEL %d", g.currentLevel)
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}

	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(ScreenWidth/2, ScreenHeight-40)

	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.LevelFont,
		Size:   16,
	}, op)

}

func (g *GameScene) Layout(width, height int) (ScreenWidth, ScreenHeight int) {
	return width, height
}

func (g *GameScene) beatSound() {
	g.beatTimer.Update()
	if g.beatTimer.IsReady() {
		if g.playBeatOne {
			_ = g.beatOnePlayer.Rewind()
			g.beatOnePlayer.Play()
			g.beatTimer.Reset()
		} else {
			_ = g.beatTwoPlayer.Rewind()
			g.beatTwoPlayer.Play()
			g.beatTimer.Reset()
		}
		g.playBeatOne = !g.playBeatOne
		/* speed up timer */
		if g.beatWaitTime > 400 {
			g.beatWaitTime = g.beatWaitTime - 25
			g.beatTimer = NewTimer(time.Millisecond * time.Duration(g.beatWaitTime))
		}
	}
}

func (g *GameScene) updateExhaust() {
	if g.exhaust != nil {
		g.exhaust.Update()
	}
}

func (g *GameScene) updateShield() {
	if g.shield != nil {
		g.shield.Update()
	}
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

func (g *GameScene) spawnAliens() {
	g.alienSpawnTimer.Update()

	if len(g.aliens) != 0 {
		return
	}

	if g.alienSpawnTimer.IsReady() {
		g.alienSpawnTimer.Reset()
		rnd := rand.Intn(100-1) + 1

		if rnd > 50 {
			a := NewAlien(baseAlienVelocity, g)
			g.space.Add(a.alienObj)
			g.alienCount++
			g.aliens[g.alienCount] = a
		}
	}
}

func (g *GameScene) removeOffscreenAliens() {
	for i, a := range g.aliens {
		if a.position.X > ScreenWidth+200 ||
			a.position.Y > ScreenHeight+200 ||
			a.position.X < -200 ||
			a.position.Y < -200 {
			g.space.Remove(a.alienObj)
			delete(g.aliens, i)

		}
	}
}

func (g *GameScene) removeOffscreenLasers() {
	for i, l := range g.lasers {
		if l.position.X > ScreenWidth+200 ||
			l.position.Y > ScreenHeight+200 ||
			l.position.X < -200 ||
			l.position.Y < -200 {
			g.space.Remove(l.laserObj)
			delete(g.lasers, i)

		}
	}

	for i, al := range g.alienLasers {
		if al.position.X > ScreenWidth+200 ||
			al.position.Y > ScreenHeight+200 ||
			al.position.X < -200 ||
			al.position.Y < -200 {
			g.space.Remove(al.laserObj)
			delete(g.alienLasers, i)

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
				/* play explosion sound */
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
				break
			} else {
				/* bounce meteor if shielded */
				g.bounceMeteor(m)
			}
		}
	}
}

func (g *GameScene) isPlayerCollidingWithAlien() {
	for _, a := range g.aliens {
		if a.alienObj.IsIntersecting(g.player.playerObj) {
			if !a.game.player.isShielded {
				/* trigger dying animation */
				a.game.player.isDying = true
				/* play explosion sound */
				if !a.game.explosionPlayer.IsPlaying() {
					_ = a.game.explosionPlayer.Rewind()
					a.game.explosionPlayer.Play()
				}
			}
		}
	}
}

func (g *GameScene) isPlayerHitByAlienLaser() {
	for _, al := range g.alienLasers {
		if al.laserObj.IsIntersecting(g.player.playerObj) {
			if !g.player.isShielded {
				/* trigger dying animation */
				g.player.isDying = true
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
			}
		}
	}
}

func (g *GameScene) isAlienHitByPlayerLaser() {
	for _, a := range g.aliens {
		for _, l := range g.lasers {
			if a.alienObj.IsIntersecting(l.laserObj) {
				laserData := l.laserObj.Data().(*ObjectData)
				delete(g.alienLasers, laserData.index)
				g.space.Remove(l.laserObj)
				a.sprite = g.explosionSprite
				g.score = g.score + 50

				/* play explosion sound*/
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
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

					/* play explosion sound */
					if !g.explosionPlayer.IsPlaying() {
						_ = g.explosionPlayer.Rewind()
						g.explosionPlayer.Play()
					}
				} else {
					/* hit large meteor */
					oldPos := m.position
					m.sprite = g.explosionSprite
					g.score++

					/* play explosion sound */
					if !g.explosionPlayer.IsPlaying() {
						_ = g.explosionPlayer.Rewind()
						g.explosionPlayer.Play()
					}

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

func (g *GameScene) bounceMeteor(m *Meteor) {
	direction := Vector{
		X: (ScreenWidth/2 - m.position.X) * -1,
		Y: (ScreenHeight/2 - m.position.Y) * -1,
	}

	normalized := direction.Normalize()
	velocity := g.baseVelocity

	movement := Vector{
		X: normalized.X * velocity,
		Y: normalized.Y * velocity,
	}

	m.movement = movement
}

func (g *GameScene) cleanupMeteorsAndAliens() {
	g.cleanupTimer.Update()
	if g.cleanupTimer.IsReady() {

		/* clean up shot meteors */
		for i, m := range g.meteors {
			if m.sprite == g.explosionSprite || m.sprite == g.explosionSmallSprite {
				delete(g.meteors, i)
				g.space.Remove(m.meteorObj)
			}
		}

		/* clean up dead aliens */
		for i, a := range g.aliens {
			if a.sprite == g.explosionSprite || a.sprite == g.explosionSmallSprite {
				delete(g.aliens, i)
				g.space.Remove(a.alienObj)
			}
		}

		g.cleanupTimer.Reset()
	}
}

func (g *GameScene) letAliensAttack() {
	if len(g.aliens) > 0 {
		if !g.alienSoundPlayer.IsPlaying() {
			_ = g.alienSoundPlayer.Rewind()
			g.alienSoundPlayer.Play()
		}

		/* update the alien attack timer */
		g.alienAttackTimer.Update()

		/* if timer reached reset timer and attack */
		if g.alienAttackTimer.IsReady() {
			g.alienAttackTimer.Reset()

			for _, a := range g.aliens {
				bounds := a.sprite.Bounds()
				halfW := float64(bounds.Dx()) / 2
				halfH := float64(bounds.Dy()) / 2

				var degreesRadian float64

				if !a.isIntelligent {
					/* fire in a random direction */
					degreesRadian = rand.Float64() * (math.Pi * 2)
				} else {
					/* fire with some accuracy */
					degreesRadian = math.Atan2(g.player.position.Y-a.position.Y, g.player.position.X-a.position.X)
					degreesRadian = degreesRadian - math.Pi*-0.5
				}

				r := degreesRadian

				offsetX := float64(a.sprite.Bounds().Dx() - int(halfW))
				offsetY := float64(a.sprite.Bounds().Dy() - int(halfH))

				spawnPos := Vector{
					X: a.position.X + halfW + math.Sin(r) - offsetX,
					Y: a.position.Y + halfH + math.Cos(r) - offsetY,
				}

				laser := NewAlienLaser(spawnPos, r)
				g.alienLaserCount++
				g.alienLasers[g.alienLaserCount] = laser

				if !g.alienLaserPlayer.IsPlaying() {
					_ = g.alienLaserPlayer.Rewind()
					g.alienLaserPlayer.Play()
				}
			}
		}
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
	if !g.player.isDead {
		return
	}
	g.player.livesRemaining--

	if g.player.livesRemaining == 0 {
		/* check for highscore */
		if g.score > originalHighScore {
			err := updateHighScore(g.score)
			if err != nil {
				log.Println(err)
			}
		}

		/* go to gameover scene */
		state.SceneManager.GoToScene(&GameOverScene{
			game:        g,
			meteors:     make(map[int]*Meteor),
			meteorCount: 5,
			stars:       GenerateStars(numberOfStars),
		})
	} else {
		/* decrement lives remaining */
		score := g.score
		livesRemaining := g.player.livesRemaining
		lifeSlice := g.player.lifeIndicators[:len(g.player.lifeIndicators)-1]
		stars := g.stars
		shieldsRemaining := g.player.shieldsRemaining
		shieldIndicatorSlice := g.player.shieldIndicators

		g.Reset()
		g.player.livesRemaining = livesRemaining
		g.score = score
		g.player.lifeIndicators = lifeSlice
		g.stars = stars
		g.player.shieldsRemaining = shieldsRemaining
		g.player.shieldIndicators = shieldIndicatorSlice
	}

}

func (g *GameScene) isLevelComplete(state *State) {
	if g.meteorCount >= g.meteorsPerLevel && len(g.meteors) == 0 {
		g.baseVelocity = baseMeteorVelocity
		g.currentLevel++

		if g.currentLevel%5 == 0 {
			if g.player.livesRemaining < 6 {
				g.player.livesRemaining++

				x := float64(20 + len(g.player.lifeIndicators)*50)
				y := 20.0

				g.player.lifeIndicators = append(g.player.lifeIndicators, NewLifeIndicator(Vector{
					X: x,
					Y: y,
				}))
			}
		}

		g.beatWaitTime = baseBeatWaitTime

		state.SceneManager.GoToScene(&LevelStartsScene{
			game:           g,
			nextLevelTimer: NewTimer(time.Second * 2),
			stars:          GenerateStars(numberOfStars),
		})
	}
}

func (g *GameScene) Reset() {
	g.player = NewPlayer(g)
	g.meteors = make(map[int]*Meteor)
	g.meteorCount = 0
	g.lasers = make(map[int]*Laser)
	g.laserCount = 0
	g.score = 0
	g.baseVelocity = baseMeteorVelocity
	g.velocityTimer.Reset()
	g.meteorSpawnTimer.Reset()
	g.playerIsDead = false
	g.exhaust = nil
	g.space.RemoveAll()
	g.space.Add(g.player.playerObj)
	g.player.shieldsRemaining = numberOfShields
	g.player.isShielded = false
	g.aliens = make(map[int]*Alien)
	g.alienCount = 0
	g.alienLasers = make(map[int]*AlienLaser)
	g.alienLaserCount = 0
}
