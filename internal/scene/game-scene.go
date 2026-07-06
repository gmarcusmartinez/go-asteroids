package scene

import (
	"fmt"
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"go-asteroids/internal/entity"
	"go-asteroids/internal/highscore"
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

	numberOfSmallMeteorsFromLargeMeteor = 4
)

type GameScene struct {
	player               *entity.Player
	baseVelocity         float64
	meteors              map[int]*entity.Meteor
	meteorCount          int
	meteorsPerLevel      int
	meteorSpawnTimer     *engine.Timer
	velocityTimer        *engine.Timer
	space                *resolv.Space
	lasers               map[int]*entity.Laser
	laserCount           int
	score                int
	explosionSprite      *ebiten.Image
	explosionSmallSprite *ebiten.Image
	explosionFrames      []*ebiten.Image
	cleanupTimer         *engine.Timer
	playerIsDead         bool
	audioContext         *audio.Context
	thrustPlayer         *audio.Player
	exhaust              *entity.Exhaust
	laserOnePlayer       *audio.Player
	laserTwoPlayer       *audio.Player
	laserThreePlayer     *audio.Player
	explosionPlayer      *audio.Player
	beatOnePlayer        *audio.Player
	beatTwoPlayer        *audio.Player
	beatTimer            *engine.Timer
	beatWaitTime         int
	playBeatOne          bool
	stars                []*entity.Star
	currentLevel         int
	shield               *entity.Shield
	shieldsUpPlayer      *audio.Player
	alienAttackTimer     *engine.Timer
	alienCount           int
	alienLaserCount      int
	alienLaserPlayer     *audio.Player
	alienLasers          map[int]*entity.AlienLaser
	alienSoundPlayer     *audio.Player
	alienSpawnTimer      *engine.Timer
	aliens               map[int]*entity.Alien
	highScore            int
	originalHighScore    int
}

/* GameScene satisfies the narrow view entities depend on. */
var _ entity.Scene = (*GameScene)(nil)

func NewGameScene() *GameScene {
	g := &GameScene{
		baseVelocity:         baseMeteorVelocity,
		meteors:              make(map[int]*entity.Meteor),
		meteorCount:          0,
		meteorsPerLevel:      2,
		meteorSpawnTimer:     engine.NewTimer(meteorSpawnTime),
		velocityTimer:        engine.NewTimer(meteorSpeedUpTime),
		space:                resolv.NewSpace(engine.ScreenWidth, engine.ScreenHeight, 16, 16),
		lasers:               make(map[int]*entity.Laser),
		laserCount:           0,
		explosionSprite:      assets.ExplosionSprite,
		explosionSmallSprite: assets.ExplosionSmallSprite,
		cleanupTimer:         engine.NewTimer(cleanupExplosionTime),
		beatTimer:            engine.NewTimer(2 * time.Second),
		beatWaitTime:         baseBeatWaitTime,
		stars:                entity.GenerateStars(numberOfStars),
		currentLevel:         1,
		aliens:               make(map[int]*entity.Alien),
		alienCount:           0,
		alienLasers:          make(map[int]*entity.AlienLaser),
		alienLaserCount:      0,
		alienSpawnTimer:      engine.NewTimer(alienSpawnTime),
		alienAttackTimer:     engine.NewTimer(alienAttackTime),
	}

	g.player = entity.NewPlayer(g)

	g.space.Add(g.player.PlayerObj)

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

	/* load the current high score */
	hs, err := highscore.Get()
	if err != nil {
		log.Println("Error getting high score", err)
	}
	g.highScore = hs
	g.originalHighScore = hs

	return g
}

/* --- entity.Scene implementation --- */

func (g *GameScene) SpawnLaser(pos engine.Vector, rotation float64) {
	g.laserCount++
	laser := entity.NewLaser(pos, rotation, g.laserCount)
	g.lasers[g.laserCount] = laser
	g.space.Add(laser.Obj)
}

func (g *GameScene) SetExhaust(e *entity.Exhaust) {
	g.exhaust = e
}

