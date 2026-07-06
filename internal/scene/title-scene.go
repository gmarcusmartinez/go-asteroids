package scene

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"
	"go-asteroids/internal/entity"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type TitleScene struct {
	meteors     map[int]*entity.Meteor
	meteorCount int
	stars       []*entity.Star
}

func NewTitleScene() *TitleScene {
	return &TitleScene{
		meteors: make(map[int]*entity.Meteor),
		stars:   entity.GenerateStars(numberOfStars),
	}
}

func (t *TitleScene) Draw(screen *ebiten.Image) {
	/* draw stars */
	for _, s := range t.stars {
		s.Draw(screen)
	}

	textToDraw := "Welcome to Hell"

	op := &text.DrawOptions{
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}

	op.ColorScale.ScaleWithColor(color.White)

	op.GeoM.Translate(float64(engine.ScreenWidth/2), engine.ScreenHeight-200)
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
		m := entity.NewMeteor(0.25, len(t.meteors)-1)
		t.meteorCount++
		t.meteors[t.meteorCount] = m
	}
	for _, m := range t.meteors {
		m.Update()
	}

	return nil
}
