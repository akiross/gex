package main

import (
	"fmt"
	glad "github.com/akiross/go-glad"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"io/ioutil"
	"math/rand"
	"strings"
)

func LoadFile(path string, repl ...interface{}) string {
	cont, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	res := string(cont)

	for i := 0; i < len(repl); i += 2 {
		sub, obj := fmt.Sprint(repl[i]), fmt.Sprint(repl[i+1])
		res = strings.Replace(res, sub, obj, -1)
	}

	return res
}

func CreateWindow(w, h int, title string) *glfw.Window {
	return NewOGLWindow(w, h, title,
		CoreProfile(true),
		ForwardCompatible(true),
		Resizable(false),
		ContextVersion(4, 4))
}

type ViewState struct {
	rows, cols int
	//program           glad.Program
	//vao, vao_txr      glad.VertexArrayObject
	//vbo_c, vbo_w      glad.VertexBufferObject
	//fbo_grid, fbo_env glad.FramebufferObject
	//count             int
	vertices []float32

	autoGrid, autoLayout *glad.AutoConfig
}

func SetupOGL(rows, cols int, aspectRatio float32) *ViewState {
	gl.ClearColor(0.6, 0.6, 0.6, 1.0)
	gl.ClearColor(0.3, 0.3, 0.3, 1.0)

	// We want to fit the points in 90% of the real estate
	var spaceW float32 = 0.9 * 2.0 * aspectRatio
	var spaceH float32 = 0.9 * 2.0

	const pho = 0.866025404 // sqrt(3/4)

	// Total width is x*(cols-1) + 0.5*x*(rows-1)
	// Total height is pho*x*rows
	// Where x is the distance between centers of hexagons
	// We want to maximize x, so find the smallest x so that
	// spaceW = x * (cols-1) + 0.5*x*(rows-1) = x * (cols - 1 + 0.5 * (rows - 1))
	// spaceH = pho * x * rows
	xw := spaceW / (float32(cols-1) + 0.5*float32(rows-1))
	xh := spaceH / (pho * float32(rows))

	var side float32 = xw
	if xh < side {
		side = xh
	}

	// Compute actual size of grid
	totH := float32(rows-1) * pho * side
	totW := float32(cols-1)*side + 0.5*side*float32(rows-1)

	var bx float32 = -aspectRatio + (2.0*aspectRatio-totW)*0.5
	var by float32 = -1.0 + (2.0-totH)*0.5

	vertexShaderSource := LoadFile("./shader_hex.vert")
	fragmentShaderSource := LoadFile("./shader_hex.frag")
	geometryShaderSource := LoadFile("./shader_hex.geom",
		"INV_ASPECT_RATIO", 1.0/aspectRatio,
		"HEX_SIDE", side,
		"PHO", pho)

	/*
		vShader := glad.NewShader(vertexShaderSource, gl.VERTEX_SHADER)
		fShader := glad.NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
		gShader := glad.NewShader(geometryShaderSource, gl.GEOMETRY_SHADER)

		program := glad.NewProgram()
		program.AttachShaders(vShader, fShader, gShader)
		program.Link()

		vShader.Delete()
		fShader.Delete()
		gShader.Delete()
	*/

	vertices := make([]float32, rows*cols*2)
	weights := make([]float32, rows*cols*3)
	colors := make([]float32, rows*cols)
	// Fill the vertices of the hex grid centers
	for i, k := 0, 0; i < rows; i++ {
		for j := 0; j < cols; j, k = j+1, k+1 {
			vertices[2*k+0] = bx + float32(i%2)*side*0.5 + float32(j+i/2)*side
			vertices[2*k+1] = by + float32(i)*pho*side
			colors[k] = rand.Float32()
			weights[3*k+0] = rand.Float32()
			weights[3*k+1] = rand.Float32()
			weights[3*k+2] = rand.Float32()
		}
	}

	autoGrid := glad.AutoBuild(&glad.Config{
		Shaders: []glad.Shader{
			glad.NewShader(vertexShaderSource, gl.VERTEX_SHADER),
			glad.NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER),
			glad.NewShader(geometryShaderSource, gl.GEOMETRY_SHADER),
		},
		Attributes: []glad.Attr{{0, "vert", 2}, {1, "color", 1}, {2, "weights", 3}},
		Data:       [][]float32{vertices, colors, weights},
		DataUsages: []uint32{gl.STATIC_DRAW, gl.DYNAMIC_DRAW, gl.DYNAMIC_DRAW},
		Primitives: gl.POINTS,
		Offscreen:  &glad.Rect{0, 0, 800, 600},
		ClearColor: []float32{0.6, 0.6, 0.6, 1.0},
	})

	passThruVS := `#version 450 core
in vec2 pos;
in vec2 uv;
out vec2 vUV;
void main() { gl_Position = vec4(pos, 0.0, 1.0); vUV = uv; }
`

	passThruFS := `#version 450 core
in vec2 vUV;
out vec4 color;
uniform sampler2D sampler;
void main() { color = texture(sampler, vUV); }
`
	autoLayout := glad.AutoBuild(&glad.Config{
		Shaders: []glad.Shader{
			glad.NewShader(passThruVS, gl.VERTEX_SHADER),
			glad.NewShader(passThruFS, gl.FRAGMENT_SHADER),
		},
		Attributes: []glad.Attr{{0, "pos", 2}, {0, "uv", 2}},
		Data: [][]float32{
			[]float32{
				-1, -1, 0.0, 0.0,
				-1, 1, 0.0, 1.0,
				1, -1, 1.0, 0.0,
				1, 1, 1.0, 1.0,
			},
		},
		DataUsages: []uint32{gl.STATIC_DRAW},
		Primitives: gl.TRIANGLE_STRIP,
		Textures:   []glad.Texture{autoGrid.BgTxr},
	})
	//, 0.3, 0.3, 0.3, 1.0}

	/*
		var (
			bindVbo   uint32 = 0
			bindVbo_c uint32 = 1
			bindVbo_w uint32 = 2
		)

		vao := glad.NewVertexArrayObject()
		vao_txr := glad.NewVertexArrayObject()

		vbo := glad.NewVertexBufferObject()
		vbo.BufferData32(vertices, gl.STATIC_DRAW)
		vao.VertexBuffer32(bindVbo, vbo, 0, 2)

		vbo_c := glad.NewVertexBufferObject()
		vbo_c.BufferData32(colors, gl.DYNAMIC_DRAW)
		vao.VertexBuffer32(bindVbo_c, vbo_c, 0, 1)

		vbo_w := glad.NewVertexBufferObject()
		vbo_w.BufferData32(weights, gl.DYNAMIC_DRAW)
		vao.VertexBuffer32(bindVbo_w, vbo_w, 0, 3)

		vertAttr := program.GetAttributeLocation("vert")
		vao.AttribFormat32(vertAttr, 2, 0)
		vao.AttribBinding(bindVbo, vertAttr)
		vao.EnableAttrib(vertAttr)

		colAttr := program.GetAttributeLocation("color")
		vao.AttribFormat32(colAttr, 1, 0)
		vao.AttribBinding(bindVbo_c, colAttr)
		vao.EnableAttrib(colAttr)

		weiAttr := program.GetAttributeLocation("weights")
		vao.AttribFormat32(weiAttr, 3, 0)
		vao.AttribBinding(bindVbo_w, weiAttr)
		vao.EnableAttrib(weiAttr)

		// Create FBOs and textures for rendering
		fbo_grid := glad.NewFramebuffer()
		txr_grid := glad.NewTexture()
		txr_grid.Storage2D(500, 500)
		txr_grid.SetFilters(gl.NEAREST, gl.NEAREST)
		fbo_grid.Texture(gl.COLOR_ATTACHMENT0, txr_grid)

		fbo_env := glad.NewFramebuffer()
		txr_env := glad.NewTexture()
		txr_env.Storage2D(500, 500)
		txr_env.SetFilters(gl.NEAREST, gl.NEAREST)
		fbo_env.Texture(gl.COLOR_ATTACHMENT0, txr_env)
	*/

	return &ViewState{
		rows,
		cols,
		//program,
		//vao,
		//vbo_c,
		//vbo_w,
		//fbo_grid,
		//fbo_env,
		//txr_grid,
		//txr_env,
		//rows * cols,
		vertices,
		autoGrid,
		autoLayout,
	}
}

