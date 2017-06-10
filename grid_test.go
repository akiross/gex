package main

import "testing"

func TestGettersSetters(t *testing.T) {
	g := NewGrid(3, 4)

	cases := []struct {
		sx, sy int
		gx, gy int
		v      float32
	}{
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 2},
		{-1, -1, 2, 3, 3},
		{3, 4, 0, 0, 4},
		{-4, -5, 2, 3, 5},
		{6, 8, 0, 0, 6},
	}

	for i, c := range cases {
		g.Set(c.sx, c.sy, c.v)
		if g.Get(c.gx, c.gy) != c.v {
			t.Error("Wrong Set/Get", i, c)
		}
	}
	// TODO fare lo stesso con i pesi
}
