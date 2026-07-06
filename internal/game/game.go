package game

import (
	"go-asteroids/internal/engine"
	"go-asteroids/internal/scene"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	sceneManager *scene.SceneManager
	input        scene.Input
}

func (g *Game) Update() error {
	if g.sceneManager == nil {
		g.sceneManager = &scene.SceneManager{}
		g.sceneManager.GoToScene(scene.NewTitleScene())
	}

	g.input.Update()
	if err := g.sceneManager.Update(&g.input); err != nil {
		return err
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sceneManager.Draw(screen)
}

func (g *Game) Layout(_, _ int) (width, height int) {
	return engine.ScreenWidth, engine.ScreenHeight
}
