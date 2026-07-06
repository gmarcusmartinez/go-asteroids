package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
)

func CircleFor(img *ebiten.Image, pos Vector) *resolv.Circle {
	return resolv.NewCircle(pos.X, pos.Y, float64(img.Bounds().Dx()/2))
}

func RectangleFor(img *ebiten.Image, pos Vector) *resolv.ConvexPolygon {
	b := img.Bounds()
	return resolv.NewRectangle(pos.X, pos.Y, float64(b.Dx()), float64(b.Dy()))
}

func CheckCollision(obj *resolv.Circle) bool {
	return obj.IntersectionTest(resolv.IntersectionTestSettings{
		TestAgainst: obj.SelectTouchingCells(1).FilterShapes(),
		OnIntersect: func(set resolv.IntersectionSet) bool {
			return true
		},
	})
}
