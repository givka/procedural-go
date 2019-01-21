package veg

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"../gfx"
	"../ter"

	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	uniqueTreesHQ []*Tree
	uniqueTreesLQ []*Tree
)

type Tree struct {
	grammar       string
	angle         float32
	rule          string
	axiom         string
	position      mgl32.Vec3
	BranchesModel *gfx.Model
	LeavesModel   *gfx.Model
}

type Branch struct {
	angleY   float32
	angleZ   float32
	height   float32
	radius   float32
	position mgl32.Vec3
}

type InstanceTree struct {
	Parent     *Tree
	Transforms []mgl32.Mat4
}

func initialiseParentTrees() {
	rules := []string{
		"F[+FF]F[-F]",
		"F[-F]F[+F][F]",
		"F[+F][-FF]F",
		"F[+F]F[-F]F",
		"F[-F+F]F[+F]",
	}

	for _, rule := range rules {
		uniqueTreesHQ = append(uniqueTreesHQ, createTreeHQ(rule))
	}

	// for now, only one model of LQ tree
	uniqueTreesLQ = append(uniqueTreesLQ, createTreeLQ())
}

func createTreeHQ(rule string) *Tree {
	tree := &Tree{
		rule:    rule,
		angle:   15.0 + rand.Float32()*15.0,
		grammar: "F",
		axiom:   "F",
	}

	for index := 0; index < 2; index++ {
		tree.grammar = strings.Replace(tree.grammar, tree.axiom, tree.rule, -1)
	}

	tree.generateFromGrammar()

	fmt.Println(tree.rule, "Tris count:", tree.BranchesModel.NbTriangles+tree.LeavesModel.NbTriangles)

	return tree
}

func createTreeLQ() *Tree {
	tree := &Tree{}
	branches := []Branch{Branch{radius: 0.001, height: -0.05}}
	tree.BranchesModel = createBranchesModel(branches)
	branches = []Branch{Branch{radius: 0.001, height: -0.05, position: mgl32.Vec3{0, -0.05, 0}}}
	tree.LeavesModel = createLeavesModel(branches, 0.05)
	return tree
}

func GetSurroundingForests(instanceTrees []*InstanceTree, chunk *ter.Chunk, isHQ bool) []*InstanceTree {
	if len(instanceTrees) == 0 {
		if len(uniqueTreesHQ) == 0 {
			initialiseParentTrees()
		}
		if isHQ {
			for _, uniqueTreeHQ := range uniqueTreesHQ {
				instanceTrees = append(instanceTrees, &InstanceTree{Parent: uniqueTreeHQ})
			}
		} else {
			for _, uniqueTreeLQ := range uniqueTreesLQ {
				instanceTrees = append(instanceTrees, &InstanceTree{Parent: uniqueTreeLQ})
			}
		}
	}

	CreateForest(chunk, isHQ, instanceTrees)

	nbrTrees := 0

	for _, instanceTree := range instanceTrees {
		if len(instanceTree.Transforms) > 0 {
			gfx.ModelToInstanceModel(instanceTree.Parent.BranchesModel, instanceTree.Transforms)
			gfx.ModelToInstanceModel(instanceTree.Parent.LeavesModel, instanceTree.Transforms)
			nbrTrees += len(instanceTree.Transforms)
		}
	}

	return instanceTrees
}

func CreateForest(chunk *ter.Chunk, isHQ bool, instanceTrees []*InstanceTree) {
	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)
	angle := float32(5.0 * math.Cos(glfw.GetTime()))
	for x := 0; x < int(chunk.NBPoints)+1; x+= 10 {
		for z := 0; z < int(chunk.NBPoints)+1; z+= 10 {
			i := x + z*int(chunk.NBPoints+1)
			posY := float32(chunk.Map[i])
			if posY < 0.0 || posY > 0.10 {
				continue
			}
			posX := float32(chunk.Position[0])*float32(chunk.WorldSize) + float32(x)*step
			posZ := float32(chunk.Position[1])*float32(chunk.WorldSize) + float32(z)*step
			transform := mgl32.Translate3D(posX, -2*posY, posZ).Mul4(mgl32.Rotate3DY(posY * 360.0).Mat4())
			transform = transform.Mul4(mgl32.Rotate3DX(mgl32.DegToRad(angle)).Mat4())
			if !isHQ {
				transform = transform.Mul4(mgl32.Scale3D(5, 5, 5))
			}
			index := rand.Intn(len(instanceTrees))
			instanceTrees[index].Transforms = append(instanceTrees[index].Transforms, transform)
		}
	}
}

