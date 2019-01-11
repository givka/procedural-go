package gfx

import (
	"github.com/go-gl/mathgl/mgl32"
)
import "github.com/go-gl/gl/v4.1-core/gl"

func Render(m Model, view mgl32.Mat4, project mgl32.Mat4) {
	lightPos := mgl32.Vec3{10, -5, 10} //temporary

	program := m.Program

	program.Use()
	gl.UniformMatrix4fv(program.GetUniformLocation("view"), 1, false, &view[0])
	gl.UniformMatrix4fv(program.GetUniformLocation("project"), 1, false, &project[0])
	gl.UniformMatrix4fv(program.GetUniformLocation("model"), 1, false,	&m.Transform[0])

	gl.BindVertexArray(m.VAO)

	gl.Uniform3f(program.GetUniformLocation("lightColor"), 1.0, 1.0, 1.0)
	gl.Uniform3f(program.GetUniformLocation("lightPos"), lightPos.X(), lightPos.Y(), lightPos.Z())

	gl.DrawElements(gl.TRIANGLES, m.NbTriangles, gl.UNSIGNED_INT, gl.Ptr(m.Indices))

	gl.BindVertexArray(0)
}
