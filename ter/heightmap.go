package ter

import (
	"../gfx"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

type HeightMap struct {
	ChunkNBPoints  uint32
	ChunkWorldSize uint32
	NbOctaves      uint32
	Chunks         map[mgl32.Vec2]*HeightMapChunk
	Perlin         noiselib.Perlin
}

type HeightMapChunk struct{
	NBPoints  uint32
	WorldSize uint32
	Position  mgl32.Vec2
	Map       []float32
	Model     *gfx.Model
}

//relative coordinates
func WorldToChunkCoordinates(hmap *HeightMap, world mgl32.Vec2) mgl32.Vec2{
	x := int32(world.X()) / int32(hmap.ChunkWorldSize)
	y := int32(world.Y()) / int32(hmap.ChunkWorldSize)
	return mgl32.Vec2{float32(x), float32(y)}
}

func ChunkToWorldCoordinates(hmap *HeightMap, chunk mgl32.Vec2) mgl32.Vec2{
	x := int32(chunk.X()) * int32(hmap.ChunkWorldSize)
	y := int32(chunk.Y()) * int32(hmap.ChunkWorldSize)
	return mgl32.Vec2{float32(x), float32(y)}
}

func generateChunk(heightMap *HeightMap, position mgl32.Vec2) *HeightMapChunk{
	chunk := HeightMapChunk{NBPoints: heightMap.ChunkNBPoints, WorldSize: heightMap.ChunkWorldSize, Position: position}
	chunk.Map = make([]float32, (chunk.NBPoints+1)*(chunk.NBPoints+1))

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)

	for x := 0; x < int(chunk.NBPoints) + 1; x ++{
		for z := 0; z < int(chunk.NBPoints) + 1; z++{
			index := x + z * int(chunk.NBPoints+1)
			posX := position.X() * float32(chunk.WorldSize) + float32(x) * step
			posZ := position.Y() * float32(chunk.WorldSize) + float32(z) * step
			chunk.Map[index] = float32(heightMap.Perlin.GetValue(float64(posX), 0, float64(posZ)))
		}
	}
	mesh := CreateChunkPolyMesh(chunk)
	model := gfx.BuildModel(mesh)
	translate := mgl32.Translate3D(position.X()*float32(chunk.WorldSize), 0, position.Y()*float32(chunk.WorldSize))
	model.Transform = translate
	chunk.Model = &model
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
		heightMap.Chunks[position] = generateChunk(heightMap, position)
	}
	return heightMap.Chunks[position]
}

func CreateChunkPolyMesh(chunk HeightMapChunk) gfx.Mesh{
	mesh := gfx.Mesh{}
	size := int(chunk.NBPoints)

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)

	//first add all vertices
	for x:=0; x < size+1; x++{
		for z:=0; z < size+1; z++ {
			position := mgl32.Vec3{float32(x) * step, chunk.Map[x + z * (size+1)], float32(z) * step}
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
	for x:=0; x < size; x++ {
		for z := 0; z < size; z++ {
			i := uint32(x) + (chunk.NBPoints+1) * uint32(z)
			tri1 := gfx.TriangleConnectivity{i, i + 1, i + uint32(chunk.NBPoints+1)};
			tri2 := gfx.TriangleConnectivity{i + 1, i + uint32(chunk.NBPoints+1) + 1, i + uint32(chunk.NBPoints+1)};
			mesh.Connectivity = append(mesh.Connectivity, tri1)
			mesh.Connectivity = append(mesh.Connectivity, tri2)
		}
	}
	return mesh
}

func GetSurroundingChunks(hmap *HeightMap, worldPosition mgl32.Vec2, size uint) []*HeightMapChunk{
	chunkPosition := WorldToChunkCoordinates(hmap, worldPosition)
	startPosition := chunkPosition.Sub(mgl32.Vec2{float32(size / 2), float32(size/2)})
	endPosition := chunkPosition.Add(mgl32.Vec2{float32(size / 2), float32(size/2)})

	chunks := []*HeightMapChunk{}

	fmt.Println(startPosition, endPosition)
	for x := startPosition.X(); x < endPosition.X(); x++{
		for y:= startPosition.Y(); y < endPosition.Y(); y++{
			key := mgl32.Vec2{x, y}
			chunks = append(chunks, GetChunk(hmap, key))
		}
	}
	return chunks
}

/*
func createChunkCubeModel(chunk HeightMapChunk) gfx.Model{

}
*/