package main

import (
	//"fmt"

	"log"
	"runtime"
	"strconv"
	"time"

	"./cam"
	"./ctx"
	"./gfx"
	"./scr"
	"./sky"
	"./ter"
	"./veg"
	"./win"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

var mesh gfx.Mesh
var model gfx.Model
var hmap ter.HeightMap

var VIEW_DISTANCE int = 4
var LOAD_DISTANCE int = 4
var CHUNK_NB_POINTS uint32 = 512
var MULTISAMPLING int = 8
var NUM_WORKERS = 6

// PERLIN CONFIG VARS
// TODO: MOVE TO JSON AND ADD GUI

func init() {
	// GLFW event handling must be run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	log.Println(glfw.GetVersionString())

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Samples, MULTISAMPLING)

	window := win.NewWindow(ctx.Width(), ctx.Height(), "ProceduralGo - Arthur BARRIERE - Adrien BOUCAUD", false)

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.MULTISAMPLE)

	var perlin = noiselib.DefaultPerlin()
	perlin.Seed = 0 * int(time.Now().Unix())
	perlin.OctaveCount = 14
	perlin.Frequency = 0.1
	perlin.Lacunarity = 2.2
	perlin.Persistence = 0.5
	perlin.Quality = noiselib.QualitySTD

	hmap = ter.HeightMap{
		ChunkNBPoints:  CHUNK_NB_POINTS,
		ChunkWorldSize: 12,
		NbOctaves:      4,
		Exponent:       1.0,
	}

	hmap.Perlin = perlin

	hmap.TerrainType = noiselib.DefaultPerlin()
	hmap.TerrainType.Frequency = 0.05
	hmap.TerrainType.Persistence = 0.25

	hmap.MountainNoise = noiselib.DefaultRidgedmulti()
	hmap.MountainNoise.Frequency = 0.05
	hmap.MountainNoise.OctaveCount = 14

	hmap.MountainScaleBias = noiselib.DefaultScaleBias()
	hmap.MountainScaleBias.SetSourceModule(0, hmap.MountainNoise)
	hmap.MountainScaleBias.Scale = 2.3
	hmap.MountainScaleBias.Bias = 0.0

	hmap.RiverNoise = noiselib.DefaultRidgedmulti()
	hmap.RiverNoise.Seed = 3
	hmap.RiverNoise.Frequency = 0.07
	hmap.RiverNoise.Gain = 1.0
	hmap.RiverAbs = noiselib.Abs{SourceModule: make([]noiselib.Module, 1)}
	hmap.RiverAbs.SetSourceModule(0, hmap.RiverNoise)

	removeMountainRivers := noiselib.DefaultSelect()
	removeMountainRivers.SetBounds(0.3, 1000)
	removeMountainRivers.SetSourceModule(0.0, noiselib.Constant{0.0})
	removeMountainRivers.SetSourceModule(1.0, hmap.RiverNoise)
	removeMountainRivers.SetSourceModule(2.0, hmap.TerrainType)
	removeMountainRivers.SetEdgeFalloff(0.1)

	hmap.RiverClamp = noiselib.Clamp{SourceModule: make([]noiselib.Module, 1)}
	hmap.RiverClamp.SetSourceModule(0, removeMountainRivers)
	hmap.RiverClamp.SetBounds(0.4, 1.0)

	hmap.RiverScaleBias = noiselib.DefaultScaleBias()
	hmap.RiverScaleBias.SetSourceModule(0, hmap.RiverClamp)
	hmap.RiverScaleBias.Scale = -3.0

	hmap.PlainNoise = noiselib.DefaultBillow()
	hmap.PlainNoise.Frequency = 0.001

	hmap.PlainScaleBias = noiselib.DefaultScaleBias()
	hmap.PlainScaleBias.SetSourceModule(0, hmap.PlainNoise)
	hmap.PlainScaleBias.Scale = 0.125
	hmap.PlainScaleBias.Bias = 0.5

	hmap.PlainAndRiver = noiselib.DefaultSelect()
	hmap.PlainAndRiver.SetSourceModule(0, hmap.RiverScaleBias)
	hmap.PlainAndRiver.SetSourceModule(1, hmap.PlainScaleBias)
	hmap.PlainAndRiver.SetSourceModule(2, hmap.RiverClamp)
	hmap.PlainAndRiver.SetBounds(0, 0.95)
	hmap.PlainAndRiver.SetEdgeFalloff(0.5)

	hmap.FinalTerrain = noiselib.DefaultSelect()
	hmap.FinalTerrain.SetSourceModule(0, hmap.MountainScaleBias)
	//hmap.FinalTerrain.SetSourceModule(1, hmap.PlainScaleBias)
	hmap.FinalTerrain.SetSourceModule(1, hmap.PlainAndRiver)
	hmap.FinalTerrain.SetSourceModule(2, hmap.TerrainType)
	hmap.FinalTerrain.LowerBound = 0.0
	hmap.FinalTerrain.UpperBound = 1000
	hmap.FinalTerrain.SetEdgeFalloff(0.7)
	//hmap.FinalTerrain.SetEdgeFalloff(0.125)

	err := programLoop(window)
	if err != nil {
		log.Fatalln(err)
	}
}

