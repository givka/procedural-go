package scr

import (
	"math"

	"../cam"
	"../ctx"
	"../gfx"
	"../sky"
	"../ter"
	"../veg"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func RenderChunks(chunks []*ter.Chunk, camera *cam.FpsCamera, program *gfx.Program) {
	for _, chunk := range chunks {
		chunk.Model.Program = program
		RenderModel(chunk.Model, camera)
	}
}

func RenderSky(dome *sky.Dome, camera *cam.FpsCamera) {
	model := dome.Model
	program := model.Program
	program.Use()

	u := float32(math.Pi*math.Cos(glfw.GetTime()/5.0) + math.Pi)
	v := float32(math.Pi*math.Cos(glfw.GetTime()/5.0) + math.Pi)
	sunPos := sky.GetSpherePosition(u, v, dome.Radius)
	hours := float32(6.0*math.Cos(glfw.GetTime()/5.0) + 6.0)

	pvm := getPVM(model, camera)

	// fmt.Println(sunPos)

	gl.UniformMatrix4fv(program.GetUniformLocation("pvm"), 1, false, &pvm[0])
	gl.Uniform1f(program.GetUniformLocation("hours"), hours)
	gl.Uniform3f(program.GetUniformLocation("sun_pos"), sunPos.X(), sunPos.Y(), sunPos.Z())
	gl.Uniform1f(program.GetUniformLocation("radius"), dome.Radius)
	gl.Uniform1i(program.GetUniformLocation("currentTexture"), int32(model.TextureID-gl.TEXTURE0))

	gl.BindVertexArray(model.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, model.Connectivity)
	gl.DrawElements(gl.TRIANGLES, model.NbTriangles*3, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

func RenderVegetation(g *veg.Gaia, camera *cam.FpsCamera, program *gfx.Program) {
	speed := 2.5
	amp := float32(2.5)
	angle := mgl32.DegToRad(amp * float32(math.Cos(speed*glfw.GetTime())))
	transform := mgl32.Rotate3DX(angle).Mul3(mgl32.Rotate3DX(angle)).Mat4()

	g.InstanceGrass.Model.Program = program
	g.InstanceGrass.Model.Transform = transform

	RenderInstances(g.InstanceGrass.Model, camera, len(g.InstanceGrass.Transforms))

	for _, instanceTree := range g.InstanceTrees {
		instanceTree.Parent.BranchesModel.Program = program
		instanceTree.Parent.BranchesModel.TextureID = gl.TEXTURE1
		instanceTree.Parent.BranchesModel.Transform = transform
		RenderInstances(instanceTree.Parent.BranchesModel, camera, len(instanceTree.Transforms))

		instanceTree.Parent.LeavesModel.Program = program
		instanceTree.Parent.LeavesModel.TextureID = gl.TEXTURE2
		instanceTree.Parent.LeavesModel.Transform = transform
		RenderInstances(instanceTree.Parent.LeavesModel, camera, len(instanceTree.Transforms))
	}

}

func RenderModel(m *gfx.Model, c *cam.FpsCamera) {
	m.Program.Use()

	view := c.GetTransform()
	project := mgl32.Perspective(mgl32.DegToRad(ctx.Fov),
		float32(ctx.Width())/float32(ctx.Height()), ctx.Near, ctx.Far)

	gl.UniformMatrix4fv(m.Program.GetUniformLocation("view"), 1, false, &view[0])
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("project"), 1, false, &project[0])
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("model"), 1, false, &m.Transform[0])
	gl.Uniform1f(m.Program.GetUniformLocation("near"), ctx.Near)
	gl.Uniform1f(m.Program.GetUniformLocation("far"), ctx.Far)
	gl.Uniform3f(m.Program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
	gl.Uniform3f(m.Program.GetUniformLocation("lightPos"), c.Position().X(), -50.0, c.Position().Z())
	gl.Uniform1i(m.Program.GetUniformLocation("textureId"), int32(m.TextureID))

	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElements(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}

func RenderInstances(m *gfx.Model, camera *cam.FpsCamera, nbrInstances int) {
	if nbrInstances == 0 {
		return
	}

	m.Program.Use()
	initialiseUniforms(m, camera)
	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElementsInstanced(gl.TRIANGLES, m.NbTriangles, gl.UNSIGNED_INT, nil, int32(nbrInstances))
	gl.BindVertexArray(0)
}

func initialiseUniforms(m *gfx.Model, camera *cam.FpsCamera) {
	view := camera.GetTransform()
	project := mgl32.Perspective(mgl32.DegToRad(ctx.Fov),
		float32(ctx.Width())/float32(ctx.Height()), ctx.Near, ctx.Far)

	gl.Uniform1i(m.Program.GetUniformLocation("currentTexture"), int32(m.TextureID-gl.TEXTURE0))
	gl.Uniform1f(m.Program.GetUniformLocation("near"), ctx.Near)
	gl.Uniform1f(m.Program.GetUniformLocation("far"), ctx.Far)
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("view"), 1, false, &view[0])
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("project"), 1, false, &project[0])
	gl.UniformMatrix4fv(m.Program.GetUniformLocation("model"), 1, false, &m.Transform[0])
	gl.Uniform3f(m.Program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
	gl.Uniform3f(m.Program.GetUniformLocation("lightPos"), camera.Position().X(), -50.0, camera.Position().Z())
	gl.Uniform1i(m.Program.GetUniformLocation("textureId"), int32(m.TextureID))

}

func getPVM(m *gfx.Model, camera *cam.FpsCamera) mgl32.Mat4 {
	view := camera.GetTransform()
	project := mgl32.Perspective(mgl32.DegToRad(ctx.Fov), float32(ctx.Width())/float32(ctx.Height()), ctx.Near, ctx.Far)
	model := m.Transform
	return project.Mul4(view).Mul4(model)
}
