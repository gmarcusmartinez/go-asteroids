package main

import (
	"go-asteroids/internal/engine"
	"go-asteroids/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("Go Asteroids")
	ebiten.SetWindowSize(engine.ScreenWidth, engine.ScreenHeight)

	err := ebiten.RunGame(&game.Game{})
	if err != nil {
		panic(err)
	}
}