func getCurrentChunkFromCam(camera cam.FpsCamera, hmap *ter.HeightMap) [2]int {
	x := camera.Position().X()
	z := camera.Position().Z()
	return ter.WorldToChunkCoordinates(hmap, mgl32.Vec2{x, z})
}

func programLoop(window *win.Window) error {
	programBasic, err := gfx.NewProgramFromVertFrag("basic")
	if err != nil {
		return err
	}
	defer programBasic.Delete()

	programInstances, err := gfx.NewProgramFromVertFrag("instances")
	if err != nil {
		return err
	}
	defer programInstances.Delete()

	programChunk, err := gfx.NewProgramFromVertFrag("chunk")
	if err != nil {
		return err
	}
	defer programChunk.Delete()

	programSky, err := gfx.NewProgramFromVertFrag("sky")
	if err != nil {
		return err
	}
	defer programSky.Delete()

	textureBranches, err := gfx.NewTextureFromFile("data/textures/tree/branches.png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	var leavesTextures []*gfx.Texture
	for index := 0; index < 7; index++ {
		texture, err := gfx.NewTextureFromFile("data/textures/tree/leaves"+strconv.Itoa(index+1)+".png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
		if err != nil {
			panic(err.Error())
		}
		leavesTextures = append(leavesTextures, texture)
	}

	textureSky, err := gfx.NewTextureFromFile("data/textures/sky/tint.png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}
	chunkTextures := ter.LoadChunkTextures()

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)

	camera := cam.NewFpsCamera(mgl32.Vec3{0, -5, 0}, mgl32.Vec3{0, 1, 0}, 0, 0, window.InputManager())

	currentChunk := getCurrentChunkFromCam(*camera, &hmap)

	//init lists
	var visibilityList []*ter.Chunk
	var renderList []*ter.Chunk
	var loadList []*ter.Chunk
	visibilityList = ter.GetVisibilityList(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, VIEW_DISTANCE)

	//create job queue
	loadQueue := make(chan *ter.Chunk, 1000)

	//start workers
	for i := 0; i < NUM_WORKERS; i++ {
		go ter.ChunkLoadingWorker(loadQueue, &hmap, &chunkTextures)
	}

	step := float32(hmap.ChunkWorldSize) / float32(hmap.ChunkNBPoints)
	gaia := veg.InitialiseVegetation(step)

	loadListChangeFlag := true
	currentChunkChanged := false
	dome := sky.CreateDome(programSky, gl.TEXTURE3)

	for !window.ShouldClose() {
		//OpenGL loading for new chunks

		for _, chunk := range loadList {
			if chunk.AtomicNeedOpenGLLoading == 1 && chunk.Loaded == false {
				gfx.LoadModelData(chunk.Model) //
				translate := mgl32.Translate3D(float32(chunk.Position[0])*float32(chunk.WorldSize), 0, float32(chunk.Position[1])*float32(chunk.WorldSize))
				chunk.Model.Transform = translate
				chunk.Model.Program = programChunk
				chunk.Loaded = true //should not need to change other flags if this one is set
				loadListChangeFlag = true
				gaia.CreateChunkVegetation(chunk, currentChunk)
			}
		}

		if currentChunk != getCurrentChunkFromCam(*camera, &hmap) {
			currentChunk = getCurrentChunkFromCam(*camera, &hmap)
			loadListChangeFlag = true
			currentChunkChanged = true
		}

		if loadListChangeFlag {
			loadList = ter.GetLoadList(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, LOAD_DISTANCE)
			visibilityList = ter.GetVisibilityList(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, VIEW_DISTANCE)
			//submit loading jobs
			for _, chunk := range loadList {
				if !chunk.Loaded && !chunk.Loading {
					chunk.Loading = true
					loadQueue <- chunk
				}
			}
			loadListChangeFlag = false
		}

		renderList = ter.GetRenderList(&hmap, visibilityList, *camera)

		if currentChunkChanged {
			gaia.ResetInstanceTransfoms()
			gaia.RedrawAllChunks(renderList, currentChunk)
			currentChunkChanged = false
		}

		window.StartFrame()
		camera.Update(window.SinceLastFrame())
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // depth buffer needed for DEPTH_TEST

		dome.UpdateSun(camera)

		chunkTextures.Bind()
		scr.RenderChunks(renderList, camera, programChunk, &chunkTextures, dome)
		chunkTextures.Unbind()

		for index, leavesTexture := range leavesTextures {
			leavesTexture.Bind(gl.TEXTURE10 + uint32(index))
		}
		textureBranches.Bind(gl.TEXTURE1)
		scr.RenderVegetation(gaia, camera, programInstances, dome)
		textureBranches.UnBind()
		for _, leavesTexture := range leavesTextures {
			leavesTexture.UnBind()
		}

		textureSky.Bind(gl.TEXTURE3)
		scr.RenderSky(dome, camera)
		textureSky.UnBind()
	}

	return nil
}
