package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.4-core/gl"
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
	rows, cols        int
	program           Program
	vao, vbo_c, vbo_w uint32
	count             int
	vertices          []float32
}

func SetupOGL(rows, cols int, aspectRatio float32) *ViewState {
	gl.ClearColor(0.6, 0.6, 0.6, 1.0)

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

	vShader := NewShader(vertexShaderSource, gl.VERTEX_SHADER)
	fShader := NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	gShader := NewShader(geometryShaderSource, gl.GEOMETRY_SHADER)

	program := NewProgram()
	program.AttachShaders(vShader, fShader, gShader)
	program.Link()

	vShader.Delete()
	fShader.Delete()
	gShader.Delete()

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

	// Hex colors
	var vbo_c uint32
	gl.GenBuffers(1, &vbo_c)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo_c)
	gl.BufferData(gl.ARRAY_BUFFER, len(colors)*4, gl.Ptr(colors), gl.DYNAMIC_DRAW)

	colAttr := program.GetAttributeLocation("color")
	colAttr.Pointer(1, gl.FLOAT, false, 0, 0)
	colAttr.Enable()

	// Hex weights
	var vbo_w uint32
	gl.GenBuffers(1, &vbo_w)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo_w)
	gl.BufferData(gl.ARRAY_BUFFER, len(weights)*4, gl.Ptr(weights), gl.DYNAMIC_DRAW)

	weiAttr := program.GetAttributeLocation("weights")
	weiAttr.Pointer(3, gl.FLOAT, false, 0, 0)
	weiAttr.Enable()

	return &ViewState{rows, cols, program, vao, vbo_c, vbo_w, rows * cols, vertices}
}

func (vs *ViewState) SetColors(colors []float32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vs.vbo_c)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(colors)*4, gl.Ptr(colors))
}

func (vs *ViewState) SetWeights(weights []float32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vs.vbo_w)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(weights)*4, gl.Ptr(weights))
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
