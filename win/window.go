package win

import (
	"log"
	"strconv"

	"../ctx"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

// Window class for the glfw window
type Window struct {
	glfw  *glfw.Window
	title string
	vsync bool

	inputManager  *InputManager
	firstFrame    bool
	dTime         float64
	lastFrameTime float64
}

// InputManager returns the current InputManager
func (w *Window) InputManager() *InputManager {
	return w.inputManager
}

func resizeCallback(w *glfw.Window, width int, height int) {
	// hack for HiDPI displays
	ww, hh := w.GetFramebufferSize()
	ctx.SetWidth(ww)
	ctx.SetHeight(hh)
	gl.Viewport(0, 0, int32(ww), int32(hh))
}

// NewWindow returns a new window initialized
func NewWindow(width int, height int, title string, vsync bool) *Window {
	gWindow, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	gWindow.MakeContextCurrent()

	if vsync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}

	im := NewInputManager()

	// uncomment this to disable cursor
	// gWindow.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	gWindow.SetCursorPosCallback(im.mouseCallback)
	gWindow.SetMouseButtonCallback(im.mouseButtonCallback)
	gWindow.SetKeyCallback(im.keyCallback)
	gWindow.SetSizeCallback(resizeCallback)

	return &Window{
		title:        title,
		glfw:         gWindow,
		inputManager: im,
		firstFrame:   true,
		vsync:        vsync,
	}

}

// Width returns window width
func (w *Window) Width() int {
	width, _ := w.glfw.GetFramebufferSize()
	return width
}

// Height returns window height
func (w *Window) Height() int {
	_, height := w.glfw.GetFramebufferSize()
	return height
}

// ShouldClose returns if the window should close
func (w *Window) ShouldClose() bool {
	return w.glfw.ShouldClose()
}

// StartFrame sets everything up to start rendering a new frame.
// This includes swapping in last rendered buffer, polling for window events,
// checkpointing cursor tracking, and updating the time since last frame.
func (w *Window) StartFrame() {
	// swap in the previous rendered buffer
	w.glfw.SwapBuffers()

	// poll for UI window events
	glfw.PollEvents()

	if w.inputManager.IsKeyActive(ProgramQuit) {
		w.glfw.SetShouldClose(true)
	}

	// base calculations of time since last frame (basic program loop idea)
	// For better advanced impl, read: http://gafferongames.com/game-physics/fix-your-timestep/
	curFrameTime := glfw.GetTime()

	if w.firstFrame {
		w.lastFrameTime = curFrameTime
		w.firstFrame = false
	}

	// display screen info every 500ms (1.0/0.500 = 2)
	if int(curFrameTime*2) != int(w.lastFrameTime*2) {
		fps := int(1.0 / w.dTime)
		w.glfw.SetTitle(w.title + " - [VSYNC: " + strconv.FormatBool(w.vsync) + "] - [FPS: " + strconv.Itoa(fps) + "]")
	}

	w.dTime = curFrameTime - w.lastFrameTime
	w.lastFrameTime = curFrameTime

	w.inputManager.CheckpointCursorChange()

}

// SinceLastFrame returns the time elapsed since last frame
func (w *Window) SinceLastFrame() float64 {
	return w.dTime
}
