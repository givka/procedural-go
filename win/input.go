package win

import (
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl64"
)

// ActionKey is a configurable abstraction of a key press
type ActionKey int

// ActionKey enum
const (
	PlayerForward  ActionKey = iota
	PlayerBackward ActionKey = iota
	PlayerLeft     ActionKey = iota
	PlayerRight    ActionKey = iota
	ProgramQuit    ActionKey = iota
)

// ActionButton is a configurable abstraction of a mouse button press
type ActionButton int

// ActionButton enum
const (
	MouseLeft   ActionButton = iota
	MouseRight  ActionButton = iota
	MouseMiddle ActionButton = iota
)

// InputManager class to get keyboard and mouseButton actions
type InputManager struct {
	actionToKeyMap    map[ActionKey]glfw.Key
	actionToButtonMap map[ActionButton]glfw.MouseButton

	keysPressed    [glfw.KeyLast]bool
	buttonsPressed [glfw.MouseButtonLast]bool

	firstCursorAction    bool
	cursor               mgl64.Vec2
	cursorChange         mgl64.Vec2
	cursorLast           mgl64.Vec2
	bufferedCursorChange mgl64.Vec2
}

// NewInputManager returns an initialized InputManager
func NewInputManager() *InputManager {
	actionToKeyMap := map[ActionKey]glfw.Key{
		PlayerForward:  glfw.KeyW,
		PlayerBackward: glfw.KeyS,
		PlayerLeft:     glfw.KeyA,
		PlayerRight:    glfw.KeyD,
		ProgramQuit:    glfw.KeyEscape,
	}

	actionToButtonMap := map[ActionButton]glfw.MouseButton{
		MouseLeft:   glfw.MouseButton1,
		MouseRight:  glfw.MouseButton2,
		MouseMiddle: glfw.MouseButton3,
	}

	return &InputManager{
		actionToKeyMap:    actionToKeyMap,
		actionToButtonMap: actionToButtonMap,
		firstCursorAction: true,
	}
}

// IsKeyActive returns whether the given Action is currently active
func (im *InputManager) IsKeyActive(a ActionKey) bool {
	return im.keysPressed[im.actionToKeyMap[a]]
}

// IsButtonActive returns whether the given ActionButton is currently active
func (im *InputManager) IsButtonActive(a ActionButton) bool {
	return im.buttonsPressed[im.actionToButtonMap[a]]
}

// Cursor returns the value of the cursor at the last time that CheckpointCursorChange() was called.
func (im *InputManager) Cursor() mgl64.Vec2 {
	return im.cursor
}

// CursorChange returns the amount of change in the underlying cursor
// since the last time CheckpointCursorChange was called
func (im *InputManager) CursorChange() mgl64.Vec2 {
	return im.cursorChange
}

// CheckpointCursorChange updates the publicly available Cursor() and CursorChange()
// methods to return the current Cursor and change since last time this method was called.
func (im *InputManager) CheckpointCursorChange() {
	im.cursorChange[0] = im.bufferedCursorChange[0]
	im.cursorChange[1] = im.bufferedCursorChange[1]
	im.cursor[0] = im.cursorLast[0]
	im.cursor[1] = im.cursorLast[1]

	im.bufferedCursorChange[0] = 0
	im.bufferedCursorChange[1] = 0
}

func (im *InputManager) keyCallback(window *glfw.Window, key glfw.Key, scancode int,
	action glfw.Action, mods glfw.ModifierKey) {

	// timing for key events occurs differently from what the program loop requires
	// so just track what key actions occur and then access them in the program loop
	switch action {
	case glfw.Press:
		im.keysPressed[key] = true
	case glfw.Release:
		im.keysPressed[key] = false
	}

}

func (im *InputManager) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		im.buttonsPressed[button] = true
	case glfw.Release:
		im.buttonsPressed[button] = false
	}

	if im.actionToButtonMap[MouseLeft] == button {
		if im.buttonsPressed[button] {
			im.firstCursorAction = true
			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		} else {
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		}
	}
}

func (im *InputManager) mouseCallback(window *glfw.Window, xpos, ypos float64) {

	if !im.IsButtonActive(MouseLeft) {
		return
	}

	if im.firstCursorAction {
		im.cursorLast[0] = xpos
		im.cursorLast[1] = ypos
		im.firstCursorAction = false
	}

	im.bufferedCursorChange[0] += xpos - im.cursorLast[0]
	im.bufferedCursorChange[1] += ypos - im.cursorLast[1]

	im.cursorLast[0] = xpos
	im.cursorLast[1] = ypos
}
