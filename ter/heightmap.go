package ter

import (
	"../gfx"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
)

type HeightMap struct {
	ChunkNBPoints  uint32
	ChunkWorldSize uint32
	NbOctaves      uint32
	Exponent 	   float64
	Chunks         map[[2]int]*Chunk

	Perlin         noiselib.Perlin

	//mountains
	RidgedMulti    noiselib.Ridgedmulti
	//flat terrain
	Billow		   noiselib.Billow
	ScaleBias      noiselib.ScaleBias
	//terrain type selector
	TerrainType    noiselib.Perlin
	//Final Terrain
	FinalTerrain   noiselib.Select
}

func getMapValue(heightMap *HeightMap, position [2] float64) float64{
	return heightMap.FinalTerrain.GetValue(position[0], 0, position[1])
}

func CreateChunkPolyMesh(chunk Chunk) gfx.Mesh {
	mesh := gfx.Mesh{}
	size := int(chunk.NBPoints)

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)

	//first add all vertices

	for x:=0; x < size+1; x++{
		for z:=0; z < size+1; z++ {
			position := mgl32.Vec3{float32(x) * step, float32(-chunk.Map[x + z * (size+1)]), float32(z) * step}
			//compute normal
			var up float64 = 0.0
			var down float64 = 0.0
			var left float64 = 0.0
			var right float64 = 0.0
			if z > 0 		{up 	= -chunk.Map[x + (z-1) * (size+1)]}
			if z < size		{down	= -chunk.Map[x + (z+1) * (size+1)]}
			if x > 0 		{left 	= -chunk.Map[x - 1 + z * (size+1)]}
			if x < size 	{right	= -chunk.Map[x + 1 + z * (size+1)]}

			normal := mgl32.Vec3{float32(right - left), -2, float32(down - up)}
			normal = normal.Normalize()

			var color mgl32.Vec4
			height := -position.Y()
			if(height > 1) {
				color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
			}else if(height > 0.5){
				color = mgl32.Vec4{0.4, 0.4, 0.4, 1.0}
			}else if(height > -0.2){
				color = mgl32.Vec4{0.0, 0.5, 0.0, 1.0}
			}else if height > -0.5{
				color = mgl32.Vec4{0.0, 0.3, 0.0, 1.0}
			}else if height > -1{
				color = mgl32.Vec4{0.0, 0.0, 0.5, 1.0}
			}else{
				color = mgl32.Vec4{0.5, 0.5, 0.0, 1.0}
			}
			texture := mgl32.Vec2{0.0, 0.0}

			v := gfx.Vertex{
				Position: position,
				Normal:   normal,
				Color:    color,
				Texture:  texture}
			mesh.Vertices = append(mesh.Vertices, v)
		}
	}

	//then build triangles
	for x := 0; x < size; x++ {
		for z := 0; z < size; z++ {
			i := uint32(x) + (chunk.NBPoints+1)*uint32(z)
			tri1 := gfx.TriangleConnectivity{i, i + 1, i + uint32(chunk.NBPoints+1)}
			tri2 := gfx.TriangleConnectivity{i + 1, i + uint32(chunk.NBPoints+1) + 1, i + uint32(chunk.NBPoints+1)}
			mesh.Connectivity = append(mesh.Connectivity, tri1)
			mesh.Connectivity = append(mesh.Connectivity, tri2)
		}
	}

	return mesh
}
