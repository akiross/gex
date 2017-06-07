package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.4-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	//"github.com/go-gl/mathgl/mgl32"
)

type HexGrid struct {
	W, H int
	Data []float32
}

func NewGrid(w, h int) *HexGrid {
	return &HexGrid{w, h, make([]float32, w*h)}
}

func (hg *HexGrid) Get(x, y int) float32 {
	return hg.Data[y*hg.W+x]
}

func (hg *HexGrid) Set(x, y int, v float32) {
	hg.Data[y*hg.W+x] = v
}

type WinOption func()

func glfwTF(v bool) int {
	if v {
		return glfw.True
	}
	return glfw.False
}

func Resizable(v bool) WinOption {
	return func() {
		glfw.WindowHint(glfw.Resizable, glfwTF(v))
	}
}

func ContextVersion(maj, min int) WinOption {
	return func() {
		glfw.WindowHint(glfw.ContextVersionMajor, maj)
		glfw.WindowHint(glfw.ContextVersionMinor, min)
	}
}

func ForwardCompatible(v bool) WinOption {
	return func() {
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfwTF(v))
	}
}

func CoreProfile(v bool) WinOption {
	return func() {
		if v {
			glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		} else {
			glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)
		}
	}
}

func Decorated(v bool) WinOption {
	return func() {
		glfw.WindowHint(glfw.Decorated, glfwTF(v))
	}
}

func NewOGLWindow(width, height int, title string, opts ...WinOption) *glfw.Window {
	// Initialize OpenGL
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize GLFW", err)
	}
	for _, opt := range opts {
		opt()
	}
	win, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		log.Fatalln("Failed to create window", err)
	}
	win.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL", err)
	}
	return win
}

type Shader uint32

func NewShader(source string, shaderType uint32) Shader {
	if source == "" {
		log.Fatalln("Unable to create shader from empty string")
	}
	var sh Shader
	sh = Shader(gl.CreateShader(shaderType))
	csrc, free := gl.Strs(source + "\x00")
	gl.ShaderSource(uint32(sh), 1, csrc, nil)
	free()
	gl.CompileShader(uint32(sh))

	if sh.GetParameter(gl.COMPILE_STATUS) == gl.FALSE {
		infoLog := sh.GetInfoLog()
		log.Fatalln("Unable to compile shader:\n", infoLog)
	}
	return sh
}

func (sh Shader) Delete() {
	gl.DeleteShader(uint32(sh))
}

func (sh Shader) GetParameter(pname uint32) int32 {
	var val int32
	gl.GetShaderiv(uint32(sh), pname, &val)
	return val
}

func (sh Shader) GetInfoLog() string {
	logLen := sh.GetParameter(gl.INFO_LOG_LENGTH)
	infoLog := string(make([]byte, int(logLen+1)))
	var savedLen int32
	gl.GetShaderInfoLog(uint32(sh), logLen, &savedLen, gl.Str(infoLog))
	if savedLen+1 != logLen {
		log.Println("Shader Info Log different lengths reported:", logLen, savedLen)
	}
	return infoLog
}

type Program uint32

func NewProgram() Program {
	return Program(gl.CreateProgram())
}

func (pr Program) AttachShaders(shaders ...Shader) {
	for _, sh := range shaders {
		gl.AttachShader(uint32(pr), uint32(sh))
	}
}

func (pr Program) Link() {
	gl.LinkProgram(uint32(pr))

	if pr.GetParameter(gl.LINK_STATUS) == gl.FALSE {
		infoLog := pr.GetInfoLog()
		log.Fatalln("Unable to link program:\n", infoLog)
	}
}

func (pr Program) GetParameter(pname uint32) int32 {
	var val int32
	gl.GetProgramiv(uint32(pr), pname, &val)
	return val
}

func (pr Program) GetInfoLog() string {
	logLen := pr.GetParameter(gl.INFO_LOG_LENGTH)
	infoLog := string(make([]byte, int(logLen+1)))
	var savedLen int32
	gl.GetProgramInfoLog(uint32(pr), logLen, &savedLen, gl.Str(infoLog))
	if savedLen+1 != logLen {
		log.Println("Program Info Log different lengths reported:", logLen, savedLen)
	}
	return infoLog
}

func (pr Program) GetAttributeLocation(name string) VertexAttrib {
	return VertexAttrib(gl.GetAttribLocation(uint32(pr), gl.Str(name+"\x00")))
}

type VertexAttrib uint32

func (va VertexAttrib) Enable() {
	gl.EnableVertexAttribArray(uint32(va))
}

