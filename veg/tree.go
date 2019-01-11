package veg

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

var cubePositions [][3]float32

type Tree struct {
}

func CreateTree() {
	renderTrunk(1.0, 5.0)

}

func renderTrunk(radius float64, height float64) {
	stepThetha := 360.0 / 100.0
	stepHeight := height / 100.0

	for h := 0.0; h < height; h += stepHeight {
		for theta := 0.0; theta <= 360.0; theta += stepThetha {
			x := radius * math.Cos(theta)
			y := h
			z := radius * math.Sin(theta)
			pos := mgl32.Vec3{float32(x), float32(y), float32(z)}
			cubePositions = append(cubePositions, [3]float32{float32(pos[0]), float32(pos[1]), float32(pos[2])})
		}
	}

}

func CubePositions() [][3]float32 {
	return cubePositions
}

func rotateX(angleDegree float32, original mgl32.Vec3) mgl32.Vec3 {
	angle := (float32(math.Pi) * angleDegree) / 180.0
	return mgl32.Rotate3DX(angle).Mul3x1(original)
}

func rotateZ(angleDegree float32, original mgl32.Vec3) mgl32.Vec3 {
	angle := (float32(math.Pi) * angleDegree) / 180.0
	return mgl32.Rotate3DZ(angle).Mul3x1(original)
}
