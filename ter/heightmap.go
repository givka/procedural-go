package ter

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

type HeightMap struct {
	ChunkSize uint32
	NbOctaves uint32
	Chunks map[mgl32.Vec2]*HeightMapChunk
	perlin noiselib.Perlin
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
	chunk := new(HeightMapChunk{Size: heightMap.ChunkSize, Position: position})

	posX := int32(position.X())
	posZ := int32(position.Y())

	for x := posX; x < posX + int32(*chunk.Size); x++{
		for z := posZ; z < posZ + int32(chunk.Size); z++ {
			chunk.Map[x - posX + (z - posZ) * int32(chunk.Size)] = float32(heightMap.perlin.GetValue(float64(x), 0, float64(z)))
		}
	}

	return chunk
}

func GetChunk(heightMap *HeightMap, position mgl32.Vec2) *HeightMapChunk{
	if heightMap.Chunks == nil{
		heightMap.Chunks = make(map[mgl32.Vec2]*HeightMapChunk)
	}
	//check if chunk exists
	if heightMap.Chunks[position] == nil {
		//create new chunk
		heightMap.Chunks[position] = generateChunk(*heightMap, position)
	}
	return heightMap.Chunks[position]
}