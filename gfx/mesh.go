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
	TextureID    uint32
}

type Model struct {
	VAO          uint32
	Connectivity uint32
	TextureID    uint32
	Program      *Program
	Transform    mgl32.Mat4
	NbTriangles  int32
	LoadingData  *ModelData
}

type ModelData struct {
	Vertices     []float32
	Connectivity []uint32
	TextureID    uint32
}

func FillModelData(mesh *Mesh) *ModelData {
	data := ModelData{}
	data.Vertices = make([]float32, (3+3+4+2)*len(mesh.Vertices))

	for i, vert := range mesh.Vertices {
		index := 12 * i
		data.Vertices[index] = vert.Position.X()
		data.Vertices[index+1] = vert.Position.Y() * 2
		data.Vertices[index+2] = vert.Position.Z()
		data.Vertices[index+3] = vert.Normal.X()
		data.Vertices[index+4] = vert.Normal.Y()
		data.Vertices[index+5] = vert.Normal.Z()

		data.Vertices[index+6] = vert.Color.X()
		data.Vertices[index+7] = vert.Color.Y()
		data.Vertices[index+8] = vert.Color.Z()
		data.Vertices[index+9] = vert.Color.W()

		data.Vertices[index+10] = vert.Texture.X()
		data.Vertices[index+11] = vert.Texture.Y()
	}

	data.Connectivity = make([]uint32, len(mesh.Connectivity)*3)
	for i, tri := range mesh.Connectivity {
		indice := i * 3
		data.Connectivity[indice] = tri.U0
		data.Connectivity[indice+1] = tri.U1
		data.Connectivity[indice+2] = tri.U2
	}

	data.TextureID = mesh.TextureID
	return &data
}

func LoadModelData(model *Model) {
	var VAO uint32
	var VBO uint32
	var IndexBO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)
	gl.GenBuffers(1, &IndexBO)

	floatSize := 4

	// Bind the Vertex Array Object first, then bind and set vertex buffer(s) and attribute pointers()
	gl.BindVertexArray(VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(model.LoadingData.Vertices)*floatSize, gl.Ptr(model.LoadingData.Vertices), gl.STATIC_DRAW)

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

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, IndexBO)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, (len(model.LoadingData.Connectivity) * 4), gl.Ptr(model.LoadingData.Connectivity), gl.STATIC_DRAW)

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)

	model.VAO = VAO
	model.Connectivity = IndexBO
	translate := mgl32.Translate3D(0, 0, 0)
	model.Transform = translate
	model.NbTriangles = int32(len(model.LoadingData.Connectivity))
	model.TextureID = model.LoadingData.TextureID
}

func BuildModel(mesh Mesh) Model {
	m := Model{}
	m.LoadingData = FillModelData(&mesh)
	LoadModelData(&m)
	return m
}

func ModelToInstanceModel(m *Model, transforms []mgl32.Mat4) {
	sizeFloat := 4 // 32 bits , 4 bytes
	sizeMat4 := 4 * 4 * sizeFloat
	sizeVec4 := 4 * sizeFloat

	var VBO uint32
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(transforms)*sizeMat4, gl.Ptr(transforms), gl.STATIC_DRAW)

	VAO := m.VAO
	gl.BindVertexArray(VAO)

	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointer(4, 4, gl.FLOAT, false, int32(4*sizeVec4), gl.PtrOffset(0))
	gl.EnableVertexAttribArray(5)
	gl.VertexAttribPointer(5, 4, gl.FLOAT, false, int32(4*sizeVec4), gl.PtrOffset((sizeVec4)))
	gl.EnableVertexAttribArray(6)
	gl.VertexAttribPointer(6, 4, gl.FLOAT, false, int32(4*sizeVec4), gl.PtrOffset((2 * sizeVec4)))
	gl.EnableVertexAttribArray(7)
	gl.VertexAttribPointer(7, 4, gl.FLOAT, false, int32(4*sizeVec4), gl.PtrOffset((3 * sizeVec4)))

	gl.VertexAttribDivisor(4, 1)
	gl.VertexAttribDivisor(5, 1)
	gl.VertexAttribDivisor(6, 1)
	gl.VertexAttribDivisor(7, 1)

	gl.BindVertexArray(0)
}
