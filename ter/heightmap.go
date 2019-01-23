package ter

import (
	"../gfx"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/worldsproject/noiselib"
	"math"
)

type HeightMap struct {
	ChunkNBPoints  uint32
	ChunkWorldSize uint32
	NbOctaves      uint32
	Exponent 	   float64
	Chunks         map[[2]int]*Chunk

	Perlin         noiselib.Perlin

	//mountains
	MountainNoise noiselib.Ridgedmulti
	MountainScaleBias noiselib.ScaleBias
	//rivers
	RiverNoise noiselib.Ridgedmulti
	RiverAbs   noiselib.Abs
	RiverScaleBias noiselib.ScaleBias
	RiverCurve noiselib.Curve
	RiverClamp noiselib.Clamp

	//flat terrain
	PlainNoise     noiselib.Billow
	PlainScaleBias noiselib.ScaleBias
	//rivers in flat terrain
	PlainAndRiver noiselib.Select
	//terrain type selector
	TerrainType    noiselib.Perlin
	//Final Terrain
	FinalTerrain   noiselib.Select
}

func getMapValue(heightMap *HeightMap, position [2] float64) float64{
	return heightMap.FinalTerrain.GetValue(position[0], 0, position[1])
}

func CreateChunkPolyMesh(chunk Chunk, textureContainer *ChunkTextureContainer, heightMap *HeightMap) gfx.Mesh {
	mesh := gfx.Mesh{}
	size := int(chunk.NBPoints)

	step := float32(chunk.WorldSize) / float32(chunk.NBPoints)

	//first add all vertices

	var textureID uint32 = 0
	for x:=0; x < size+1; x++{
		for z:=0; z < size+1; z++ {
			position := mgl32.Vec3{float32(x) * step, float32(-chunk.Map[x + z * (size+1)]), float32(z) * step}
			//compute normal
			var up float64 = 0.0
			var down float64 = 0.0
			var left float64 = 0.0
			var right float64 = 0.0
			{
				//up = -heightMap.FinalTerrain.GetValue(float64(x + size * chunk.Position[0])*float64(step), 0, float64(z + size * chunk.Position[1] - 1)*float64(step))

				if z > 10 {
					up = -chunk.Map[x+(z-1)*(size+1)]
				} else {
					up = -heightMap.FinalTerrain.GetValue(float64(x + size * chunk.Position[0])*float64(step), 0, float64(z + size * chunk.Position[1] - 1)*float64(step))
				}
				if z < size-10 {
					down = -chunk.Map[x+(z+1)*(size+1)]
				} else {
					down = -heightMap.FinalTerrain.GetValue(float64(x + size * chunk.Position[0])*float64(step), 0, float64(z + size * chunk.Position[1] + 1)*float64(step))
				}
				if x > 10 {
					left = -chunk.Map[x-1+z*(size+1)]
				} else {
					left = -heightMap.FinalTerrain.GetValue(float64(x + size * chunk.Position[0] - 1)*float64(step), 0, float64(z + size * chunk.Position[1])*float64(step))
				}
				if x < size-10 {
					right = -chunk.Map[x+1+z*(size+1)]
				} else {
					right = -heightMap.FinalTerrain.GetValue(float64(x + size * chunk.Position[0] + 1)*float64(step), 0, float64(z + size * chunk.Position[1])*float64(step))
				}
			}
			normal := mgl32.Vec3{float32(left - right) / step, -2, float32(down - up) / step}
			normal = normal.Normalize()

			var color mgl32.Vec4
			height := -position.Y()
			if(height > 1.5) {
				color = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
				textureID = textureContainer.SnowID
			}else if(height > 0.8){
				color = mgl32.Vec4{0.4, 0.4, 0.4, 1.0}
				textureID = textureContainer.RockID
			}else if(height > -0.2){
				color = mgl32.Vec4{0.0, 0.4, 0.0, 1.0}
				textureID = textureContainer.GrassID
			}else if height > -0.5{
				color = mgl32.Vec4{0.7, 0.7, 0.0, 1.0}
				textureID = textureContainer.SandID
			}else if height > -1{
				color = mgl32.Vec4{0.0, 0.3, 0.5, 1.0}
				//textureID = textureContainer.SnowID
			}else{
				color = mgl32.Vec4{0.0, 0.0, 0.7, 1.0}
				//textureID = textureContainer.SnowID
			}
			color = mgl32.Vec4{color.X(), color.Y(), color.Z(), -float32(chunk.WaterMap[x + z * (size+1)])}
			chunk.NormalY[x + z * (size + 1)] = float64(normal.Y())
			var textureScale float64 = 1.0/16.0

			texture := mgl32.Vec2{float32(   math.Mod(( (float64(x)/float64(size+1)) / textureScale) , 1.0)  ), float32(   math.Mod(((float64(z)/float64(size+1)) / textureScale) , 1.0)  )}

			v := gfx.Vertex{
				Position: position,
				Normal:   normal,
				Color:    color,
				Texture:  texture,
			}
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

	mesh.TextureID = textureID
	return mesh
}
