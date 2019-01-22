package veg

import (
	"../gfx"
	"../ter"
)

var (
	uniqueTrees []*Tree
	uniqueGrass *gfx.Model
)

type Gaia struct {
	InstanceTrees []*InstanceTree
	InstanceGrass *InstanceGrass
}

func InitialiseVegetation(step float32) *Gaia {
	g := Gaia{}
	uniqueTrees = CreateUniqueTrees()
	uniqueGrass = CreateUniqueGrass(step)
	g.ResetVegetation()
	return &g
}

func (g *Gaia) CreateChunkVegetation(chunk *ter.Chunk, currentChunk [2]int) {
	if currentChunk == chunk.Position {
		g.InstanceTrees = GetSurroundingForests(g.InstanceTrees, chunk, true)
	} else {
		g.InstanceTrees = GetSurroundingForests(g.InstanceTrees, chunk, false)
	}
	g.InstanceGrass = GetSurroundingGrass(g.InstanceGrass, chunk)
}

func (g *Gaia) ResetVegetation() {
	g.InstanceTrees = []*InstanceTree{}
	g.InstanceGrass = nil
	for _, uniqueTree := range uniqueTrees {
		g.InstanceTrees = append(g.InstanceTrees, &InstanceTree{Parent: uniqueTree})
	}
	g.InstanceGrass = &InstanceGrass{Model: uniqueGrass}
}
