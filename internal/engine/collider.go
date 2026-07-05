package engine

import "github.com/solarlune/resolv"

func CheckCollision(obj *resolv.Circle) bool {
	return obj.IntersectionTest(resolv.IntersectionTestSettings{
		TestAgainst: obj.SelectTouchingCells(1).FilterShapes(),
		OnIntersect: func(set resolv.IntersectionSet) bool {
			return true
		},
	})
}
