package veg

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"../gfx"

	"github.com/go-gl/mathgl/mgl32"
)

var cubePositions [][3]float32

type Tree struct {
	grammar string
	angle   float32
	rule    string
	axiom   string
}

func CreateTree() *Tree {
	tree := generateTree()
	renderTree(tree)

	return tree
}

// ChunkNBPoints:  128,
// ChunkWorldSize: 32,

func CylinderModel() *gfx.Model {

	nRadius := 5
	nHeight := 5

	radius := 0.10
	height := 1.0

	dr := 2.0 * math.Pi / float64(nRadius)
	dh := float32(height) / float32(nHeight)

	mesh := gfx.Mesh{}

	var index uint32

	for i := 0; i < nRadius; i++ {
		p1 := mgl32.Vec3{
			float32(math.Cos(dr * float64(i))),
			0,
			float32(math.Sin(dr * float64(i))),
		}.Mul(float32(radius))
		p2 := mgl32.Vec3{
			float32(math.Cos(dr * float64(i+1))),
			0,
			float32(math.Sin(dr * float64(i+1))),
		}.Mul(float32(radius))
		p3 := p1
		p4 := p2

		for j := 0; j < nHeight; j++ {
			p1[1] = dh * float32(j)
			p2[1] = dh * float32(j)
			p3[1] = dh * float32(j+1)
			p4[1] = dh * float32(j+1)
			// fmt.Println(p1, p2, p3, p4)
			normal := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}
			normal = normal.Normalize()
			color := mgl32.Vec4{0.0, 0.5, 0.0, 1.0}
			texture := mgl32.Vec2{0.0, 0.0}
			v1 := gfx.Vertex{Position: p1, Normal: normal, Color: color, Texture: texture}
			v2 := gfx.Vertex{Position: p2, Normal: normal, Color: color, Texture: texture}
			v3 := gfx.Vertex{Position: p3, Normal: normal, Color: color, Texture: texture}
			v4 := gfx.Vertex{Position: p4, Normal: normal, Color: color, Texture: texture}
			mesh.Vertices = append(mesh.Vertices, v1)
			mesh.Vertices = append(mesh.Vertices, v2)
			mesh.Vertices = append(mesh.Vertices, v3)
			mesh.Vertices = append(mesh.Vertices, v4)
			tri1 := gfx.TriangleConnectivity{index, index + 1, index + 3}
			tri2 := gfx.TriangleConnectivity{index, index + 3, index + 2}
			mesh.Connectivity = append(mesh.Connectivity, tri1)
			mesh.Connectivity = append(mesh.Connectivity, tri2)
			index += 4
		}

	}

	fmt.Println(len(mesh.Vertices))

	model := gfx.BuildModel(mesh)
	translate := mgl32.Translate3D(0, 0, 0)
	model.Transform = translate
	return &model
}

func generateTree() *Tree {
	// t.angle = rand.Float32() * 90.0

	t := &Tree{
		rule: "F[-F]F[+F][F]",
		// rule: "F[+F]F[-F]F",

		angle:   30.0,
		grammar: "F",
		axiom:   "F",
	}

	for index := 0; index < 4; index++ {
		t.grammar = strings.Replace(t.grammar, t.axiom, t.rule, -1)
		fmt.Println(t.grammar)
	}
	return t
}

type BranchNode struct {
	angleZ   float32
	angleX   float32
	height   float32
	position mgl32.Vec3
}

func renderTree(t *Tree) {
	var branchNodes []BranchNode
	var branchNode BranchNode
	var toAdd mgl32.Vec3
	for _, letter := range strings.Split(t.grammar, "") {
		switch letter {
		case "F":
			toAdd = rotateZ(branchNode.angleZ, mgl32.Vec3{0, -1, 0})
			toAdd = rotateX(branchNode.angleX, toAdd)
			branchNode.position = branchNode.position.Add(toAdd)
			cubePositions = append(cubePositions, branchNode.position)
			break
		case "+":
			branchNode.angleZ += t.angle
			branchNode.angleX += rand.Float32() * 360.0
			break
		case "-":
			branchNode.angleZ -= t.angle
			branchNode.angleX -= rand.Float32() * 360.0
			break
		case "[":
			branchNodes = append(branchNodes, branchNode) //push
			break
		case "]":
			branchNode, branchNodes = branchNodes[len(branchNodes)-1], branchNodes[:len(branchNodes)-1] //pop
			break
		}
		// fmt.Println(position)
	}

}

func renderTrunk(radius float64, height float64) {
	stepThetha := 360.0 / 100.0
	stepHeight := height / 100.0

	for h := 0.0; h < height; h += stepHeight {
		for theta := 0.0; theta <= 360.0; theta += stepThetha {
			x := radius * math.Cos(theta)
			y := h
			z := radius * math.Sin(theta)
			pos := mgl32.Vec3{float32(x), float32(y), float32(z)}
			cubePositions = append(cubePositions, [3]float32{float32(pos[0]), float32(pos[1]), float32(pos[2])})
		}
	}

}

func CubePositions() [][3]float32 {
	return cubePositions
}

func rotateX(angleDegree float32, original mgl32.Vec3) mgl32.Vec3 {
	angle := (float32(math.Pi) * angleDegree) / 180.0
	return mgl32.Rotate3DX(angle).Mul3x1(original)
}

func rotateY(angleDegree float32, original mgl32.Vec3) mgl32.Vec3 {
	angle := (float32(math.Pi) * angleDegree) / 180.0
	return mgl32.Rotate3DY(angle).Mul3x1(original)
}

func rotateZ(angleDegree float32, original mgl32.Vec3) mgl32.Vec3 {
	angle := (float32(math.Pi) * angleDegree) / 180.0
	return mgl32.Rotate3DZ(angle).Mul3x1(original)
}

func translateY(height float32, original mgl32.Vec3) mgl32.Vec3 {
	matrix := mgl32.Translate3D(0, height, 0)
	vector4 := mgl32.Vec4{original.X(), original.Y(), original.Z(), 1.0}
	return matrix.Mul4x1(vector4).Vec3()
}
