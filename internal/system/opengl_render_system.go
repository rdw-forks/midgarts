package system

import (
	_ "embed"

	"github.com/EngoEngine/ecs"
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/project-midgard/midgarts/internal/camera"
	"github.com/project-midgard/midgarts/internal/graphic"
	"github.com/project-midgard/midgarts/internal/opengl"
	"github.com/project-midgard/midgarts/internal/system/rendercmd"
)

//go:embed shader/box.vert
var boxVertexShader string

//go:embed shader/box.frag
var boxFragmentShader string

//go:embed shader/sprite.vert
var spriteVertexShader string

//go:embed shader/sprite.frag
var spriteFragmentShader string

type RenderCommands struct {
	sprite []rendercmd.SpriteRenderCommand
}

// OpenGLRenderSystem defines an OpenGL-based rendering system.
type OpenGLRenderSystem struct {
	cam            *camera.Camera
	renderCommands *RenderCommands

	// Buffer of reusable sprites
	spritesBuf []*graphic.Sprite
}

func NewOpenGLRenderSystem(cam *camera.Camera, commands *RenderCommands) *OpenGLRenderSystem {
	return &OpenGLRenderSystem{
		cam:            cam,
		renderCommands: commands,
		spritesBuf:     []*graphic.Sprite{},
	}
}

func (s *OpenGLRenderSystem) Update(dt float32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// 2D Sprite Box
	s.renderSpriteBoxes()

	// 2D Sprites
	s.renderSprites()
}
func (s *OpenGLRenderSystem) renderSpriteBoxes() {
	shader := opengl.NewShader(boxVertexShader, boxFragmentShader)
	pid := shader.Program().ID()
	gl.UseProgram(pid)

	s.EnsureSpritesBufLen(len(s.renderCommands.sprite))

	for i, cmd := range s.renderCommands.sprite {
		if cmd.FlipVertically {
			cmd.Size = mgl32.Vec2{-cmd.Size.X(), cmd.Size.Y()}
		}

		sprite := s.spritesBuf[i]
		sprite.SetBounds(cmd.Size.X(), cmd.Size.Y())
		sprite.SetPosition(mgl32.Vec3{cmd.Position.X(), cmd.Position.Y(), cmd.Position.Z()})

		view := s.cam.ViewMatrix()
		viewu := gl.GetUniformLocation(pid, gl.Str("view\x00"))
		gl.UniformMatrix4fv(viewu, 1, false, &view[0])

		model := sprite.Model()
		modelu := gl.GetUniformLocation(pid, gl.Str("model\x00"))
		gl.UniformMatrix4fv(modelu, 1, false, &model[0])

		projection := s.cam.ProjectionMatrix()
		projectionu := gl.GetUniformLocation(pid, gl.Str("projection\x00"))
		gl.UniformMatrix4fv(projectionu, 1, false, &projection[0])

		sizeu := gl.GetUniformLocation(pid, gl.Str("size\x00"))
		gl.Uniform2fv(sizeu, 1, &cmd.Size[0])

		iden := mgl32.Ident4()
		rotation := iden.Mul4(mgl32.HomogRotate3D(cmd.RotationRadians, graphic.Backwards))
		rotationu := gl.GetUniformLocation(pid, gl.Str("rotation\x00"))
		gl.UniformMatrix4fv(rotationu, 1, false, &rotation[0])

		gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
		sprite.Render(shader)
	}
}

func (s *OpenGLRenderSystem) renderSprites() {
	shader := opengl.NewShader(spriteVertexShader, spriteFragmentShader)

	pid := shader.Program().ID()
	gl.UseProgram(pid)
	s.EnsureSpritesBufLen(len(s.renderCommands.sprite))

	for i, cmd := range s.renderCommands.sprite {
		if cmd.FlipVertically {
			cmd.Size = mgl32.Vec2{-cmd.Size.X(), cmd.Size.Y()}
		}

		sprite := s.spritesBuf[i]
		sprite.SetBounds(cmd.Size.X(), cmd.Size.Y())
		sprite.SetTexture(cmd.Texture)
		sprite.SetPosition(mgl32.Vec3{cmd.Position.X(), cmd.Position.Y(), cmd.Position.Z()})
		sprite.Texture.Bind(0)

		view := s.cam.ViewMatrix()
		viewu := gl.GetUniformLocation(pid, gl.Str("view\x00"))
		gl.UniformMatrix4fv(viewu, 1, false, &view[0])

		model := sprite.Model()
		modelu := gl.GetUniformLocation(pid, gl.Str("model\x00"))
		gl.UniformMatrix4fv(modelu, 1, false, &model[0])

		projection := s.cam.ProjectionMatrix()
		projectionu := gl.GetUniformLocation(pid, gl.Str("projection\x00"))
		gl.UniformMatrix4fv(projectionu, 1, false, &projection[0])

		sizeu := gl.GetUniformLocation(pid, gl.Str("size\x00"))
		gl.Uniform2fv(sizeu, 1, &cmd.Size[0])

		offsetu := gl.GetUniformLocation(pid, gl.Str("offset\x00"))
		gl.Uniform2fv(offsetu, 1, &cmd.Offset[0])

		iden := mgl32.Ident4()
		rotation := iden.Mul4(mgl32.HomogRotate3D(cmd.RotationRadians, graphic.Backwards))
		rotationu := gl.GetUniformLocation(pid, gl.Str("rotation\x00"))
		gl.UniformMatrix4fv(rotationu, 1, false, &rotation[0])

		gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
		sprite.Render(shader)
		sprite.Texture.Unbind(0)
	}
}

func (s *OpenGLRenderSystem) Remove(e ecs.BasicEntity) {
	panic("implement me")
}

func (s *OpenGLRenderSystem) EnsureSpritesBufLen(minLen int) {
	s.spritesBuf = ensureSpritesBufferLength(s.spritesBuf, minLen)
}

func ensureSpritesBufferLength(slice []*graphic.Sprite, minLen int) []*graphic.Sprite {
	oldLen := len(slice)

	if cacheOverflow := minLen - oldLen; cacheOverflow <= 0 {
		// no need to resize
		return slice
	}

	if minLen > cap(slice) {
		newSlice := make([]*graphic.Sprite, oldLen, minLen)
		copy(newSlice, slice)
		slice = newSlice
	}

	slice = slice[0:minLen]

	for i := oldLen; i < minLen; i++ {
		slice[i] = graphic.NewSprite(0, 0, nil)
	}
	return slice
}