func (g *GameScene) SetShield(s *entity.Shield) {
	g.space.Add(s.Obj)
	g.shield = s
}

func (g *GameScene) ClearShield() {
	g.space.Remove(g.shield.Obj)
	g.shield = nil
}

func (g *GameScene) SetPlayerDead() {
	g.playerIsDead = true
}

func (g *GameScene) PlayThrust() {
	if !g.thrustPlayer.IsPlaying() {
		_ = g.thrustPlayer.Rewind()
		g.thrustPlayer.Play()
	}
}

func (g *GameScene) PauseThrust() {
	if g.thrustPlayer.IsPlaying() {
		g.thrustPlayer.Pause()
	}
}

func (g *GameScene) PlayLaserSound(shot int) {
	switch shot {
	case 1:
		if !g.laserOnePlayer.IsPlaying() {
			_ = g.laserOnePlayer.Rewind()
			g.laserOnePlayer.Play()
		}
	case 2:
		if !g.laserTwoPlayer.IsPlaying() {
			_ = g.laserTwoPlayer.Rewind()
			g.laserTwoPlayer.Play()
		}
	case 3:
		if !g.laserThreePlayer.IsPlaying() {
			_ = g.laserThreePlayer.Rewind()
			g.laserThreePlayer.Play()
		}
	}
}

func (g *GameScene) PlayShieldSound() {
	if !g.shieldsUpPlayer.IsPlaying() {
		_ = g.shieldsUpPlayer.Rewind()
		g.shieldsUpPlayer.Play()
	}
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
	if len(g.player.LifeIndicators) > 0 {
		for _, li := range g.player.LifeIndicators {
			li.Draw(screen)
		}
	}

	/* draw shield indicators */
	if len(g.player.ShieldIndicators) > 0 {
		for _, si := range g.player.ShieldIndicators {
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
	if g.player.HyperspaceTimer == nil || g.player.HyperspaceTimer.IsReady() {
		g.player.HyperspaceIndicator.Draw(screen)
	}

	/* draw score */
	textToDraw := fmt.Sprintf("%06d", g.score)
	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(engine.ScreenWidth/2, 40)

	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.ScoreFont,
		Size:   24,
	}, op)

	/* draw high score */
	if g.score >= g.highScore {
		g.highScore = g.score
	}

	textToDraw = fmt.Sprintf("HIGH SCORE %06d", g.highScore)
	op = &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}

	op.ColorScale.ScaleWithColor(color.White)
	op.GeoM.Translate(engine.ScreenWidth/2, 80)

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
	op.GeoM.Translate(engine.ScreenWidth/2, engine.ScreenHeight-40)

	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.LevelFont,
		Size:   16,
	}, op)

}

func (g *GameScene) Layout(width, height int) (int, int) {
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
			g.beatTimer = engine.NewTimer(time.Millisecond * time.Duration(g.beatWaitTime))
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
			m := entity.NewMeteor(g.baseVelocity, len(g.meteors)-1)
			/* add meteors to game space */
			g.space.Add(m.Obj)
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
			a := entity.NewAlien(baseAlienVelocity, g.player.Position)
			g.space.Add(a.Obj)
			g.alienCount++
			g.aliens[g.alienCount] = a
		}
	}
}

func (g *GameScene) removeOffscreenAliens() {
	for i, a := range g.aliens {
		if a.Position.X > engine.ScreenWidth+200 ||
			a.Position.Y > engine.ScreenHeight+200 ||
			a.Position.X < -200 ||
			a.Position.Y < -200 {
			g.space.Remove(a.Obj)
			delete(g.aliens, i)

		}
	}
}

