package entity

import (
	"go-asteroids/assets"
	"go-asteroids/internal/engine"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

type Shield struct {
	player   *Player
	position engine.Vector
	rotation float64
	sprite   *ebiten.Image
	Obj      *resolv.Circle
}

func NewShield(player *Player) *Shield {
	/* set the sprite */
	sprite := assets.ShieldSprite

	bounds := sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2

	/* create a shield obj (space registration and per-frame positioning
	   happen in Scene.SetShield and Update respectively) */
	return &Shield{
		player:   player,
		rotation: player.Rotation,
		sprite:   sprite,
		Obj:      resolv.NewCircle(0, 0, halfW),
	}
}

func (s *Shield) Update() {
	/* offset for shield */
	deltaX := float64(s.sprite.Bounds().Dx()-s.player.Sprite.Bounds().Dx()) * 0.5
	deltaY := float64(s.sprite.Bounds().Dy()-s.player.Sprite.Bounds().Dy()) * 0.5

	pos := engine.Vector{
		X: s.player.Position.X - deltaX,
		Y: s.player.Position.Y - deltaY,
	}

	s.position = pos
	s.rotation = s.player.Rotation
	s.Obj.Move(pos.X, pos.Y)

}

func (s *Shield) Draw(screen *ebiten.Image) {
	engine.DrawSprite(screen, s.sprite, s.position, s.rotation)
}
