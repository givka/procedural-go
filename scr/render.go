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

func RenderChunks(chunks []*ter.Chunk, camera *cam.FpsCamera, program *gfx.Program, textureContainer *ter.ChunkTextureContainer) {
	for _, chunk := range chunks {
		chunk.Model.Program = program
		RenderChunkModel(chunk.Model, camera, textureContainer)
	}
}

func RenderChunkModel(m *gfx.Model, c *cam.FpsCamera, textureContainer *ter.ChunkTextureContainer){
	m.Program.Use()
	initialiseUniforms(m, c)
	setChunkTextureUniforms(m, textureContainer)
	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElements(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}
func RenderModel(m *gfx.Model, c *cam.FpsCamera) {
	m.Program.Use()
	initialiseUniforms(m, c)

	gl.BindVertexArray(m.VAO)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElements(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil)

	gl.BindVertexArray(0)
}

func setChunkTextureUniforms(m *gfx.Model, textureContainer *ter.ChunkTextureContainer) {
	gl.Uniform1i(m.Program.GetUniformLocation("dirtTexture"), int32(textureContainer.DirtID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("snowTexture"), int32(textureContainer.SnowID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("sandTexture"), int32(textureContainer.SandID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("rockTexture"), int32(textureContainer.RockID-gl.TEXTURE0))
	gl.Uniform1i(m.Program.GetUniformLocation("grassTexture"), int32(textureContainer.GrassID-gl.TEXTURE0))

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
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.Connectivity)
	gl.DrawElementsInstanced(gl.TRIANGLES, m.NbTriangles*3, gl.UNSIGNED_INT, nil, int32(nbrInstances))
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
