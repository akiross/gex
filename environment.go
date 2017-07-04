package main

// The environment where one or more agents will move and act
// It's core is simple wave diffusion

import (
	glad "github.com/akiross/go-glad"
	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"image"
	"image/color"
	"log"
	"runtime"
)

const (
	WIDTH  = 128
	HEIGHT = 128
	DAMP   = 0.95
)

type Environment struct {
	w, h      int
	materials []int
	grids     [][]float32
	active    int
}

func NewEnvironment(w, h int) *Environment {
	g := Environment{w: w, h: h}
	g.materials = make([]int, w*h)
	g.grids = make([][]float32, 2)
	g.grids[0] = make([]float32, w*h)
	g.grids[1] = make([]float32, w*h)

	for i := 0; i < w; i++ {
		g.materials[i] = 1
		g.materials[(h-1)*w+i] = 1
	}
	for i := 0; i < h; i++ {
		g.materials[i*w] = 1
		g.materials[(i+1)*w-1] = 1
	}

	for i := 0; i < 32; i++ {
		g.materials[16*w+w*i+32] = 1
	}

	return &g
}

func (g *Environment) Get(x, y int) float32 {
	return g.grids[0][y*g.w+x]
}

func (g *Environment) Set(x, y int, v float32) {
	g.grids[0][y*g.w+x] = v
}

func (g *Environment) Update(damp float32) {
	cgrid := g.grids[0] // Current grid
	pgrid := g.grids[1] // Previous grid
	for y := 1; y < g.h-1; y++ {
		for x := 1; x < g.w-1; x++ {
			switch g.materials[y*g.w+x] {
			case 0:
				var (
					old    = pgrid[y*g.w+x]
					left   = cgrid[y*g.w+x-1]
					right  = cgrid[y*g.w+x+1]
					top    = cgrid[(y+1)*g.w+x]
					bottom = cgrid[(y-1)*g.w+x]
				)
				pgrid[y*g.w+x] = damp * ((left+right+top+bottom)*0.5 - old)
			case 1:
			}
		}
	}

	g.grids[0], g.grids[1] = g.grids[1], g.grids[0]
}

// When clicked on window, set a value on the grid
func makeClicker(g *Environment) func(*glfw.Window, glfw.MouseButton, glfw.Action, glfw.ModifierKey) {
	return func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		if action == glfw.Press {
			cx, cy := w.GetCursorPos()
			ww, wh := w.GetSize()
			x, y := float64(cx)/float64(ww), float64(cy)/float64(wh)
			px, py := int(x*WIDTH), int(HEIGHT-y*HEIGHT)
			g.Set(px, py, 20.0)
		}
	}
}

func mainard() {
	runtime.LockOSThread()

	log.Println("Starting")

	win := glad.NewOGLWindow(512, 512, "Waves",
		glad.CoreProfile(true),
		glad.Resizable(false),
		glad.ContextVersion(4, 4),
		//glad.VSync(true),
	)
	defer glad.Terminate()
	// Enable VSync
	glad.SwapInterval(1)

	bgCol := []float32{0.3, 0.3, 0.3, 1.0}

	vertShader := glad.NewShader(vertexShaderSource, gl.VERTEX_SHADER)
	fragShader := glad.NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER)

	program := glad.NewProgram()
	program.AttachShaders(vertShader, fragShader)
	program.Link()

	vertShader.Delete()
	fragShader.Delete()

	// Data to be used when drawing
	// Format: X, Y, U, V
	vertPosAndUV := []float32{
		-1.0, -1.0, 0.0, 0.0,
		-1.0, 1.0, 0.0, 1.0,
		1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
	}

	// Create a texture
	txrImg := image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT))

	updateColors := func(g *Environment, time int) {
		for y := 0; y < HEIGHT; y++ {
			for x := 0; x < WIDTH; x++ {
				switch g.materials[y*WIDTH+x] {
				case 0: // fluid
					val := g.Get(x, y)
					var pcol, ocol, ncol uint8 // Positive color, negative color, overflow color
					if val > 1 {
						ocol = 255
						val -= 1.0
					}
					if val > 0 {
						pcol = uint8(255 * val)
					} else {
						ncol = uint8(-255 * val)
					}
					txrImg.SetRGBA(x, y, color.RGBA{pcol, ocol, ncol, 255}) //color.RGBA{uint8(float32(time%255) * float32(x%8) / 7.0), col * uint8(float32(y%16)/15.0), col, 255})
				case 1: // wall
					txrImg.SetRGBA(x, y, color.RGBA{0, 255, 0, 255})
				}
			}
		}
	}

	grid := NewEnvironment(WIDTH, HEIGHT)

	win.SetMouseButtonCallback(makeClicker(grid))

	var bindPos uint32 = 0
	vao := glad.NewVertexArrayObject()
	vbo := glad.NewVertexBufferObject()
	vbo.BufferData32(vertPosAndUV, gl.STATIC_DRAW)
	vao.VertexBuffer32(bindPos, vbo, 0, 4)

	txr := glad.NewTexture()
	txr.Storage2D(WIDTH, HEIGHT)
	txr.Bind()
	txr.Image2D(txrImg)
	//txr.Clear(255, 0, 0, 255)
	txr.SetFilters(gl.NEAREST, gl.NEAREST)

	attrPos := program.GetAttributeLocation("pos")
	vao.AttribFormat32(attrPos, 2, 0)
	vao.AttribBinding(bindPos, attrPos)

	attrUV := program.GetAttributeLocation("uv")
	vao.AttribFormat32(attrUV, 2, 2)
	vao.AttribBinding(bindPos, attrUV)

	vao.EnableAttrib(attrPos)
	vao.EnableAttrib(attrUV)

	vao.Bind()

	i := 0
	for !win.ShouldClose() {
		gl.ClearBufferfv(gl.COLOR, 0, &bgCol[0])
		gl.Clear(gl.COLOR_BUFFER_BIT)
		program.Use()
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

		i += 1
		grid.Update(DAMP)
		updateColors(grid, i)
		txr.Image2D(txrImg)

		win.SwapBuffers()
		glad.PollEvents()
	}
}

var (
	vertexShaderSource = `#version 440 core
in vec2 pos;
in vec2 uv;
out vec2 vUV;
void main() { gl_Position = vec4(pos, 0.0, 1.0); vUV = uv; }`
	fragmentShaderSource = `#version 440 core
in vec2 vUV;
out vec4 color;
uniform sampler2D sampler;
void main() { color = vec4(0.1, 0.1, 0.1, 1.0) + texture(sampler, vUV); }`
)
