package ter

import (
	"fmt"

	"../gfx"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

type HeightMap struct {
	ChunkNBPoints  uint32
	ChunkWorldSize uint32
	NbOctaves      uint32
	Chunks         map[[2]int]*HeightMapChunk
	Perlin         noiselib.Perlin
}

type HeightMapChunk struct {
	NBPoints  uint32
	WorldSize uint32
	Position  [2]int
	Map       []float32
	Model     *gfx.Model
}

//relative coordinates
func WorldToChunkCoordinates(hmap *HeightMap, world mgl32.Vec2) [2]int {
	x := int(world.X()) / int(hmap.ChunkWorldSize)
	y := int(world.Y()) / int(hmap.ChunkWorldSize)
	return [2]int{x, y}
}

func ChunkToWorldCoordinates(hmap *HeightMap, chunk [2]int) mgl32.Vec2 {
	x := int(chunk[0]) * int(hmap.ChunkWorldSize)
	y := int(chunk[1]) * int(hmap.ChunkWorldSize)
	return mgl32.Vec2{float32(x), float32(y)}
}

func generateChunk(heightMap *HeightMap, position [2]int) *HeightMapChunk {
	chunk := HeightMapChunk{NBPoints: heightMap.ChunkNBPoints, WorldSize: heightMap.ChunkWorldSize, Position: position}
	chunk.Map = make([]float32, (chunk.NBPoints+1)*(chunk.NBPoints+1))

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)

	for x := 0; x < int(chunk.NBPoints) + 1; x ++{
		for z := 0; z < int(chunk.NBPoints) + 1; z++{
			index := x + z * int(chunk.NBPoints+1)
			posX := float32(position[0]) * float32(chunk.WorldSize) + float32(x) * step
			posZ := float32(position[1]) * float32(chunk.WorldSize) + float32(z) * step
			chunk.Map[index] = float32(
				heightMap.Perlin.GetValue(float64(posX), 0, float64(posZ)) )//+
		}
	}
	mesh := CreateChunkPolyMesh(chunk)
	model := gfx.BuildModel(mesh)
	translate := mgl32.Translate3D(float32(position[0])*float32(chunk.WorldSize), 0, float32(position[1])*float32(chunk.WorldSize))
	model.Transform = translate
	chunk.Model = &model
	return &chunk
}

func GetChunk(heightMap *HeightMap, position [2]int) *HeightMapChunk {
	if heightMap.Chunks == nil {
		//check if the map has been initialized
		heightMap.Chunks = make(map[[2]int]*HeightMapChunk)
	}
	//check if chunk exists
	if heightMap.Chunks[position] == nil {
		//create new chunk
		heightMap.Chunks[position] = generateChunk(heightMap, position)
	}
	return heightMap.Chunks[position]
}

func CreateChunkPolyMesh(chunk HeightMapChunk) gfx.Mesh {
	mesh := gfx.Mesh{}
	size := int(chunk.NBPoints)

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)

	//first add all vertices

	for x:=0; x < size+1; x++{
		for z:=0; z < size+1; z++ {
			position := mgl32.Vec3{float32(x) * step, chunk.Map[x + z * (size+1)], float32(z) * step}
			//compute normal
			var up float32 = 0.0
			var down float32 = 0.0
			var left float32 = 0.0
			var right float32 = 0.0
			if z > 0 		{up 	= chunk.Map[x + (z-1) * (size+1)]}
			if z < size		{down	= chunk.Map[x + (z+1) * (size+1)]}
			if x > 0 		{left 	= chunk.Map[x - 1 + z * (size+1)]}
			if x < size 	{right	= chunk.Map[x + 1 + z * (size+1)]}

			normal := mgl32.Vec3{right - left, -2, down - up}
			normal = normal.Normalize()

			color := mgl32.Vec4{0.0, 0.5, 0.0, 1.0}
			texture := mgl32.Vec2{0.0, 0.0}

			v := gfx.Vertex{
				Position: position,
				Normal:   normal,
				Color:    color,
				Texture:  texture}
			mesh.Vertices = append(mesh.Vertices, v)
		}
	}

	//then build triangles
	for x := 0; x < size; x++ {
		for z := 0; z < size; z++ {
			i := uint32(x) + (chunk.NBPoints+1)*uint32(z)
			tri1 := gfx.TriangleConnectivity{i, i + 1, i + uint32(chunk.NBPoints+1)}
			tri2 := gfx.TriangleConnectivity{i + 1, i + uint32(chunk.NBPoints+1) + 1, i + uint32(chunk.NBPoints+1)}
			mesh.Connectivity = append(mesh.Connectivity, tri1)
			mesh.Connectivity = append(mesh.Connectivity, tri2)
		}
	}

	return mesh
}

func GetSurroundingChunks(hmap *HeightMap, worldPosition mgl32.Vec2, size uint) []*HeightMapChunk {
	chunkPosition := WorldToChunkCoordinates(hmap, worldPosition)

	startPosition := [2]int{chunkPosition[0] - int(size/2), chunkPosition[1] - int(size/2)}
	endPosition := [2]int{chunkPosition[0] + int(size/2), chunkPosition[1] + int(size/2)}

	chunks := []*HeightMapChunk{}

	fmt.Println(startPosition, endPosition)
	for x := startPosition[0]; x < endPosition[0]; x++ {
		for y := startPosition[1]; y < endPosition[1]; y++ {
			key := [2]int{x, y}
			chunks = append(chunks, GetChunk(hmap, key))
		}
	}
	return chunks
}

/*
func createChunkCubeModel(chunk HeightMapChunk) gfx.Model{

}
*/
