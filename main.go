package main

import (
	"go-asteroids/goasteroids"
	"go-asteroids/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowTitle("Go Asteroids")
	ebiten.SetWindowSize(engine.ScreenWidth, engine.ScreenHeight)

	err := ebiten.RunGame(&goasteroids.Game{})
	if err != nil {
		panic(err)
	}
}
