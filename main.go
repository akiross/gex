package main

import (
	"flag"
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
	"runtime"
	"time"
)

var (
	rows = flag.Int("rows", 20, "Number of rows in the grid")
	cols = flag.Int("cols", 20, "NUmber of columns in the grid")

	width  = flag.Int("width", 1000, "Window width")
	height = flag.Int("height", 600, "Window height")
)

var (
	last_hit_x     float32
	last_hit_y     float32
	last_hit       bool
	update_request bool
)

func myMouse(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if action == glfw.Press && button == glfw.MouseButtonLeft {
		x, y := w.GetCursorPos()
		w, h := w.GetSize()
		rx, ry := 2.0*x/float64(w)-1.0, 2.0*(float64(h)-y)/float64(h)-1.0
		last_hit_x = float32(rx) * float32(*width) / float32(*height)
		last_hit_y = float32(ry)
		log.Println("Cliclato in position", last_hit_x, last_hit_y)
		last_hit = true
	}
}

func myKey(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press && key == glfw.KeySpace {
		update_request = true
	}
}

func main() {
	// OpenGL context is bound to a CPU thread
	runtime.LockOSThread()

	flag.Parse()

	// Create a window for OpenGL
	win := CreateWindow(int(*width), int(*height), "Gex")
	defer glfw.Terminate()

	win.SetMouseButtonCallback(myMouse)
	win.SetKeyCallback(myKey)

	state := SetupOGL(*rows, *cols, float32(*width)/float32(*height))

	grid := NewGrid(*cols, *rows)

	start := time.Now()

	// Main loop
	for !win.ShouldClose() {
		// Check if there were mouse click
		if last_hit {
			last_hit = false // Reset state
			// Get nearest vertex of the grid
			nx, ny := state.NearestVertex(last_hit_x, last_hit_y)
			log.Println("Clickato era con valore", grid.Get(nx, ny))
			grid.Set(nx, ny, 1.0-grid.Get(nx, ny)) // Toggle value
			state.SetColors(grid.Data)
			grid.SetW(nx, ny, 0, 0.5)  //1.0-grid.Get(nx, ny))
			grid.SetW(nx, ny, 1, 0.75) //1.0-grid.Get(nx, ny))
			grid.SetW(nx, ny, 2, 1.0)  //-grid.Get(nx, ny))
			state.SetWeights(grid.WData)
		}

		state.DrawFrame()

		win.SwapBuffers()
		glfw.PollEvents()

		if update_request {
			// Update world step
			grid.Update()
			state.SetColors(grid.Data)
			state.SetWeights(grid.WData)
			update_request = false
		}

		if time.Since(start) > time.Second {
			// Save new time
			start = time.Now()
		}

	}
}
