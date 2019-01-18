package scr

import (
	"math"

	"../cam"
	"../ctx"
	"../gfx"
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
	gl.Uniform3f(m.Program.GetUniformLocation("lightPos"), c.Position().X(), c.Position().Y(), c.Position().Z())
	gl.Uniform1i(m.Program.GetUniformLocation("textureId"), int32(m.TextureID))

	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElements(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}

func RenderForest(tree *veg.Tree, camera *cam.FpsCamera, program *gfx.Program, nbrTrees int) {
	if nbrTrees == 0 {
		return
	}
	speed := 2.5
	amp := float32(2.5)
	angle := mgl32.DegToRad(amp * float32(math.Cos(speed*glfw.GetTime())))
	transform := mgl32.Rotate3DX(angle).Mul3(mgl32.Rotate3DX(angle)).Mat4()

	tree.BranchesModel.Program = program
	tree.BranchesModel.TextureID = gl.TEXTURE1
	tree.BranchesModel.Transform = transform
	RenderInstances(tree.BranchesModel, camera, nbrTrees)

	tree.LeavesModel.Program = program
	tree.LeavesModel.TextureID = gl.TEXTURE2
	tree.LeavesModel.Transform = transform
	RenderInstances(tree.LeavesModel, camera, nbrTrees)
}

func RenderInstances(m *gfx.Model, camera *cam.FpsCamera, nbrInstances int) {
	m.Program.Use()

	initialiseUniforms(m, camera)

	gl.BindVertexArray(m.VAO)
	gl.DrawElementsInstanced(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, gl.Ptr(m.LoadingData.Connectivity), int32(nbrInstances))
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
	gl.Uniform3f(m.Program.GetUniformLocation("lightPos"), camera.Position().X(), camera.Position().Y(), camera.Position().Z())
	gl.Uniform1i(m.Program.GetUniformLocation("textureId"), int32(m.TextureID))

}
