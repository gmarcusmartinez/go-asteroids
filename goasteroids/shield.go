package goasteroids

import (
	"go-asteroids/assets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

type Shield struct {
	game      *GameScene
	position  Vector
	rotation  float64
	sprite    *ebiten.Image
	shieldObj *resolv.Circle
}

func NewShield(pos Vector, rotation float64, g *GameScene) *Shield {
	/* set the sprite */
	sprite := assets.ShieldSprite

	/* position x and y coords from center of sprite */
	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	pos.X -= halfW
	pos.Y -= halfH

	/* create a shield obj */
	s := &Shield{
		game:      g,
		position:  pos,
		rotation:  rotation,
		sprite:    sprite,
		shieldObj: resolv.NewCircle(0, 0, halfW),
	}

	s.game.space.Add(s.shieldObj)

	return s

}

func (s *Shield) Update() {
	/* offset for shield */
	deltaX := float64(s.sprite.Bounds().Dx()-s.game.player.sprite.Bounds().Dx()) * 0.5
	deltaY := float64(s.sprite.Bounds().Dy()-s.game.player.sprite.Bounds().Dy()) * 0.5

	pos := Vector{
		X: s.game.player.position.X - deltaX,
		Y: s.game.player.position.Y - deltaY,
	}

	s.position = pos
	s.rotation = s.game.player.rotation
	s.shieldObj.Move(pos.X, pos.Y)

}

func (s *Shield) Draw(screen *ebiten.Image) {
	bounds := s.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(s.rotation)
	op.GeoM.Translate(halfW, halfH)
	op.GeoM.Translate(s.position.X, s.position.Y)

	screen.DrawImage(s.sprite, op)

}
