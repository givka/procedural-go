package main

import (
	//"fmt"

	"log"
	"runtime"
	"time"

	"./cam"
	"./ctx"
	"./gfx"
	"./scr"
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

var trees []*veg.Tree

var chunks []*ter.Chunk

var NBChunks uint = 2

var VIEW_DISTANCE = 4
var LOAD_DISTANCE = 4
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

	window := win.NewWindow(ctx.Width(), ctx.Height(), "ProceduralGo - Arthur BARRIERE - Adrien BOUCAUD", false)

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var perlin = noiselib.DefaultPerlin()
	perlin.Seed = 0 * int(time.Now().Unix())
	perlin.OctaveCount = 14
	perlin.Frequency = 0.1
	perlin.Lacunarity = 2.2
	perlin.Persistence = 0.5
	perlin.Quality = noiselib.QualitySTD

	hmap = ter.HeightMap{
		ChunkNBPoints:  512,
		ChunkWorldSize: 12,
		NbOctaves:      4,
		Exponent:       1.0,
	}

	hmap.Perlin = perlin

	hmap.MountainNoise = noiselib.DefaultRidgedmulti()
	hmap.MountainNoise.Frequency = 0.05
	hmap.MountainNoise.OctaveCount = 14

	hmap.MountainScaleBias = noiselib.DefaultScaleBias()
	hmap.MountainScaleBias.SetSourceModule(0, hmap.MountainNoise)
	hmap.MountainScaleBias.Scale = 2.3
	hmap.MountainScaleBias.Bias = 0.0

	hmap.PlainNoise = noiselib.DefaultBillow()
	hmap.PlainNoise.Frequency = 0.01

	hmap.PlainScaleBias = noiselib.DefaultScaleBias()
	hmap.PlainScaleBias.SetSourceModule(0, hmap.PlainNoise)
	hmap.PlainScaleBias.Scale = 0.125
	hmap.PlainScaleBias.Bias = 0.5

	hmap.TerrainType = noiselib.DefaultPerlin()
	hmap.TerrainType.Frequency = 0.05
	hmap.TerrainType.Persistence = 0.25

	hmap.FinalTerrain = noiselib.DefaultSelect()
	hmap.FinalTerrain.SetSourceModule(0, hmap.MountainScaleBias)
	hmap.FinalTerrain.SetSourceModule(1, hmap.PlainScaleBias)
	hmap.FinalTerrain.SetSourceModule(2, hmap.TerrainType)
	hmap.FinalTerrain.LowerBound = 0
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

	programTree, err := gfx.NewProgramFromVertFrag("tree")
	if err != nil {
		return err
	}
	defer programTree.Delete()

	programChunk, err := gfx.NewProgramFromVertFrag("chunk")
	if err != nil {
		return err
	}
	defer programChunk.Delete()

	// textureBranches, err := gfx.NewTextureFromFile("data/textures/tree/branches.png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// textureLeaves, err := gfx.NewTextureFromFile("data/textures/tree/leaves.png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)

	camera := cam.NewFpsCamera(mgl32.Vec3{0, -5, 0}, mgl32.Vec3{0, 1, 0}, 45, 45, window.InputManager())

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
		go ter.ChunkLoadingWorker(loadQueue, &hmap)
	}

	loadListChangeFlag := true

	// var instanceTreesHQ []*veg.InstanceTree
	// var instanceTreesLQ []*veg.InstanceTree

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
				// if currentChunk == chunk.Position {
				// 	instanceTreesHQ = veg.GetSurroundingForests(instanceTreesHQ, chunk, true)
				// } else {
				// 	instanceTreesLQ = veg.GetSurroundingForests(instanceTreesLQ, chunk, false)
				// }
			}
		}

		if currentChunk != getCurrentChunkFromCam(*camera, &hmap) {
			currentChunk = getCurrentChunkFromCam(*camera, &hmap)
			loadListChangeFlag = true

			// instanceTreesHQ = []*veg.InstanceTree{}
			// instanceTreesLQ = []*veg.InstanceTree{}
			// for _, chunk := range renderList {
			// 	if currentChunk == chunk.Position {
			// 		instanceTreesHQ = veg.GetSurroundingForests(instanceTreesHQ, chunk, true)
			// 	} else {
			// 		instanceTreesLQ = veg.GetSurroundingForests(instanceTreesLQ, chunk, false)
			// 	}
			// }
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

		window.StartFrame()
		camera.Update(window.SinceLastFrame())
		gl.ClearColor(135.0/255.0, 206.0/255.0, 250.0/255.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // depth buffer needed for DEPTH_TEST

		scr.RenderChunks(renderList, camera, programChunk)

		// textureBranches.Bind(gl.TEXTURE1)
		// textureLeaves.Bind(gl.TEXTURE2)
		// for _, instanceTreeHQ := range instanceTreesHQ {
		// 	scr.RenderForest(instanceTreeHQ.Parent, camera, programTree, len(instanceTreeHQ.Transforms))
		// }
		// for _, instanceTreeLQ := range instanceTreesLQ {
		// 	scr.RenderForest(instanceTreeLQ.Parent, camera, programTree, len(instanceTreeLQ.Transforms))
		// }
		// textureLeaves.UnBind()
		// textureBranches.UnBind()
	}

	return nil
}
