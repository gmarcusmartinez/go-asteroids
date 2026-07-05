package engine

import (
	"math"
	"testing"
)

func TestVectorNormalize(t *testing.T) {
	// 3-4-5 triangle: {3,4} normalizes to {0.6,0.8}.
	got := Vector{X: 3, Y: 4}.Normalize()

	if math.Abs(got.X-0.6) > 1e-9 || math.Abs(got.Y-0.8) > 1e-9 {
		t.Fatalf("Normalize() = %+v, want {0.6 0.8}", got)
	}

	if mag := math.Sqrt(got.X*got.X + got.Y*got.Y); math.Abs(mag-1) > 1e-9 {
		t.Fatalf("normalized magnitude = %v, want 1", mag)
	}
}
