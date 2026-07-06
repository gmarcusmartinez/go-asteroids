package scene

import (
	"go-asteroids/internal/engine"
	"go-asteroids/internal/entity"
	"math/rand"
)

func (g *GameScene) isPlayerCollidingWithMeteor() {
	for _, m := range g.meteors {
		if m.Obj.IsIntersecting(g.player.PlayerObj) {
			if !g.player.IsShielded {
				/* trigger dying animation */
				g.player.IsDying = true
				/* play explosion sound */
				playOnce(g.explosionPlayer)
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
				playOnce(g.explosionPlayer)
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
				playOnce(g.explosionPlayer)
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
				playOnce(g.explosionPlayer)
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
					playOnce(g.explosionPlayer)
				} else {
					/* hit large meteor */
					oldPos := m.Position
					m.Sprite = g.explosionSprite
					g.score++

					/* play explosion sound */
					playOnce(g.explosionPlayer)

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
