package goasteroids

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	baseMeteorVelocity  = 0.25
	meteorSpawnTime     = 100 * time.Millisecond
	meteorSpeedUpAmount = 0.1
	meteorSpeedUpTime   = 1000 * time.Millisecond
)

type GameScene struct {
	player           *Player
	baseVelocity     float64
	meteors          map[int]*Meteor
	meteorCount      int
	meteorsPerLevel  int
	meteorSpawnTimer *Timer
	velocityTimer    *Timer
}

func NewGameScence() *GameScene {
	g := &GameScene{
		baseVelocity:     baseMeteorVelocity,
		meteors:          make(map[int]*Meteor),
		meteorCount:      0,
		meteorsPerLevel:  2,
		meteorSpawnTimer: NewTimer(meteorSpawnTime),
		velocityTimer:    NewTimer(meteorSpeedUpTime),
	}
	g.player = NewPlayer(g)

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()

	g.spawnMeteors()

	for _, m := range g.meteors {
		m.Update()
	}

	g.speedUpMeteors()

	return nil
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	for _, m := range g.meteors {
		m.Draw(screen)
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