func (g *GameScene) removeOffscreenLasers() {
	for i, l := range g.lasers {
		if l.Position.X > engine.ScreenWidth+200 ||
			l.Position.Y > engine.ScreenHeight+200 ||
			l.Position.X < -200 ||
			l.Position.Y < -200 {
			g.space.Remove(l.Obj)
			delete(g.lasers, i)

		}
	}

	for i, al := range g.alienLasers {
		if al.Position.X > engine.ScreenWidth+200 ||
			al.Position.Y > engine.ScreenHeight+200 ||
			al.Position.X < -200 ||
			al.Position.Y < -200 {
			g.space.Remove(al.LaserObj)
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
		if m.Obj.IsIntersecting(g.player.PlayerObj) {
			if !g.player.IsShielded {
				/* trigger dying animation */
				g.player.IsDying = true
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
		if a.Obj.IsIntersecting(g.player.PlayerObj) {
			if !g.player.IsShielded {
				/* trigger dying animation */
				g.player.IsDying = true
				/* play explosion sound */
				if !g.explosionPlayer.IsPlaying() {
					_ = g.explosionPlayer.Rewind()
					g.explosionPlayer.Play()
				}
			}
		}
	}
}

func (g *GameScene) isPlayerHitByAlienLaser() {
	for _, al := range g.alienLasers {
		if al.LaserObj.IsIntersecting(g.player.PlayerObj) {
			if !g.player.IsShielded {
				/* trigger dying animation */
				g.player.IsDying = true
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
			if a.Obj.IsIntersecting(l.Obj) {
				laserData := l.Obj.Data().(*engine.ObjectData)
				delete(g.alienLasers, laserData.Index)
				g.space.Remove(l.Obj)
				a.Sprite = g.explosionSprite
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
			if m.Obj.IsIntersecting(l.Obj) {
				if m.Obj.Tags().Has(engine.TagSmall) {
					/* hit small meteor */
					m.Sprite = g.explosionSmallSprite
					g.score++

					/* play explosion sound */
					if !g.explosionPlayer.IsPlaying() {
						_ = g.explosionPlayer.Rewind()
						g.explosionPlayer.Play()
					}
				} else {
					/* hit large meteor */
					oldPos := m.Position
					m.Sprite = g.explosionSprite
					g.score++

					/* play explosion sound */
					if !g.explosionPlayer.IsPlaying() {
						_ = g.explosionPlayer.Rewind()
						g.explosionPlayer.Play()
					}

					numToSpawn := rand.Intn(numberOfSmallMeteorsFromLargeMeteor)
					for range numToSpawn {
						meteor := entity.NewSmallMeteor(baseMeteorVelocity, len(g.meteors)-1)
						meteor.Position = engine.Vector{
							X: oldPos.X + float64(rand.Intn(100-50)) + 50,
							Y: oldPos.Y + float64(rand.Intn(100-50)) + 50,
						}
						meteor.Obj.SetPosition(meteor.Position.X, meteor.Position.Y)
						g.space.Add(meteor.Obj)
						g.meteorCount++
						g.meteors[g.meteorCount] = meteor
					}

				}
			}
		}
	}
}

func (g *GameScene) bounceMeteor(m *entity.Meteor) {
	direction := engine.Vector{
		X: (engine.ScreenWidth/2 - m.Position.X) * -1,
		Y: (engine.ScreenHeight/2 - m.Position.Y) * -1,
	}

	normalized := direction.Normalize()
	velocity := g.baseVelocity

	movement := engine.Vector{
		X: normalized.X * velocity,
		Y: normalized.Y * velocity,
	}

	m.Movement = movement
}

func (g *GameScene) cleanupMeteorsAndAliens() {
	g.cleanupTimer.Update()
	if g.cleanupTimer.IsReady() {

		/* clean up shot meteors */
		for i, m := range g.meteors {
			if m.Sprite == g.explosionSprite || m.Sprite == g.explosionSmallSprite {
				delete(g.meteors, i)
				g.space.Remove(m.Obj)
			}
		}

		/* clean up dead aliens */
		for i, a := range g.aliens {
			if a.Sprite == g.explosionSprite || a.Sprite == g.explosionSmallSprite {
				delete(g.aliens, i)
				g.space.Remove(a.Obj)
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
				bounds := a.Sprite.Bounds()
				halfW := float64(bounds.Dx()) / 2
				halfH := float64(bounds.Dy()) / 2

				var degreesRadian float64

				if !a.IsIntelligent {
					/* fire in a random direction */
					degreesRadian = rand.Float64() * (math.Pi * 2)
				} else {
					/* fire with some accuracy */
					degreesRadian = math.Atan2(g.player.Position.Y-a.Position.Y, g.player.Position.X-a.Position.X)
					degreesRadian = degreesRadian - math.Pi*-0.5
				}

				r := degreesRadian

				offsetX := float64(a.Sprite.Bounds().Dx() - int(halfW))
				offsetY := float64(a.Sprite.Bounds().Dy() - int(halfH))

				spawnPos := engine.Vector{
					X: a.Position.X + halfW + math.Sin(r) - offsetX,
					Y: a.Position.Y + halfH + math.Cos(r) - offsetY,
				}

				laser := entity.NewAlienLaser(spawnPos, r)
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

	if !g.player.IsDying {
		return
	}

	g.player.DyingTimer.Update()
	if !g.player.DyingTimer.IsReady() {
		return
	}

	g.player.DyingTimer.Reset()
	g.player.DyingCounter++

	if g.player.DyingCounter == maxDyingFrames {
		g.player.IsDying = false
		g.player.IsDead = true
		return
	}

	/* run animation */
	g.player.Sprite = g.explosionFrames[g.player.DyingCounter]
}

func (g *GameScene) isPlayerDead(state *State) {
	if !g.player.IsDead {
		return
	}
	g.player.LivesRemaining--

	if g.player.LivesRemaining == 0 {
		/* check for highscore */
		if g.score > g.originalHighScore {
			err := highscore.Update(g.score)
			if err != nil {
				log.Println(err)
			}
		}

		/* go to gameover scene */
		state.SceneManager.GoToScene(&GameOverScene{
			game:        g,
			meteors:     make(map[int]*entity.Meteor),
			meteorCount: 5,
			stars:       entity.GenerateStars(numberOfStars),
		})
	} else {
		/* decrement lives remaining */
		score := g.score
		livesRemaining := g.player.LivesRemaining
		lifeSlice := g.player.LifeIndicators[:len(g.player.LifeIndicators)-1]
		stars := g.stars
		shieldsRemaining := g.player.ShieldsRemaining
		shieldIndicatorSlice := g.player.ShieldIndicators

		g.Reset()
		g.player.LivesRemaining = livesRemaining
		g.score = score
		g.player.LifeIndicators = lifeSlice
		g.stars = stars
		g.player.ShieldsRemaining = shieldsRemaining
		g.player.ShieldIndicators = shieldIndicatorSlice
	}

}

func (g *GameScene) isLevelComplete(state *State) {
	if g.meteorCount >= g.meteorsPerLevel && len(g.meteors) == 0 {
		g.baseVelocity = baseMeteorVelocity
		g.currentLevel++

		if g.currentLevel%5 == 0 {
			if g.player.LivesRemaining < 6 {
				g.player.LivesRemaining++

				x := float64(20 + len(g.player.LifeIndicators)*50)
				y := 20.0

				g.player.LifeIndicators = append(g.player.LifeIndicators, entity.NewLifeIndicator(engine.Vector{
					X: x,
					Y: y,
				}))
			}
		}

		g.beatWaitTime = baseBeatWaitTime

		state.SceneManager.GoToScene(&LevelStartsScene{
			game:           g,
			nextLevelTimer: engine.NewTimer(time.Second * 2),
			stars:          entity.GenerateStars(numberOfStars),
		})
	}
}

func (g *GameScene) Reset() {
	g.player = entity.NewPlayer(g)
	g.meteors = make(map[int]*entity.Meteor)
	g.meteorCount = 0
	g.lasers = make(map[int]*entity.Laser)
	g.laserCount = 0
	g.score = 0
	g.baseVelocity = baseMeteorVelocity
	g.velocityTimer.Reset()
	g.meteorSpawnTimer.Reset()
	g.playerIsDead = false
	g.exhaust = nil
	g.space.RemoveAll()
	g.space.Add(g.player.PlayerObj)
	g.aliens = make(map[int]*entity.Alien)
	g.alienCount = 0
	g.alienLasers = make(map[int]*entity.AlienLaser)
	g.alienLaserCount = 0
}
