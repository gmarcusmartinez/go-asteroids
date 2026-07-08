package engine

import "github.com/hajimehoshi/ebiten/v2"

func CenterSprite(center Vector, img *ebiten.Image) Vector {
	bounds := img.Bounds()
	return Vector{
		X: center.X - float64(bounds.Dx())/2,
		Y: center.Y - float64(bounds.Dy())/2,
	}
}

func DrawSprite(screen, img *ebiten.Image, pos Vector, rotation float64) {
	bounds := img.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(halfW, halfH)
	op.GeoM.Translate(pos.X, pos.Y)

	screen.DrawImage(img, op)
}
