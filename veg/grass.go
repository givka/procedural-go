package veg

import (
	"math/rand"

	"../gfx"
	"github.com/go-gl/mathgl/mgl32"
)

type InstanceGrass struct {
	Model      *gfx.Model
	Transforms []mgl32.Mat4
}

func CreateUniqueGrass(step float32) *gfx.Model {
	mesh := gfx.Mesh{}
	width := step
	height := float32(-0.050)
	index := uint32(0)
	angle := 360.0 * rand.Float32()
	p1 := rotateY(angle, mgl32.Vec3{0, 0, 0.0})
	p2 := rotateY(angle, mgl32.Vec3{width / 2.0, height, 0.0})
	p3 := rotateY(angle, mgl32.Vec3{width, 0, 0.0})
	U := p2.Sub(p1)
	V := p3.Sub(p1)
	normal := U.Cross(V)
	normal = normal.Normalize()
	normal = mgl32.Vec3{0, -2, 0}
	color := mgl32.Vec4{0.0, 0.5, 0.0, 1.0}
	texture := mgl32.Vec2{0.0, 0.0}
	mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p1, Normal: normal, Color: color, Texture: texture})
	mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p2, Normal: normal, Color: color, Texture: texture})
	mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p3, Normal: normal, Color: color, Texture: texture})
	t1 := gfx.TriangleConnectivity{U0: index + 0, U1: index + 1, U2: index + 2}
	mesh.Connectivity = append(mesh.Connectivity, t1)

	model := gfx.BuildModel(mesh)
	return &model
}
