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

// rule: "F[+F]F[-F]F",
type Tree struct {
	grammar  string
	angle    float32
	rule     string
	axiom    string
	nbrIndex int
	Branches []Branch
}

type Branch struct {
	angleY   float32
	angleZ   float32
	height   float32
	radius   float32
	position mgl32.Vec3
	Model    *gfx.Model
}

func CreateTree() *Tree {
	tree := &Tree{
		rule:     "F[-F]F[+F][F]",
		angle:    30.0,
		grammar:  "F",
		axiom:    "F",
		nbrIndex: 3,
	}
	tree.createGrammar()
	tree.createBranches()

	fmt.Println("nbrOfTriangles: ", int32(len(tree.Branches))*tree.Branches[0].Model.NbTriangles)
	return tree
}

func createModel(b Branch) *gfx.Model {
	mesh := gfx.Mesh{}
	nRadius := 3
	dr := 2.0 * math.Pi / float64(nRadius)
	var index uint32

	start := b.position
	toAdd := rotateZ(b.angleZ, mgl32.Vec3{0, b.height, 0})
	toAdd = rotateY(b.angleY, toAdd)
	end := start.Add(toAdd)

	radiusDec := b.radius / 100.0

	for i := 0; i < nRadius; i++ {
		p1 := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}.Mul(float32(b.radius)).Add(start)
		p2 := mgl32.Vec3{float32(math.Cos(dr * float64(i+1))), 0, float32(math.Sin(dr * float64(i+1)))}.Mul(float32(b.radius)).Add(start)
		p3 := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}.Mul(float32(b.radius - radiusDec)).Add(end)
		p4 := mgl32.Vec3{float32(math.Cos(dr * float64(i+1))), 0, float32(math.Sin(dr * float64(i+1)))}.Mul(float32(b.radius - radiusDec)).Add(end)
		normal := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}
		normal = normal.Normalize()
		color := mgl32.Vec4{0.5, 0.2, 0.1, 1.0}
		texture := mgl32.Vec2{0.0, 0.0}
		mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p1, Normal: normal, Color: color, Texture: texture})
		mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p2, Normal: normal, Color: color, Texture: texture})
		mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p3, Normal: normal, Color: color, Texture: texture})
		mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p4, Normal: normal, Color: color, Texture: texture})
		t1 := gfx.TriangleConnectivity{U0: index, U1: index + 1, U2: index + 3}
		t2 := gfx.TriangleConnectivity{U0: index, U1: index + 3, U2: index + 2}
		mesh.Connectivity = append(mesh.Connectivity, t1)
		mesh.Connectivity = append(mesh.Connectivity, t2)
		index += 4
	}

	model := gfx.BuildModel(mesh)
	translate := mgl32.Translate3D(0, 0, 0)
	model.Transform = translate
	return &model
}

func (t *Tree) createGrammar() {
	for index := 0; index < t.nbrIndex; index++ {
		t.grammar = strings.Replace(t.grammar, t.axiom, t.rule, -1)
	}
}

func (t *Tree) createBranches() {
	var branches []Branch
	var branch = Branch{radius: 0.10, height: -1}
	var addSomething = false

	for _, letter := range strings.Split(t.grammar, "") {
		switch letter {
		case "F":
			if addSomething {
				toAdd := rotateZ(branch.angleZ, mgl32.Vec3{0, branch.height, 0})
				toAdd = rotateY(branch.angleY, toAdd)
				branch.position = branch.position.Add(toAdd)
			}
			branch.radius -= branch.radius / 100.0
			addSomething = true
			branch.Model = createModel(branch)
			t.Branches = append(t.Branches, branch)
			fmt.Println(branch.radius)
			break
		case "+":
			branch.angleZ += t.angle
			addSomething = false
			break
		case "-":
			branch.angleZ -= t.angle
			addSomething = false
			break
		case "[":
			if len(branches) == 0 {
				branch.angleY += float32(rand.Float64() * 360.0)
			}
			branches = append(branches, branch) //push
			break
		case "]":
			branch, branches = branches[len(branches)-1], branches[:len(branches)-1] //pop
			break
		}
	}
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