// size: number of components per vertex (e.g. 3D vertices -> 3)
// dataType: gl.FLOAT, etc
// normalized: define if data have to be normalized
// stride: bytes between two vertices, 0 means they are tightly packed
// offset: bytes of offset to the first element in the array
func (va VertexAttrib) Pointer(size int32, dataType uint32, normalize bool, stride, offset int32) {
	gl.VertexAttribPointer(uint32(va), size, dataType, normalize, stride, gl.PtrOffset(int(offset)))
}

var (
	vertexShaderSource = `#version 410
in vec3 vert; // Input center position for this hexagon
in vec3 color; // Input color for this hexagon
out vec3 vColor; // Color to be forwarded to geometry shader
void main() { gl_Position = vec4(vert, 1); vColor = color; }
`

	fragmentShaderSource = `#version 410
in vec3 gColor; // Color from geometry shader
out vec4 oColor; // Color of fragment
void main() { oColor = vec4(gColor, 1.0); }
`

	geometryShaderSource = `#version 410 core
layout (points) in;
layout (triangle_strip, max_vertices = 10) out;
in vec3 vColor[]; // Color of each input vertex (just 1)
out vec3 gColor; // Color for output primitives

const float PI = 3.14159265;
const float r = 0.1;

// Compute position of i-th vertex using sine and cosine
void pos(int i) {
	float a = PI * (0.5 + i / 3.0);
	vec2 offs = vec2(r * cos(a), r * sin(a));
	gl_Position = gl_in[0].gl_Position + vec4(offs, 0.0, 0.0);
	EmitVertex();
}

// Produce two triangle strips using vertices (top vertex is number 0, CCW)
void main() {
	gColor = vColor[0];
	pos(1);
	pos(0);
	gl_Position = gl_in[0].gl_Position;
	pos(5);
	pos(4);
	pos(1);
	pos(2);
	gl_Position = gl_in[0].gl_Position;
	pos(3);
	pos(4);
	EndPrimitive();
}
`
)

func main() {
	// OpenGL context is bound to a CPU thread
	runtime.LockOSThread()
	// Create a window for OpenGL
	win := NewOGLWindow(800, 600, "GEX",
		CoreProfile(true),
		ForwardCompatible(true),
		Resizable(false),
		ContextVersion(4, 4))
	defer glfw.Terminate()

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("Starting GEX! OpenGL version", version)

	gl.ClearColor(0.6, 0.6, 0.6, 1.0)

	vShader := NewShader(vertexShaderSource, gl.VERTEX_SHADER)
	fShader := NewShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	gShader := NewShader(geometryShaderSource, gl.GEOMETRY_SHADER)

	program := NewProgram()
	program.AttachShaders(vShader, fShader, gShader)
	program.Link()

	vShader.Delete()
	fShader.Delete()
	gShader.Delete()

	const (
		rows = 5
		cols = 5
	)

	// Since kernel is compiled at runtime, we can put constants in it

	var hw, hh float32 = 1.0 / cols, 1.0 / rows
	var bx, by float32 = -0.5, -0.5
	//hexGrid := NewGrid(cols, rows)
	//asd
	vertices := make([]float32, rows*cols*2)
	colors := make([]float32, rows*cols*3)
	// Fill the vertices of the hex grid centers
	for i, k := 0, 0; i < rows; i++ {
		for j := 0; j < cols; j, k = j+1, k+1 {
			if i%2 == 0 {
				vertices[2*k+0] = bx + 0.5*hw + float32(j+i/2)*hw
				vertices[2*k+1] = by + hh*2/4 + float32(i/2)*hh*3/2
			} else {
				vertices[2*k+0] = bx + hw + float32(j+i/2)*hw
				vertices[2*k+1] = by + hh*5/4 + float32(i/2)*hh*3/2
			}
			colors[3*k+0] = rand.Float32()
			colors[3*k+1] = rand.Float32()
			colors[3*k+2] = rand.Float32()
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
	colAttr.Pointer(3, gl.FLOAT, false, 0, 0)
	colAttr.Enable()

	// glBufferSubData can be used to change the colors. Create a secondary buffer, bind it to an attribute, then update color as needed

	start := time.Now()

	// Main loop
	for !win.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(uint32(program))
		gl.BindVertexArray(vao)

		gl.DrawArrays(gl.POINTS, 0, rows*cols)

		win.SwapBuffers()
		glfw.PollEvents()

		if time.Since(start) > time.Second {
			// Replace color data
			for i, k := 0, 0; i < rows; i++ {
				for j := 0; j < cols; j, k = j+1, k+1 {
					colors[3*k+0] = rand.Float32()
					colors[3*k+1] = rand.Float32()
					colors[3*k+2] = rand.Float32()
				}
			}
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo_c)
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(colors)*4, gl.Ptr(colors))
			// Save new time
			start = time.Now()
		}

	}
}
