package sky

import (
	"math"

	"../cam"
	"../gfx"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Dome struct {
	Model         *gfx.Model
	Radius        float32
	SunPosition   mgl32.Vec3
	LightPosition mgl32.Vec3
}

func CreateDome(program *gfx.Program, textureId uint32) *Dome {
	mesh := gfx.Mesh{}
	nU := 100
	nV := 100
	radius := float32(50)
	startU := float32(0)
	startV := float32(math.Pi / 4.0)
	endU := float32(math.Pi * 2.0)
	endV := float32(math.Pi)
	stepU := (endU - startU) / float32(nU)
	stepV := (endV - startV) / float32(nV)
	index := uint32(0)

	for i := 0; i < nU; i++ { // U-points
		for j := 0; j < nV; j++ { // V-points
			var un float32
			var vn float32
			u := float32(i)*stepU + startU
			v := float32(j)*stepV + startV

			if i+1 == nU {
				un = endU
			} else {
				un = float32(i+1)*stepU + startU
			}
			if j+1 == nV {
				vn = endV
			} else {
				vn = float32(j+1)*stepV + startV
			}
			p1 := getSpherePosition(u, v, radius)
			p2 := getSpherePosition(u, vn, radius)
			p3 := getSpherePosition(un, v, radius)
			p4 := getSpherePosition(un, vn, radius)

			normal := mgl32.Vec3{0, -2, 0}
			normal = normal.Normalize()
			color := mgl32.Vec4{1.0, 1.0, 1.0, 1.0}
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p1, Normal: normal, Color: color})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p2, Normal: normal, Color: color})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p3, Normal: normal, Color: color})
			mesh.Vertices = append(mesh.Vertices, gfx.Vertex{Position: p4, Normal: normal, Color: color})
			t1 := gfx.TriangleConnectivity{U0: index, U1: index + 1, U2: index + 3}
			t2 := gfx.TriangleConnectivity{U0: index, U1: index + 3, U2: index + 2}
			mesh.Connectivity = append(mesh.Connectivity, t1)
			mesh.Connectivity = append(mesh.Connectivity, t2)
			index += 4
		}
	}
	model := gfx.BuildModel(mesh)
	model.Program = program
	model.TextureID = textureId
	return &Dome{Model: &model, Radius: radius}
}

func getSpherePosition(u float32, v float32, r float32) mgl32.Vec3 {
	uu := float64(u)
	vv := float64(v)

	return mgl32.Vec3{
		float32(math.Cos(uu)*math.Sin(vv)) * r,
		float32(math.Cos(vv)) * r,
		float32(math.Sin(uu)*math.Sin(vv)) * r,
	}
}

func (d *Dome) UpdateSun(camera *cam.FpsCamera) {
	rotation := mgl32.Rotate3DX(mgl32.DegToRad(float32(glfw.GetTime() * 10.0)))
	d.SunPosition = rotation.Mul3x1(mgl32.Vec3{0.0, d.Radius, 0.0})
	d.LightPosition = mgl32.Vec3{
		camera.Position().X() + d.SunPosition.X(),
		d.SunPosition.Y() / 2.0,
		camera.Position().X() + d.SunPosition.Z(),
	}
}
