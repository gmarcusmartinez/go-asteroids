package goasteroids

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
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
	space            *resolv.Space
	lasers           map[int]*Laser
	laserCount       int
}

func NewGameScence() *GameScene {
	g := &GameScene{
		baseVelocity:     baseMeteorVelocity,
		meteors:          make(map[int]*Meteor),
		meteorCount:      0,
		meteorsPerLevel:  2,
		meteorSpawnTimer: NewTimer(meteorSpawnTime),
		velocityTimer:    NewTimer(meteorSpeedUpTime),
		space:            resolv.NewSpace(ScreenWidth, ScreenHeight, 16, 16),
		lasers:           make(map[int]*Laser),
		laserCount:       0,
	}

	g.player = NewPlayer(g)
	g.space.Add(g.player.playerObj)

	return g
}

func (g *GameScene) Update(state *State) error {
	g.player.Update()

	g.spawnMeteors()

	for _, m := range g.meteors {
		m.Update()
	}

	for _, l := range g.lasers {
		l.Update()
	}

	g.speedUpMeteors()
	g.isPlayerCollidingWithMeteor()

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
			data := m.meteorObj.Data().(*ObjectData)
			fmt.Println("player collided with meteor", data.index)
		}
	}
}
