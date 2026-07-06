package engine

import "math"

type Vector struct {
	X float64
	Y float64
}

func (v Vector) Normalize() Vector {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y)
	return Vector{v.X / magnitude, v.Y / magnitude}
}

// WrapPosition wraps pos to the opposite edge when it leaves the screen.
func WrapPosition(pos Vector) Vector {
	if pos.X >= ScreenWidth {
		pos.X = 0
	} else if pos.X < 0 {
		pos.X = ScreenWidth
	}

	if pos.Y >= ScreenHeight {
		pos.Y = 0
	} else if pos.Y < 0 {
		pos.Y = ScreenHeight
	}

	return pos
}
