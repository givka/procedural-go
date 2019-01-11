package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Vertex struct {
	Position mgl32.Vec3
	Normal   mgl32.Vec3
	Color    mgl32.Vec4
	Texture  mgl32.Vec2
}

type TriangleConnectivity struct {
	U0 uint32
	U1 uint32
	U2 uint32
}

type Mesh struct {
	Test         int
	Vertices     []Vertex
	Connectivity []TriangleConnectivity
}

type Model struct {
	VAO          uint32
	Connectivity uint32
	Indices      []uint32
	TextureID    uint32
	Program      *Program
	Transform    mgl32.Mat4
	NbTriangles  int32
}

func BuildModel(mesh Mesh) Model {
	model := Model{}
	vertArray := make([]float32, (3+3+4+2)*len(mesh.Vertices))
	floatSize := 4 //float32 -> 4 bytes

	for i, vert := range mesh.Vertices {
		index := 12 * i
		vertArray[index] = vert.Position.X()
		vertArray[index+1] = vert.Position.Y() * 2
		vertArray[index+2] = vert.Position.Z()
		vertArray[index+3] = vert.Normal.X()
		vertArray[index+4] = vert.Normal.Y()
		vertArray[index+5] = vert.Normal.Z()

		vertArray[index+6] = vert.Color.X()
		vertArray[index+7] = vert.Color.Y()
		vertArray[index+8] = vert.Color.Z()
		vertArray[index+9] = vert.Color.W()

		vertArray[index+10] = vert.Texture.X()
		vertArray[index+11] = vert.Texture.Y()

	}

	var VAO uint32
	var VBO uint32
	var IndexBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &IndexBO)

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertArray)*floatSize, gl.Ptr(vertArray), gl.STATIC_DRAW)

	var stride int32 = int32(floatSize * (3 + 3 + 4 + 2)) //pos + norm + col + tex
	var offset int = 0

	//set attribs
	{
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
		gl.EnableVertexAttribArray(0)
		offset += 3 * floatSize
		//normal
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(offset))
		gl.EnableVertexAttribArray(1)
		offset += 3 * floatSize
		//color
		gl.VertexAttribPointer(2, 4, gl.FLOAT, false, stride, gl.PtrOffset(offset))
		gl.EnableVertexAttribArray(2)
		offset += 4 * floatSize
		//texture
		gl.VertexAttribPointer(3, 2, gl.FLOAT, false, stride, gl.PtrOffset(offset))
		gl.EnableVertexAttribArray(3)
		offset += 2 * floatSize
	}

	gl.BindVertexArray(0)
	//connectivity
	connectivity := make([]uint32, len(mesh.Connectivity)*3)
	for i, tri := range mesh.Connectivity {
		indice := i * 3
		connectivity[indice] = tri.U0
		connectivity[indice+1] = tri.U1
		connectivity[indice+2] = tri.U2
	}
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, IndexBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(mesh.Connectivity)*3*4, gl.Ptr(connectivity), gl.STATIC_DRAW)

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	model.Indices = connectivity
	model.VAO = VAO
	model.Connectivity = IndexBO
	translate := mgl32.Translate3D(0, 0, 0)
	model.Transform = translate
	model.NbTriangles = int32(len(model.Indices))
	model.TextureID = 0

	return model
}

func NewVertex() *Vertex {
	v := new(Vertex)
	v.Position = mgl32.Vec3{0.0, 0.0, 0.0}
	v.Normal = mgl32.Vec3{0.0, 0.0, 0.0}
	v.Color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
	v.Texture = mgl32.Vec2{0.0, 0.0}
	return v
}
