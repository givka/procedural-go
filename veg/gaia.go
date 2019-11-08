package veg

import (
	"math"
	"math/rand"

	"github.com/givka/procedural-go/gfx"
	"github.com/givka/procedural-go/ter"
	"github.com/go-gl/mathgl/mgl32"
)

type Gaia struct {
	InstanceTrees [][2]*InstanceTree
	InstanceGrass *InstanceGrass
}

func InitialiseVegetation(step float32) *Gaia {
	uniqueTrees := createUniqueTrees()
	uniqueGrass := createUniqueGrass(step)
	return &Gaia{
		InstanceGrass: getInstanceGrass(uniqueGrass),
		InstanceTrees: getInstanceTrees(uniqueTrees),
	}
}

func getInstanceGrass(uniqueGrass *gfx.Model) *InstanceGrass {
	return &InstanceGrass{Model: uniqueGrass}
}

func getInstanceTrees(uniqueTrees [][2]*Tree) [][2]*InstanceTree {
	var instanceTrees [][2]*InstanceTree
	for _, uniqueTree := range uniqueTrees {
		instanceTrees = append(instanceTrees, [2]*InstanceTree{
			&InstanceTree{
				BranchesModel: uniqueTree[0].BranchesModel,
				LeavesModel:   uniqueTree[0].LeavesModel,
			},
			&InstanceTree{
				BranchesModel: uniqueTree[1].BranchesModel,
				LeavesModel:   uniqueTree[1].LeavesModel,
			},
		})
	}
	return instanceTrees
}

func (g *Gaia) CreateChunkVegetation(chunk *ter.Chunk, currentChunk [2]int) {
	// start := time.Now()
	chunk.IsHQ = isCloseToCurrentChunk(chunk, currentChunk)
	if chunk.IsHQ {
		g.InstanceGrass.Transforms = append(g.InstanceGrass.Transforms, chunk.GrassTransforms...)
		gfx.ModelToInstanceModel(g.InstanceGrass.Model, g.InstanceGrass.Transforms)
	}
	for _, transform := range chunk.TreesTransforms {
		index := rand.Intn(len(g.InstanceTrees))
		chunk.TreesModelID = append(chunk.TreesModelID, index)
		if chunk.IsHQ {
			g.InstanceTrees[index][0].Transforms = append(g.InstanceTrees[index][0].Transforms, transform)
		} else {
			g.InstanceTrees[index][1].Transforms = append(g.InstanceTrees[index][1].Transforms, transform)
		}

	}
	for _, instanceTree := range g.InstanceTrees {
		gfx.ModelToInstanceModel(instanceTree[0].BranchesModel, instanceTree[0].Transforms)
		gfx.ModelToInstanceModel(instanceTree[0].LeavesModel, instanceTree[0].Transforms)
		gfx.ModelToInstanceModel(instanceTree[1].BranchesModel, instanceTree[1].Transforms)
		gfx.ModelToInstanceModel(instanceTree[1].LeavesModel, instanceTree[1].Transforms)
	}
	// fmt.Println(
	// 	"new chunk", time.Now().Sub(start),
	// 	"grass:", len(chunk.GrassTransforms),
	// 	"trees:", len(chunk.TreesTransforms),
	// )
	chunk.HasVegetation = true
}

func (g *Gaia) ResetInstanceTransfoms() {
	g.InstanceGrass.Transforms = []mgl32.Mat4{}
	for _, instanceTree := range g.InstanceTrees {
		instanceTree[0].Transforms = []mgl32.Mat4{}
		instanceTree[1].Transforms = []mgl32.Mat4{}
	}
}

func (g *Gaia) RedrawAllChunks(chunks []*ter.Chunk, currentChunk [2]int) {
	// start := time.Now()
	for _, chunk := range chunks {
		chunk.IsHQ = isCloseToCurrentChunk(chunk, currentChunk)
		if !chunk.HasVegetation {
			continue
		}
		if chunk.IsHQ {
			g.InstanceGrass.Transforms = append(g.InstanceGrass.Transforms, chunk.GrassTransforms...)
		}
		for i, transform := range chunk.TreesTransforms {
			index := chunk.TreesModelID[i]
			if chunk.IsHQ {
				g.InstanceTrees[index][0].Transforms = append(g.InstanceTrees[index][0].Transforms, transform)
			} else {
				g.InstanceTrees[index][1].Transforms = append(g.InstanceTrees[index][1].Transforms, transform)
			}
		}
	}

	gfx.ModelToInstanceModel(g.InstanceGrass.Model, g.InstanceGrass.Transforms)
	nbrTrees := 0
	for _, instanceTree := range g.InstanceTrees {
		gfx.ModelToInstanceModel(instanceTree[0].BranchesModel, instanceTree[0].Transforms)
		gfx.ModelToInstanceModel(instanceTree[0].LeavesModel, instanceTree[0].Transforms)
		gfx.ModelToInstanceModel(instanceTree[1].BranchesModel, instanceTree[1].Transforms)
		gfx.ModelToInstanceModel(instanceTree[1].LeavesModel, instanceTree[1].Transforms)
		nbrTrees += len(instanceTree[0].Transforms) + len(instanceTree[1].Transforms)
	}
	// fmt.Println(
	// 	"redraw all", time.Now().Sub(start),
	// 	"grass:", len(g.InstanceGrass.Transforms),
	// 	"trees:", nbrTrees,
	// )
}

func isCloseToCurrentChunk(chunk *ter.Chunk, currentChunk [2]int) bool {
	if math.Abs(float64(chunk.Position[0]-currentChunk[0])) <= 1 &&
		math.Abs(float64(chunk.Position[1]-currentChunk[1])) <= 1 {
		return true
	}
	return false

}
