package main

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"time"
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
		rx, ry := (4.0/3.0)*2.0*x/float64(w)-1.0, 2.0*(float64(h)-y)/float64(h)-1.0
		last_hit_x = float32(rx)
		last_hit_y = float32(ry)
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

	// Create a window for OpenGL
	win := CreateWindow(800, 600, "Gex")
	defer glfw.Terminate()

	win.SetMouseButtonCallback(myMouse)
	win.SetKeyCallback(myKey)

	const (
		rows = 10
		cols = 10
	)

	state := SetupOGL(rows, cols)

	grid := NewGrid(cols, rows)

	start := time.Now()

	// Main loop
	for !win.ShouldClose() {
		// Check if there were mouse click
		if last_hit {
			last_hit = false // Reset state
			// Get nearest vertex of the grid
			nx, ny := state.NearestVertex(last_hit_x, last_hit_y)
			grid.Set(nx, ny, 1.0-grid.Get(nx, ny)) // Toggle value
			state.SetColors(grid.Data)
		}

		state.DrawFrame()

		win.SwapBuffers()
		glfw.PollEvents()

		if update_request {
			// Update world step
			grid.Update()
			state.SetColors(grid.Data)
			update_request = false
		}

		if time.Since(start) > time.Second {
			// Save new time
			start = time.Now()
		}

	}
}
