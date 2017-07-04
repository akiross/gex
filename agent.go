package main

// An agent in a space

import "math"

type Agent struct {
	X, Y float32 // Where is the agent located
	R    float32 // Current rotation
}

// Change the state of the agent by rotating it of some amount
func (ag *Agent) Rotate(v float32) {
	ag.R += v
}

// Advance the agent by moving in the direction of rotation
func (ag *Agent) Advance(v float32) {
	// Compute direction vector
	ag.X += v * float32(math.Cos(float64(ag.R)))
	ag.Y += v * float32(math.Sin(float64(ag.R)))
}
