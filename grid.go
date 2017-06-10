package main

type HexGrid struct {
	W, H  int
	Data  []float32
	WData []float32
	xWrap func(int, int) int
	yWrap func(int, int) int
}

func torus(x, xn int) int {
	return (x%xn + xn) % xn
}

// Every cell has 3 edges: West, North, East (the other 3 are owned by lower cells)
func NewGrid(w, h int) *HexGrid {
	wdata := make([]float32, w*h*3)
	// Weights are initialized to 1
	for i := range wdata {
		wdata[i] = 1.0
	}
	return &HexGrid{
		w,
		h,
		make([]float32, w*h),
		wdata,
		torus,
		torus,
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
	nbors := [6]struct{ x_, y_, wx_, wy_, ww_ int }{
		{-1, 1, 0, 0, 0},
		{0, 1, 0, 0, 1},
		{1, 0, 0, 0, 2},
		{-1, 0, -1, 0, 2},
		{0, -1, 0, -1, 1},
		{1, -1, 1, -1, 0},
	}
	var act float32
	for _, n := range nbors {
		act += hg.Get(x+n.x_, y+n.y_) * hg.GetW(x+n.wx_, y+n.wy_, n.ww_)
	}
	// Binary activation
	if act < 0.5 {
		return 0
	} else {
		return 1
	}
	//return act
}

func (hg *HexGrid) Update() {
	// This can be kept to avoid reallocating memory all the times
	res := make([]float32, hg.W*hg.H)
	// For every position
	for i := 0; i < hg.H; i++ {
		for j := 0; j < hg.W; j++ {
			res[i*hg.W+j] = hg.Activation(j, i)
		}
	}
	hg.Data = res
}
