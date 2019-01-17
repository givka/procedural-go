package ter

import (
	"../cam"
	"../gfx"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"sync/atomic"
)

type Chunk struct {
	NBPoints  uint32
	WorldSize uint32
	Position  [2]int
	Map       []float64
	Model     *gfx.Model

	//loading related flags
	Loaded                  bool
	Loading                 bool
	AtomicNeedOpenGLLoading int32
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

func GetVisibilityList(heightMap *HeightMap, worldPos mgl32.Vec2, radius int) []*Chunk{
	if heightMap.Chunks == nil {
		//check if the map has been initialized
		heightMap.Chunks = make(map[[2]int]*Chunk)
	}

	chunks := []*Chunk{}
	chunkPos := WorldToChunkCoordinates(heightMap, worldPos)

	rsquared := radius * radius //precompute
	for x := -radius; x <= radius; x++{
		for y := -radius; y <= radius; y++{
			coord := [2]int{x + chunkPos[0], y + chunkPos[1]}
			if x*x + y*y <= rsquared{
				if heightMap.Chunks[coord] == nil {
					chunk := Chunk{NBPoints: heightMap.ChunkNBPoints, WorldSize: heightMap.ChunkWorldSize, Position: coord, Loading: false, Loaded: false, AtomicNeedOpenGLLoading: 0}
					heightMap.Chunks[coord] = &chunk}
					chunks = append(chunks, heightMap.Chunks[coord])
				}
			}
		}

	return chunks
}

func GetRenderList(heightMap *HeightMap, visList []*Chunk, camera cam.FpsCamera) []*Chunk{
	renderList := []*Chunk{}
	for _, chunk := range visList{
		//TODO: some form of frustum culling
		if chunk.Loaded {
			renderList = append(renderList, chunk)
		}
	}
	return renderList
}

func GetLoadList(heightMap *HeightMap, worldPos mgl32.Vec2, radius int) []*Chunk{
	radiusChunks := GetVisibilityList(heightMap, worldPos, radius)
	loadList := []*Chunk{}
	for _, chunk := range radiusChunks{
		if !chunk.Loaded {
			loadList = append(loadList, chunk)
		}
	}

	return loadList
}

func LoadChunk(chunk *Chunk, heightMap *HeightMap){

	if chunk.Loaded {
		return
	}
	//fill up heightmap
	chunk.Map = make([]float64, (chunk.NBPoints+1)*(chunk.NBPoints+1))
	step := float64(chunk.WorldSize) / float64(chunk.NBPoints)
	position := chunk.Position
	//float conversions before loop
	var posf = [2]float64{float64(position[0]), float64(position[1])}
	var worldSizef = float64(chunk.WorldSize)

	for x := 0; x < int(chunk.NBPoints) + 1; x ++{
		for z := 0; z < int(chunk.NBPoints) + 1; z++{
			index := x + z * int(chunk.NBPoints+1)
			posX := (posf[0]) * worldSizef + float64(x) * step
			posZ := (posf[1]) * worldSizef + float64(z) * step

			chunk.Map[index] = heightMap.FinalTerrain.GetValue(float64(posX), 0, float64(posZ))
		}
	}

	//build mesh
	mesh := CreateChunkPolyMesh(*chunk)
	//build model's vertex and connectivity arrays
	chunk.Model = new (gfx.Model)
	chunk.Model.LoadingData = gfx.FillModelData(&mesh)
	//Chunk loaded. Only opengl loading left.
}

func ChunkLoadingWorker(chunks <-chan *Chunk, heightMap *HeightMap){
	fmt.Println("Starting worker")
	for chunk := range chunks {
		LoadChunk(chunk, heightMap)
		atomic.StoreInt32(&chunk.AtomicNeedOpenGLLoading, 1) //flag as loaded
	}
}
