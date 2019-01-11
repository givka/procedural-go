package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"./cam"
	"./gfx"
	"./ter"
	"./win"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

var mesh gfx.Mesh
var model gfx.Model
var hmap ter.HeightMap

var chunks []*ter.HeightMapChunk

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

	window := win.NewWindow(1280, 720, "ProceduralGo - Arthur BARRIERE - Adrien BOUCAUD", true)

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var perlin = noiselib.DefaultPerlin()
	perlin.Seed = int(time.Now().Unix())

	hmap = ter.HeightMap{ChunkNBPoints: 16, ChunkWorldSize: 10, NbOctaves: 4}
	hmap.Perlin = perlin

	chunks = ter.GetSurroundingChunks(&hmap, mgl32.Vec2{0, 0}, 8)

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
	defer program.Delete()

	for _, chunk := range chunks {
		chunk.Model.Program = program
	}

	// ensure that triangles that are "behind" others do not draw over top of them
	gl.Enable(gl.DEPTH_TEST)

	camera := cam.NewFpsCamera(mgl32.Vec3{0, -5, 0}, mgl32.Vec3{0, 1, 0}, 45, 45, window.InputManager())

	currentChunk := getCurrentChunkFromCam(*camera, &hmap)
	for !window.ShouldClose() {
		if currentChunk != getCurrentChunkFromCam(*camera, &hmap) {
			currentChunk = getCurrentChunkFromCam(*camera, &hmap)
			fmt.Println("New Chunk", currentChunk)
			chunks = ter.GetSurroundingChunks(&hmap, mgl32.Vec2{camera.Position().X(), camera.Position().Z()}, 8)
			for _, chunk := range chunks {
				chunk.Model.Program = program
			}
		}
		// swaps in last buffer, polls for window events, and generally sets up for a new render frame
		window.StartFrame()

		// update camera position and direction from input evevnts
		camera.Update(window.SinceLastFrame())

		// background color
		gl.ClearColor(135.0/255.0, 206.0/255.0, 250.0/255.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT) // depth buffer needed for DEPTH_TEST

		// creates perspective
		fov := float32(90.0)
		far := float32(100.0)
		projectTransform := mgl32.Perspective(mgl32.DegToRad(fov), float32(window.Width())/float32(window.Height()), 0.1, far)

		camTransform := camera.GetTransform()
		/*		lightTransform := mgl32.Translate3D(lightPos.X(), lightPos.Y(), lightPos.Z()).Mul4(
				mgl32.Scale3D(5, 5, 5))
		*/

		program.Use()

		//DEBUT RENDER
		gl.UniformMatrix4fv(program.GetUniformLocation("view"), 1, false, &camTransform[0])
		gl.UniformMatrix4fv(program.GetUniformLocation("project"), 1, false, &projectTransform[0])

		// obj is colored, light is white
		gl.Uniform3f(program.GetUniformLocation("objectColor"), 0.0, 0.5, 0.0)
		gl.Uniform3f(program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
		gl.Uniform3f(program.GetUniformLocation("lightPos"), camera.Position().X(), camera.Position().Y(), camera.Position().Z())

		//	gfx.Render(model, camTransform, projectTransform)
		for _, chunk := range chunks {
			gfx.Render(*(chunk.Model), camTransform, projectTransform, camera.Position())
		}
		gl.BindVertexArray(0)

		// end of draw loop
	}

	return nil
}
