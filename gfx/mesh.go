package gfx

import "github.com/go-gl/mathgl/mgl32"

type Vertex struct{
	Position 	mgl32.Vec3
	Normal 		mgl32.Vec3
	Color 		mgl32.Vec4
	Texture 	mgl32.Vec2
}

type TriangleConnectivity struct{
	U0 uint
	U1 uint
	U2 uint
}

type Mesh struct {
	Test int
	Vertices []Vertex
	Connectivity []TriangleConnectivity
}

type Model struct {
	VAO uint32
	TextureID uint32
	Program *Program
	Transform mgl32.Mat4
	NbTriangles int32
}

func NewVertex() *Vertex{
	v := new(Vertex)
	v.Position = mgl32.Vec3{0.0, 0.0, 0.0}
	v.Normal = mgl32.Vec3{0.0, 0.0, 0.0}
	v.Color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	v.Texture = mgl32.Vec2{0.0, 0.0}
	return v
}


