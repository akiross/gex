package main

import (
	"flag"
	glad "github.com/akiross/go-glad"
	"github.com/go-gl/glfw/v3.2/glfw"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	rows = flag.Int("rows", 20, "Number of rows in the grid")
	cols = flag.Int("cols", 20, "NUmber of columns in the grid")

	width  = flag.Int("width", 1000, "Window width")
	height = flag.Int("height", 600, "Window height")

	inPath = flag.String("input_data", "", "File containing input data (space separated floats)")
)

var (
	lastHitX       float32
	lastHitY       float32
	lastHit        bool
	bindInput      bool
	updateRequest  bool
	updateInterval time.Duration = 2 * time.Second
)

func myMouse(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if action == glfw.Press {
		x, y := w.GetCursorPos()
		w, h := w.GetSize()
		rx, ry := 2.0*x/float64(w)-1.0, 2.0*(float64(h)-y)/float64(h)-1.0
		lastHitX = float32(rx) * float32(*width) / float32(*height)
		lastHitY = float32(ry)
		lastHit = true
		if button == glfw.MouseButtonLeft {
			log.Println("Click in position", lastHitX, lastHitY)
		} else if button == glfw.MouseButtonRight {
			log.Println("Binding input data to", lastHitX, lastHitY)
			bindInput = true
		}
	}
}

func myKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && key == glfw.KeySpace {
		updateRequest = true
	}
	if action == glfw.Press && key == glfw.KeyQ {
		updateInterval /= 2
		log.Println("Update interval changed to", updateInterval)
	}
	if action == glfw.Press && key == glfw.KeyA {
		updateInterval *= 2
		log.Println("Update interval changed to", updateInterval)
	}
}

func readFloats(path string) []float32 {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fields := strings.Fields(string(bytes))
	data := make([]float32, len(fields))
	for i, s := range fields {
		val, err := strconv.ParseFloat(s, 32)
		if err != nil {
			log.Println("Conversion impossible for datum", s)
		} else {
			data[i] = float32(val)
		}
	}
	return data
}

func main() {
	// OpenGL context is bound to a CPU thread
	runtime.LockOSThread()

	flag.Parse()

	// Load data file
	var inData []float32
	if *inPath != "" {
		inData = readFloats(*inPath)
	}

	// Create a window for OpenGL
	win := glad.NewOGLWindow(int(*width), int(*height), "Gex",
		glad.CoreProfile(true),
		glad.Resizable(false),
		glad.ContextVersion(4, 4),
		//glad.VSync(true),
	)
	defer glad.Terminate()
	// Enable VSync
	glad.SwapInterval(1)

	win.SetMouseButtonCallback(myMouse)
	win.SetKeyCallback(myKey)

	state := SetupOGL(*rows, *cols, float32(*width)/float32(*height))

	grid := NewGrid(*cols, *rows)

	start := time.Now()

	state.SetColors(grid.Data)
	state.SetWeights(grid.WData)

	// Main loop
	for !win.ShouldClose() {
		// Check if there were mouse click
		if lastHit {
			lastHit = false // Reset state
			nx, ny := state.NearestVertex(lastHitX, lastHitY)
			if bindInput {
				grid.Bind(nx, ny, inData)
				bindInput = false
			} else {
				// Get nearest vertex of the grid
				grid.Set(nx, ny, 1.0) //-grid.Get(nx, ny)) // Toggle value
				//grid.SetW(nx, ny, 0, 0.5)  //1.0-grid.Get(nx, ny))
				//grid.SetW(nx, ny, 1, 0.75) //1.0-grid.Get(nx, ny))
				//grid.SetW(nx, ny, 2, 1.0)  //-grid.Get(nx, ny))
				//state.SetWeights(grid.WData)
			}
			state.SetColors(grid.Data)
		}

		state.DrawFrame()

		win.SwapBuffers()
		glfw.PollEvents()

		if updateRequest {
			// Update world step
			grid.Update()
			state.SetColors(grid.Data)
			state.SetWeights(grid.WData)
			updateRequest = false
		}

		if time.Since(start) > updateInterval {
			// Save new time
			start = time.Now()
			updateRequest = true
		}

	}
}

/*
func mainard() {
	runtime.LockOSThread()

	log.Println("Starting")

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
*/