func createLeavesModel(branches []Branch, customSizeLeaves ...float32) *gfx.Model {
	mesh := gfx.Mesh{}
	nbRadius := 2
	index := uint32(0)

	for _, branch := range branches {
		dr := 180.0 / (nbRadius)
		offsetRotY := rand.Float32() * 360.0
		sizeLeaves := (branch.position.Y()) / 2.0

		if len(customSizeLeaves) > 0 {
			sizeLeaves = customSizeLeaves[0]
		}

		for i := 0; i < nbRadius; i++ {
			toAdd := rotateZ(branch.angleZ, mgl32.Vec3{0, branch.height / 2, 0})
			toAdd = rotateY(branch.angleY, toAdd)
			start := branch.position.Add(toAdd)
			end := start.Add(rotateY(branch.angleY, rotateZ(branch.angleZ, mgl32.Vec3{0, sizeLeaves, 0})))

			toAdd2 := rotateY(offsetRotY+float32(i*dr), mgl32.Vec3{sizeLeaves, 0, sizeLeaves})
			p1 := start.Sub(toAdd2)
			p2 := start.Add(toAdd2)
			p3 := end.Sub(toAdd2)
			p4 := end.Add(toAdd2)

			normal := mgl32.Vec3{0, -1, 0}
			normal = normal.Normalize()
			color := mgl32.Vec4{0.0, 0.5, 0.0, 1.0}
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p1, Normal: normal, Color: color, Texture: mgl32.Vec2{0.0, 0.0}})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p2, Normal: normal, Color: color, Texture: mgl32.Vec2{1.0, 0.0}})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p3, Normal: normal, Color: color, Texture: mgl32.Vec2{0.0, 1.0}})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p4, Normal: normal, Color: color, Texture: mgl32.Vec2{1.0, 1.0}})
			t1 := gfx.TriangleConnectivity{U0: index, U1: index + 1, U2: index + 3}
			t2 := gfx.TriangleConnectivity{U0: index, U1: index + 3, U2: index + 2}
			mesh.Connectivity = append(mesh.Connectivity, t1)
			mesh.Connectivity = append(mesh.Connectivity, t2)
			index += 4
		}
	}

	model := gfx.BuildModel(mesh)
	return &model
}

func createBranchesModel(branches []Branch) *gfx.Model {
	mesh := gfx.Mesh{}
	nbRadius := 3
	dr := 2.0 * math.Pi / float64(nbRadius)
	index := uint32(0)

	for _, branch := range branches {
		start := branch.position
		toAdd := rotateZ(branch.angleZ, mgl32.Vec3{0, branch.height, 0})
		toAdd = rotateY(branch.angleY, toAdd)
		end := start.Add(toAdd)
		radiusDec := branch.radius / 5.0
		for i := 0; i < nbRadius; i++ {
			p1 := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}.Mul(float32(branch.radius)).Add(start)
			p2 := mgl32.Vec3{float32(math.Cos(dr * float64(i+1))), 0, float32(math.Sin(dr * float64(i+1)))}.Mul(float32(branch.radius)).Add(start)
			p3 := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}.Mul(float32(branch.radius - radiusDec)).Add(end)
			p4 := mgl32.Vec3{float32(math.Cos(dr * float64(i+1))), 0, float32(math.Sin(dr * float64(i+1)))}.Mul(float32(branch.radius - radiusDec)).Add(end)
			normal := mgl32.Vec3{float32(math.Cos(dr * float64(i))), 0, float32(math.Sin(dr * float64(i)))}
			normal = normal.Normalize()
			color := mgl32.Vec4{0.5, 0.5, 0.1, 1.0}
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p1, Normal: normal, Color: color, Texture: mgl32.Vec2{0.0, 0.0}})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p2, Normal: normal, Color: color, Texture: mgl32.Vec2{1.0, 0.0}})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p3, Normal: normal, Color: color, Texture: mgl32.Vec2{0.0, 1.0}})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p4, Normal: normal, Color: color, Texture: mgl32.Vec2{1.0, 1.0}})
			t1 := gfx.TriangleConnectivity{U0: index, U1: index + 1, U2: index + 3}
			t2 := gfx.TriangleConnectivity{U0: index, U1: index + 3, U2: index + 2}
			mesh.Connectivity = append(mesh.Connectivity, t1)
			mesh.Connectivity = append(mesh.Connectivity, t2)
			index += 4
		}
	}

	model := gfx.BuildModel(mesh)
	return &model
}

func (t *Tree) generateFromGrammar() {
	rootBranches := []Branch{}
	branches := []Branch{}
	leaves := []Branch{}
	branch := Branch{radius: 0.005, height: -0.05}
	addSomething := false

	for _, letter := range strings.Split(t.grammar, "") {
		switch letter {
		case "F":
			//FIXME: fix branch height position
			if addSomething {
				toAdd := rotateZ(branch.angleZ, mgl32.Vec3{0, branch.height, 0})
				toAdd = rotateY(branch.angleY, toAdd)
				branch.position = branch.position.Add(toAdd)
			}
			branch.radius -= branch.radius / 5.0
			addSomething = true
			branches = append(branches, branch)
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
			if len(rootBranches) == 0 {
				branch.angleY += float32(rand.Float64() * 360.0)
			}
			rootBranches = append(rootBranches, branch) //push
			break
		case "]":
			leaves = append(leaves, branch)
			branch, rootBranches = rootBranches[len(rootBranches)-1], rootBranches[:len(rootBranches)-1] //pop
			break
		}
	}

	t.BranchesModel = createBranchesModel(branches)
	t.LeavesModel = createLeavesModel(leaves)
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

func translate(translate mgl32.Vec3, origin mgl32.Vec3) mgl32.Vec3 {
	transform := mgl32.Translate3D(translate.X(), translate.Y(), translate.Z())
	origin4 := mgl32.Vec4{origin.X(), origin.Y(), origin.Z(), 1.0}
	return transform.Mul4x1(origin4).Vec3()
}