func (vs *ViewState) SetColors(colors []float32) {
	vs.autoGrid.VBOs[1].BufferSubData32(colors, 0)
	//vs.vbo_c.BufferSubData32(colors, 0)
}

func (vs *ViewState) SetWeights(weights []float32) {
	//vs.vbo_w.BufferSubData32(weights, 0)
	vs.autoGrid.VBOs[2].BufferSubData32(weights, 0)
}

func (vs *ViewState) DrawFrame() {
	//bgCol := []float32{0.6, 0.6, 0.6, 1.0, 0.3, 0.3, 0.3, 1.0}

	vs.autoGrid.AutoDraw()
	vs.autoLayout.AutoDraw()

	/*
		vs.program.Use()

		vs.fbo_grid.Bind()
		gl.ClearBufferfv(gl.COLOR, 0, &bgCol[0])
		vs.vao.Bind()
		gl.DrawArrays(gl.POINTS, 0, int32(vs.count))
		vs.vao.Unbind()
		vs.fbo_grid.Unbind()

		vs.fbo_env.Bind()
		gl.ClearBufferfv(gl.COLOR, 0, &bgCol[0])
		vs.vao.Bind()
		gl.DrawArrays(gl.POINTS, 0, int32(vs.count))
		vs.vao.Unbind()
		vs.fbo_env.Unbind()

		// Draw quads with textures
		vs.vao_txr.Bind()
		vs.program_txr.Use()
		gl.ClearBufferfv(gl.COLOR, 0, &bgCol[4])
	*/
}

func (vs *ViewState) NearestVertex(x, y float32) (int, int) {
	mx, my := -1, -1
	var minDist float32
	for i, k := 0, 0; i < vs.rows; i++ {
		for j := 0; j < vs.cols; j, k = j+1, k+1 {
			pt := vs.vertices[2*k : 2*(k+1)] // Point to compare to xy
			dist := (pt[0]-x)*(pt[0]-x) + (pt[1]-y)*(pt[1]-y)
			if mx < 0 || dist < minDist {
				minDist = dist
				mx = j
				my = i
			}
		}
	}
	return mx, my
}
