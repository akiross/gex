package main

import (
	"log"

	"github.com/go-gl/gl/v4.4-core/gl"
)

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
