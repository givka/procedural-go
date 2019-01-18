package ctx

var (
	width  = 1280
	height = 720
)

const (
	Fov  float32 = 90.0
	Far  float32 = 100.0
	Near float32 = 0.001
)

func SetWidth(w int) {
	width = w
}

func SetHeight(h int) {
	height = h
}

func Width() int {
	return width
}

func Height() int {
	return height
}
