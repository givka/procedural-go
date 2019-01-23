package veg

import (
	"math"
	"math/rand"

	"../gfx"
	"../ter"
	"github.com/go-gl/mathgl/mgl32"
)

type Gaia struct {
	InstanceTrees []*InstanceTree
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

func getInstanceTrees(uniqueTrees []*Tree) []*InstanceTree {
	var instanceTrees []*InstanceTree
	for _, uniqueTree := range uniqueTrees {
		instanceTrees = append(instanceTrees, &InstanceTree{
			BranchesModel: uniqueTree.BranchesModel,
			LeavesModel:   uniqueTree.LeavesModel,
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
		index := rand.Intn(len(g.InstanceTrees) - 1)
		chunk.TreesModelID = append(chunk.TreesModelID, index)
		if !chunk.IsHQ {
			index = len(g.InstanceTrees) - 1
		}
		g.InstanceTrees[index].Transforms = append(g.InstanceTrees[index].Transforms, transform)
	}
	for _, instanceTree := range g.InstanceTrees {
		gfx.ModelToInstanceModel(instanceTree.BranchesModel, instanceTree.Transforms)
		gfx.ModelToInstanceModel(instanceTree.LeavesModel, instanceTree.Transforms)
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
		instanceTree.Transforms = []mgl32.Mat4{}
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
			gfx.ModelToInstanceModel(g.InstanceGrass.Model, g.InstanceGrass.Transforms)
			for i, transform := range chunk.TreesTransforms {
				index := chunk.TreesModelID[i]
				g.InstanceTrees[index].Transforms = append(g.InstanceTrees[index].Transforms, transform)
			}
		} else {
			for _, transform := range chunk.TreesTransforms {
				index := len(g.InstanceTrees) - 1
				g.InstanceTrees[index].Transforms = append(g.InstanceTrees[index].Transforms, transform)
			}
		}
	}
	nbrTrees := 0
	for _, instanceTree := range g.InstanceTrees {
		gfx.ModelToInstanceModel(instanceTree.BranchesModel, instanceTree.Transforms)
		gfx.ModelToInstanceModel(instanceTree.LeavesModel, instanceTree.Transforms)
		nbrTrees += len(instanceTree.Transforms)
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
