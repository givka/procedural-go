package ter

import (
	"../gfx"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

type HeightMap struct {
	ChunkSize uint32
	NbOctaves uint32
	Chunks map[mgl32.Vec2]*HeightMapChunk
	Perlin noiselib.Perlin
}

type HeightMapChunk struct{
	Size 		uint32
	Position 	mgl32.Vec2
	Map  		[]float32
}

//relative coordinates
func WorldToChunkCoordinates(hmap HeightMap, world mgl32.Vec2) mgl32.Vec2{
	x := int32(world.X()) / int32(hmap.ChunkSize)
	y := int32(world.Y()) / int32(hmap.ChunkSize)
	return mgl32.Vec2{float32(x), float32(y)}
}

func ChunkToWorldCoordinates(hmap HeightMap, chunk mgl32.Vec2) mgl32.Vec2{
	x := int32(chunk.X()) * int32(hmap.ChunkSize)
	y := int32(chunk.Y()) * int32(hmap.ChunkSize)
	return mgl32.Vec2{float32(x), float32(y)}
}

func generateChunk(heightMap HeightMap, position mgl32.Vec2) *HeightMapChunk{
	chunk := HeightMapChunk{Size: heightMap.ChunkSize, Position: position}
	chunk.Map = make([]float32, chunk.Size * chunk.Size)

	posX := int32(position.X())
	posZ := int32(position.Y())

	for x := posX; x < posX + int32(chunk.Size); x++{
		for z := posZ; z < posZ + int32(chunk.Size); z++ {
			index := x - posX + (z - posZ) * int32(chunk.Size)
			chunk.Map[index] = float32(heightMap.Perlin.GetValue(float64(x)/100.0, 0, float64(z)/100.0))
		}
	}

	return &chunk
}

func GetChunk(heightMap *HeightMap, position mgl32.Vec2) *HeightMapChunk{
	if heightMap.Chunks == nil{
		//check if the map has been initialized
		heightMap.Chunks = make(map[mgl32.Vec2]*HeightMapChunk)
	}
	//check if chunk exists
	if heightMap.Chunks[position] == nil {
		//create new chunk
		heightMap.Chunks[position] = generateChunk(*heightMap, position)
	}
	return heightMap.Chunks[position]
}

func CreateChunkPolyMesh(chunk HeightMapChunk) gfx.Mesh{
	mesh := gfx.Mesh{}
	size := int(chunk.Size)

	//first add all vertices
	for x:=0; x < size; x++{
		for z:=0; z < size; z++ {
			position := mgl32.Vec3{float32(x) + chunk.Position.X(), chunk.Map[x + z * size], float32(z) + chunk.Position.Y()}
			normal := mgl32.Vec3{0.0, -1.0, 0.0}
			color := mgl32.Vec4{0.0, 0.5, 0.0, 1.0}
			texture := mgl32.Vec2{0.0, 0.0}

			v := gfx.Vertex{
				Position: 	position,
				Normal: 	normal,
				Color:		color,
				Texture: 	texture}
			mesh.Vertices = append(mesh.Vertices, v)
		}
	}

	//then build triangles
	for x:=0; x < size - 1; x++ {
		for z := 0; z < size - 1; z++ {
			i := uint32(x) + chunk.Size * uint32(z)
			tri1 := gfx.TriangleConnectivity{i, i + 1, i + uint32(chunk.Size)};
			tri2 := gfx.TriangleConnectivity{i + 1, i + uint32(chunk.Size) + 1, i + uint32(chunk.Size)};
			mesh.Connectivity = append(mesh.Connectivity, tri1)
			mesh.Connectivity = append(mesh.Connectivity, tri2)
		}
	}
	return mesh
}

/*
func createChunkCubeModel(chunk HeightMapChunk) gfx.Model{

}
*/