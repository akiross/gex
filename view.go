package main

import (
	"github.com/go-gl/gl/v4.4-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"io/ioutil"
	"math/rand"
)

func LoadFile(path string) string {
	cont, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(cont)
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
	program    Program
	vao, vbo_c uint32
	count      int
	vertices   []float32
}

func SetupOGL(rows, cols int) *ViewState {
	gl.ClearColor(0.6, 0.6, 0.6, 1.0)

	vertexShaderSource := LoadFile("./shader_hex.vert")
	fragmentShaderSource := LoadFile("./shader_hex.frag")
	geometryShaderSource := LoadFile("./shader_hex.geom")

	vShader := NewShader(vertexShaderSource, gl.VERTEX_SHADER)
	fShader := NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	gShader := NewShader(geometryShaderSource, gl.GEOMETRY_SHADER)

	program := NewProgram()
	program.AttachShaders(vShader, fShader, gShader)
	program.Link()

	vShader.Delete()
	fShader.Delete()
	gShader.Delete()

	var side float32 = 1.5 / float32(cols)
	//var hw, hh float32 = 1.5 / float32(cols), 1.5 / float32(rows)
	var bx, by float32 = -0.9, -0.9

	const pho = 0.866025404 // sqrt(3/4)

	vertices := make([]float32, rows*cols*2)
	colors := make([]float32, rows*cols)
	// Fill the vertices of the hex grid centers
	for i, k := 0, 0; i < rows; i++ {
		for j := 0; j < cols; j, k = j+1, k+1 {
			vertices[2*k+0] = bx + float32(i%2)*side*0.5 + float32(j+i/2)*side
			vertices[2*k+1] = by + float32(i)*pho*side
			colors[k] = rand.Float32()
		}
	}

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	vertAttr := program.GetAttributeLocation("vert")
	vertAttr.Pointer(2, gl.FLOAT, false, 0, 0)
	vertAttr.Enable()

	var vbo_c uint32
	gl.GenBuffers(1, &vbo_c)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo_c)
	gl.BufferData(gl.ARRAY_BUFFER, len(colors)*4, gl.Ptr(colors), gl.DYNAMIC_DRAW)

	colAttr := program.GetAttributeLocation("color")
	colAttr.Pointer(1, gl.FLOAT, false, 0, 0)
	colAttr.Enable()

	return &ViewState{rows, cols, program, vao, vbo_c, rows * cols, vertices}
}

func (vs *ViewState) SetColors(colors []float32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vs.vbo_c)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(colors)*4, gl.Ptr(colors))
}

func (vs *ViewState) DrawFrame() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(uint32(vs.program))
	gl.BindVertexArray(vs.vao)
	gl.DrawArrays(gl.POINTS, 0, int32(vs.count))
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
