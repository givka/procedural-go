package ter

import (
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"

	"../cam"
	"../gfx"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Chunk struct {
	NBPoints        uint32
	WorldSize       uint32
	Position        [2]int
	Map             []float64
	Model           *gfx.Model
	GrassTransforms []mgl32.Mat4
	TreesTransforms []mgl32.Mat4
	TreesModelID    []int
	IsHQ            bool
	HasVegetation   bool

	//loading related flags
	Loaded                  bool
	Loading                 bool
	AtomicNeedOpenGLLoading int32
}

type ChunkTextureContainer struct {
	Dirt  *gfx.Texture
	Sand  *gfx.Texture
	Snow  *gfx.Texture
	Grass *gfx.Texture
	Rock  *gfx.Texture

	DirtID  uint32
	SandID  uint32
	SnowID  uint32
	GrassID uint32
	RockID  uint32
}

func LoadChunkTextures() ChunkTextureContainer {
	container := ChunkTextureContainer{}
	var err error

	container.Dirt, err = gfx.NewTextureFromFile("data/textures/chunks/dirt.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	container.Snow, err = gfx.NewTextureFromFile("data/textures/chunks/snow.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	container.Grass, err = gfx.NewTextureFromFile("data/textures/chunks/grass.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	container.Rock, err = gfx.NewTextureFromFile("data/textures/chunks/rock.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	container.Sand, err = gfx.NewTextureFromFile("data/textures/chunks/sand.jpg", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	container.DirtID = gl.TEXTURE3
	container.SandID = gl.TEXTURE4
	container.SnowID = gl.TEXTURE5
	container.GrassID = gl.TEXTURE6
	container.RockID = gl.TEXTURE7
	return container
}

func (container *ChunkTextureContainer) Bind() {
	container.Dirt.Bind(container.DirtID)
	container.Sand.Bind(container.SandID)
	container.Snow.Bind(container.SnowID)
	container.Grass.Bind(container.GrassID)
	container.Rock.Bind(container.RockID)
}

func (container *ChunkTextureContainer) Unbind() {
	container.Dirt.UnBind()
	container.Sand.UnBind()
	container.Snow.UnBind()
	container.Grass.UnBind()
	container.Rock.UnBind()
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

func GetVisibilityList(heightMap *HeightMap, worldPos mgl32.Vec2, radius int) []*Chunk {
	if heightMap.Chunks == nil {
		//check if the map has been initialized
		heightMap.Chunks = make(map[[2]int]*Chunk)
	}

	chunks := []*Chunk{}
	chunkPos := WorldToChunkCoordinates(heightMap, worldPos)

	rsquared := radius * radius //precompute
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			coord := [2]int{x + chunkPos[0], y + chunkPos[1]}
			if x*x+y*y <= rsquared {
				if heightMap.Chunks[coord] == nil {
					chunk := Chunk{NBPoints: heightMap.ChunkNBPoints, WorldSize: heightMap.ChunkWorldSize, Position: coord, Loading: false, Loaded: false, AtomicNeedOpenGLLoading: 0}
					heightMap.Chunks[coord] = &chunk
				}
				chunks = append(chunks, heightMap.Chunks[coord])
			}
		}
	}

	return chunks
}

func GetRenderList(heightMap *HeightMap, visList []*Chunk, camera cam.FpsCamera) []*Chunk {
	renderList := []*Chunk{}
	for _, chunk := range visList {
		//TODO: some form of frustum culling
		if chunk.Loaded {
			renderList = append(renderList, chunk)
		}
	}
	return renderList
}

func GetLoadList(heightMap *HeightMap, worldPos mgl32.Vec2, radius int) []*Chunk {
	radiusChunks := GetVisibilityList(heightMap, worldPos, radius)
	loadList := []*Chunk{}
	for _, chunk := range radiusChunks {
		if !chunk.Loaded {
			loadList = append(loadList, chunk)
		}
	}

	return loadList
}

func ChunkLoadingWorker(chunks <-chan *Chunk, heightMap *HeightMap, textureContainer *ChunkTextureContainer) {
	fmt.Println("Starting worker")
	for chunk := range chunks {
		LoadChunk(chunk, heightMap, textureContainer)
		atomic.StoreInt32(&chunk.AtomicNeedOpenGLLoading, 1) //flag as loaded
	}
}

func LoadChunk(chunk *Chunk, heightMap *HeightMap, textureContainer *ChunkTextureContainer) {

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

	for x := 0; x < int(chunk.NBPoints)+1; x++ {
		for z := 0; z < int(chunk.NBPoints)+1; z++ {
			index := x + z*int(chunk.NBPoints+1)
			posX := (posf[0])*worldSizef + float64(x)*step
			posZ := (posf[1])*worldSizef + float64(z)*step

			chunk.Map[index] = heightMap.FinalTerrain.GetValue(float64(posX), 0, float64(posZ))
		}
	}

	//build mesh
	mesh := CreateChunkPolyMesh(*chunk, textureContainer, heightMap)
	//build model's vertex and connectivity arrays
	chunk.Model = new(gfx.Model)
	chunk.Model.LoadingData = gfx.FillModelData(&mesh)

	chunk.GrassTransforms = getGrassTransforms(chunk)
	chunk.TreesTransforms = getTreesTransforms(chunk)

	//Chunk loaded. Only opengl loading left.
}

// TODO: isHQ ? transform = transform.Mul4(mgl32.Scale3D(5, 5, 5)) : nil
func getTreesTransforms(chunk *Chunk) []mgl32.Mat4 {
	var transforms []mgl32.Mat4

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)
	angle := float32(5.0 * math.Cos(glfw.GetTime()))
	for x := 0; x < int(chunk.NBPoints)+1; x += int(chunk.NBPoints / 32) {
		for z := 0; z < int(chunk.NBPoints)+1; z += int(chunk.NBPoints / 32) {
			i := x + z*int(chunk.NBPoints+1)
			posY := float32(chunk.Map[i])
			if posY < 0.0 || posY > 0.10 {
				continue
			}
			posX := float32(chunk.Position[0])*float32(chunk.WorldSize) + float32(x)*step
			posZ := float32(chunk.Position[1])*float32(chunk.WorldSize) + float32(z)*step
			transform := mgl32.Translate3D(posX, -2*posY, posZ).Mul4(mgl32.Rotate3DY(posY * 360.0).Mat4())
			transform = transform.Mul4(mgl32.Rotate3DX(mgl32.DegToRad(angle)).Mat4())

			transforms = append(transforms, transform)
		}
	}
	return transforms
}

func getGrassTransforms(chunk *Chunk) []mgl32.Mat4 {
	var transforms []mgl32.Mat4

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)
	for x := 0; x < int(chunk.NBPoints)+1; x++ {
		for z := 0; z < int(chunk.NBPoints)+1; z++ {
			i := x + z*int(chunk.NBPoints+1)
			posY := float32(chunk.Map[i])
			if posY < 0.20 || posY > 0.30 {
				continue
			}
			posX := float32(chunk.Position[0])*float32(chunk.WorldSize) + float32(x)*step
			posZ := float32(chunk.Position[1])*float32(chunk.WorldSize) + float32(z)*step
			transform := mgl32.Translate3D(posX, -2*posY, posZ).Mul4(mgl32.Rotate3DY(360.0 * rand.Float32()).Mat4())
			transforms = append(transforms, transform)
		}
	}
	return transforms
}
