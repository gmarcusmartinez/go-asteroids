package goasteroids

import (
	"go-asteroids/assets"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type TitleScene struct {
	meteors     map[int]*Meteor
	meteorCount int
	stars       []*Star
}

var highScore int
var originalHighScore int

func (t *TitleScene) Draw(screen *ebiten.Image) {
	/* draw stars */
	for _, s := range t.stars {
		s.Draw(screen)
	}

	textToDraw := "Wellcome to Hell"

	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}

	op.ColorScale.ScaleWithColor(color.White)

	op.GeoM.Translate(float64(ScreenWidth/2), ScreenHeight-200)
	text.Draw(screen, textToDraw, &text.GoTextFace{
		Source: assets.TitleFont,
		Size:   48,
	}, op)

	/* draw meteors */
	for _, m := range t.meteors {
		m.Draw(screen)
	}

}

func (t *TitleScene) Update(state *State) error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		state.SceneManager.GoToScene(NewGameScene())
		return nil
	}

	/* add some meteors */
	if len(t.meteors) < 10 {
		m := NewMeteor(0.25, &GameScene{}, len(t.meteors)-1)
		t.meteorCount++
		t.meteors[t.meteorCount] = m
	}
	for _, m := range t.meteors {
		m.Update()
	}

	return nil
}

func init() {
	hs, err := getHighScore()
	if err != nil {
		log.Println("Error getting high score", err)
	}

	highScore = hs
	originalHighScore = hs
}
