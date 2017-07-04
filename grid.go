package main

import "math/rand"

// The weight from (x,y) to (x+x_, y+y_) is in position (x+wx_ y+wy_, ww_)
var nbors = [6]struct{ x_, y_, wx_, wy_, ww_ int }{
	{-1, 1, 0, 0, 0},
	{0, 1, 0, 0, 1},
	{1, 0, 0, 0, 2},
	{-1, 0, -1, 0, 2},
	{0, -1, 0, -1, 1},
	{1, -1, 1, -1, 0},
}

type bind struct {
	x, y, i int
	data    []float32
}

func (b *bind) next() float32 {
	v := b.data[b.i]
	b.i++
	if b.i >= len(b.data) {
		b.i = 0
	}
	return v
}

type HexGrid struct {
	W, H  int
	Data  []float32 // Value
	WData []float32 // Weights
	Thres []float32 // Threshold
	xWrap func(int, int) int
	yWrap func(int, int) int

	binds []bind
}

func torus(x, xn int) int {
	return (x%xn + xn) % xn
}

// Every cell has 3 edges: West, North, East (the other 3 are owned by lower cells)
func NewGrid(w, h int) *HexGrid {
	data := make([]float32, w*h)
	wdata := make([]float32, w*h*3)
	tdata := make([]float32, w*h)
	// Weights are initialized to 1
	for i := range data {
		data[i] = rand.Float32() * 0.5
		tdata[i] = rand.Float32()
	}
	for i := range wdata {
		wdata[i] = rand.Float32()
	}
	return &HexGrid{
		w,
		h,
		data,
		wdata,
		tdata,
		torus,
		torus,
		make([]bind, 0),
	}
}

func (hg *HexGrid) wrap(x, y int) (x_, y_ int) {
	x_ = hg.xWrap(x, hg.W)
	y_ = hg.yWrap(y, hg.H)
	return
}

func (hg *HexGrid) Get(x, y int) float32 {
	x, y = hg.wrap(x, y)
	return hg.Data[y*hg.W+x]
}

func (hg *HexGrid) Set(x, y int, v float32) {
	x, y = hg.wrap(x, y)
	hg.Data[y*hg.W+x] = v
}

func (hg *HexGrid) GetW(x, y, w int) float32 {
	x, y = hg.wrap(x, y)
	return hg.WData[y*hg.W*3+x*3+w]
}

func (hg *HexGrid) SetW(x, y, w int, v float32) {
	x, y = hg.wrap(x, y)
	hg.WData[y*hg.W*3+x*3+w] = v
}

func (hg *HexGrid) GetT(x, y int) float32 {
	x, y = hg.wrap(x, y)
	return hg.Thres[y*hg.W+x]
}

func (hg *HexGrid) SetT(x, y int, v float32) {
	x, y = hg.wrap(x, y)
	hg.Thres[y*hg.W+x] = v
}

// This will read the values from vals every time an update
// is performed, and will automatically set the value of the
// cell (x,y) to that value after the update
func (hg *HexGrid) Bind(x, y int, vals []float32) {
	b := bind{x, y, 0, vals}
	found := false
	for i := range hg.binds {
		if hg.binds[i].x == x && hg.binds[i].y == y {
			hg.binds[i] = b // Overwrite existing bindings
			found = true
			break
		}
	}
	if !found {
		hg.binds = append(hg.binds, b)
	}
	hg.Set(x, y, vals[0]) // Set initial value
}

// Returns the weights for this edge
func (hg *HexGrid) ContactWeights(x, y int) [6]float32 {
	return [6]float32{
		hg.GetW(x, y, 0),
		hg.GetW(x, y, 1),
		hg.GetW(x, y, 2),
		hg.GetW(x-1, y, 2),
		hg.GetW(x, y-1, 1),
		hg.GetW(x+1, y-1, 0),
	}
}

// Computes the output value of a cell, summing its weighted inputs
func (hg *HexGrid) Activation(x, y int) float32 {
	var act float32
	for _, n := range nbors {
		act += hg.Get(x+n.x_, y+n.y_) * hg.GetW(x+n.wx_, y+n.wy_, n.ww_)
	}
	// Binary activation
	if act < hg.GetT(x, y) {
		return 0
	} else {
		return 1
	}
	//return act
}

/*
	C'é una soglia di attivazione
	Se il neurone ha un valore sopra quella soglia, in update decade di un certo fattore
	Se il neurone è sotto la soglia, allora controlla i vicini e si attiva

	Se un neurone è sopra una certa soglia e un vicino si attiva in seguito ad update, il peso tra loro aumenta di un tot
	In tutti gli altri casi, i pesi diminuiscono di un piccolo valore
*/
func (hg *HexGrid) Update() {
	// This can be kept to avoid reallocating memory all the times
	val := make([]float32, hg.W*hg.H)
	wei := make([]float32, hg.W*hg.H*3)
	thr := make([]float32, hg.W*hg.H)

	const (
		ActivationThreshold     = 0.20
		DecayFactor             = 0.75
		WeightIncreaseFactor    = 1.02
		WeightDecreaseFactor    = 0.99
		ThresholdIncreaseFactor = 1.01
		ThresholdDecreaseFactor = 0.99
	)

	// For every position
	for i := 0; i < hg.H; i++ {
		for j := 0; j < hg.W; j++ {
			a := hg.Activation(j, i)
			t := hg.GetT(j, i)
			if hg.Get(j, i) > ActivationThreshold {
				// If greater than the threshold, decay
				val[i*hg.W+j] = hg.Get(j, i) * DecayFactor
				if a == 1 {
					// If activated, we might be too sensible to this stimuli
					// As it seems to trigger me too often, increase threshold
					t *= ThresholdIncreaseFactor
				} else {
					// Decrease threshold
					t *= ThresholdDecreaseFactor
				}
			} else {
				// If less than threshold, compute activation
				val[i*hg.W+j] = a
				// Decrease threshold
				t *= ThresholdDecreaseFactor
			}
			thr[i*hg.W+j] = t
		}
	}
	// Save updated values so we can use Get methods
	hg.Data = val
	// Also save thresholds even if we don't need them
	hg.Thres = thr

	// Apply bound values
	for i := range hg.binds {
		bx, by := hg.binds[i].x, hg.binds[i].y
		nv := hg.binds[i].next()
		hg.Set(bx, by, nv)
	}

	// Weight depends on the newly computed value
	for i := 0; i < hg.H; i++ {
		for j := 0; j < hg.W; j++ {
			for k := 0; k < 3; k++ {
				if hg.Get(j, i) > ActivationThreshold && hg.Get(j+nbors[k].x_, i+nbors[k].y_) > ActivationThreshold {
					nw := hg.GetW(j, i, k) * WeightIncreaseFactor
					if nw > 1.0 {
						nw = 1.0
					}
					wei[i*hg.W*3+j*3+k] = nw
				} else {
					wei[i*hg.W*3+j*3+k] = hg.GetW(j, i, k) * WeightDecreaseFactor
				}
			}
		}
	}
	// Save weight data as well
	hg.WData = wei
}
