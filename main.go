package main

import (
	//"fmt"
	"log"
	"runtime"
	"time"

	"./cam"
	"./gfx"
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
var LOAD_DISTANCE = 6
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

	window := win.NewWindow(1920, 1080, "ProceduralGo - Arthur BARRIERE - Adrien BOUCAUD", false)

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var perlin = noiselib.DefaultPerlin()
	perlin.Seed = 0*int(time.Now().Unix())
	perlin.OctaveCount = 14
	perlin.Frequency = 0.1
	perlin.Lacunarity  = 2.2
	perlin.Persistence = 0.5
	perlin.Quality = noiselib.QualitySTD


	hmap = ter.HeightMap{
		ChunkNBPoints: 512,
		ChunkWorldSize: 12,
		NbOctaves:4,
		Exponent:1.0,
	}

	hmap.Perlin = perlin

	hmap.MountainNoise = noiselib.DefaultRidgedmulti()
	hmap.MountainNoise.Frequency = 0.05
	hmap.MountainNoise.OctaveCount = 14

	hmap.MountainScaleBias = noiselib.DefaultScaleBias()
	hmap.MountainScaleBias.SetSourceModule(0, hmap.MountainNoise)
	hmap.MountainScaleBias.Scale = 1.8
	hmap.MountainScaleBias.Bias = 0.0

	hmap.PlainNoise = noiselib.DefaultBillow()
	hmap.PlainNoise.Frequency = 0.02

	hmap.PlainScaleBias = noiselib.DefaultScaleBias()
	hmap.PlainScaleBias.SetSourceModule(0, hmap.PlainNoise)
	hmap.PlainScaleBias.Scale = 0.125
	hmap.PlainScaleBias.Bias = 0.5

	hmap.TerrainType = noiselib.DefaultPerlin()
	hmap.TerrainType.Frequency = 0.1
	hmap.TerrainType.Persistence = 0.25

	hmap.FinalTerrain = noiselib.DefaultSelect()
	hmap.FinalTerrain.SetSourceModule(0, hmap.MountainScaleBias)
	hmap.FinalTerrain.SetSourceModule(1, hmap.PlainScaleBias)
	hmap.FinalTerrain.SetSourceModule(2, hmap.TerrainType)
	hmap.FinalTerrain.LowerBound = 0
	hmap.FinalTerrain.UpperBound = 1000
	hmap.FinalTerrain.SetEdgeFalloff(0.125)

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
	// the linked shader program determines how the data will be rendered
	vertShader, err := gfx.NewShaderFromFile("shaders/shader.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShaderFromFile("shaders/shader.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}

	chunkVertShader, err := gfx.NewShaderFromFile("shaders/chunk.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	chunkFragShader, err := gfx.NewShaderFromFile("shaders/chunk.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	chunkProgram, err := gfx.NewProgram(chunkVertShader, chunkFragShader)
	if err != nil {
		return err
	}

	defer program.Delete()

	textureBranches, err := gfx.NewTextureFromFile("data/branches.png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	textureLeaves, err := gfx.NewTextureFromFile("data/leaves.png", gl.CLAMP_TO_EDGE, gl.CLAMP_TO_EDGE)
	if err != nil {
		panic(err.Error())
	}

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)

	camera := cam.NewFpsCamera(mgl32.Vec3{0, -5, 0}, mgl32.Vec3{0, 1, 0}, 45, 45, window.InputManager())

	currentChunk := getCurrentChunkFromCam(*camera, &hmap)

	for indexRule := range veg.Rules() {
		trees = append(trees, veg.CreateTree(indexRule, mgl32.Vec3{5, -2, float32(indexRule) * 4.0}))
	}

	//init lists
	var visibilityList []*ter.Chunk
	var renderList[]*ter.Chunk
	var loadList[]*ter.Chunk
	visibilityList = ter.GetVisibilityList(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, VIEW_DISTANCE)

	//create job queue
	loadQueue := make(chan *ter.Chunk, 1000)

	//start workers
	for i := 0; i < NUM_WORKERS; i++{
		go ter.ChunkLoadingWorker(loadQueue, &hmap)
	}

	loadListChangeFlag := true

	for !window.ShouldClose() {
		//OpenGL loading for new chunks

		for _,chunk := range loadList{
			if chunk.AtomicNeedOpenGLLoading == 1 && chunk.Loaded == false{
				gfx.LoadModelData(chunk.Model) //
				translate := mgl32.Translate3D(float32(chunk.Position[0])*float32(chunk.WorldSize), 0, float32(chunk.Position[1])*float32(chunk.WorldSize))
				chunk.Model.Transform = translate
				chunk.Model.Program = chunkProgram
				chunk.Loaded = true //should not need to change other flags if this one is set
				loadListChangeFlag = true
			}
		}

		if currentChunk != getCurrentChunkFromCam(*camera, &hmap) {
			currentChunk = getCurrentChunkFromCam(*camera, &hmap)
			loadListChangeFlag = true
		}

		if loadListChangeFlag {
			loadList = ter.GetLoadList(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, LOAD_DISTANCE)
			visibilityList = ter.GetVisibilityList(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, VIEW_DISTANCE)
			//submit loading jobs
			for _, chunk := range loadList{
				if !chunk.Loaded && !chunk.Loading{
					chunk.Loading = true
					loadQueue <- chunk
				}
			}
			loadListChangeFlag = false
		}

		renderList = ter.GetRenderList(&hmap, visibilityList, *camera)

		// swaps in last buffer, polls for window events, and generally sets up for a new render frame
		window.StartFrame()

		// update camera position and direction from input evevnts
		camera.Update(window.SinceLastFrame())

		// background color
		gl.ClearColor(135.0/255.0, 206.0/255.0, 250.0/255.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // depth buffer needed for DEPTH_TEST

		// creates perspective
		fov := float32(90.0)
		near := float32(0.001)
		far := float32(100.0)
		projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(window.Width())/float32(window.Height()), near, far)
		camTransform := camera.GetTransform()

		program.Use()

		gl.UniformMatrix4fv(program.GetUniformLocation("view"), 1, false, &camTransform[0])
		gl.UniformMatrix4fv(program.GetUniformLocation("project"), 1, false, &projectTransform[0])

		// obj is colored, light is white
		gl.Uniform3f(program.GetUniformLocation("objectColor"), 0.0, 0.5, 0.0)
		gl.Uniform3f(program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
		gl.Uniform3f(program.GetUniformLocation("lightPos"), camera.Position().X(), camera.Position().Y(), camera.Position().Z())

		for _, chunk := range renderList{
			gfx.Render(*(chunk.Model), camTransform, projectTransform, camera.Position())
		}

		textureBranches.Bind(gl.TEXTURE1)
		textureLeaves.Bind(gl.TEXTURE2)

		for _, tree := range trees {

			textureBranches.SetUniform(program.GetUniformLocation("currentTexture"))
			tree.BranchesModel.Program = program
			tree.BranchesModel.TextureID = gl.TEXTURE1
			gfx.Render(*(tree.BranchesModel), camTransform, projectTransform, camera.Position())

			textureLeaves.SetUniform(program.GetUniformLocation("currentTexture"))
			tree.LeavesModel.Program = program
			tree.LeavesModel.TextureID = gl.TEXTURE2
			gfx.Render(*(tree.LeavesModel), camTransform, projectTransform, camera.Position())
		}

		textureLeaves.UnBind()
		textureBranches.UnBind()

		gl.BindVertexArray(0)

		// end of draw loop
	}

	return nil
}
