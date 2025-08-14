package fluid

import "math"

type Fluid struct {
	d, vx, vy [][]float64
	k         float64
	iters     int
}

func newSlice2D(w, h int) [][]float64 {
	s := make([][]float64, w)
	for i := range s {
		s[i] = make([]float64, h)
	}
	return s
}

func NewFluid(w, h int, k float64, iters int) *Fluid {
	d := newSlice2D(w, h)
	vx := newSlice2D(w, h)
	vy := newSlice2D(w, h)
	return &Fluid{d: d, vx: vx, vy: vy, k: k, iters: iters}
}

func val(a [][]float64, x, y int) float64 {
	if x < 0 || y < 0 || x >= len(a) || y >= len(a[0]) {
		return 0
	}
	return a[x][y]
}

func (f *Fluid) D(x, y int) float64 {
	return val(f.d, x, y)
}

func (f *Fluid) Set(x, y int, d, vx, vy float64) {
	if x < 0 || y < 0 || x >= len(f.d) || y >= len(f.d[0]) {
		return
	}
	f.d[x][y] = d
	f.vx[x][y] = vx
	f.vy[x][y] = vy
}

func diffuse(x [][]float64, k float64, iters int) [][]float64 {
	xn := newSlice2D(len(x), len(x[0]))
	for range iters {
		for c := range xn {
			for r := range xn[c] {
				sn := (val(xn, c-1, r) + val(xn, c+1, r) + val(xn, c, r-1) + val(xn, c, r+1)) / 4.0
				xn[c][r] = (x[c][r] + k*sn) / (1 + k)
			}
		}
	}
	return xn
}

func (f *Fluid) advect() [][]float64 {
	lerp := func(a, b, k float64) float64 {
		return a + k*(b-a)
	}
	d := func(x, y int) float64 {
		return f.D(x, y)
	}
	dn := newSlice2D(len(f.d), len(f.d[0]))
	for x := range f.d {
		for y := range f.d[x] {
			fx, fy := float64(x)-f.vx[x][y], float64(y)-f.vy[x][y]
			ix, iy := int(fx), int(fy)
			jx, jy := fx-float64(ix), fy-float64(iy)
			z1 := lerp(d(ix, iy), d(ix+1, iy), jx)
			z2 := lerp(d(ix, iy+1), d(ix+1, iy+1), jx)
			dn[x][y] = math.Max(lerp(z1, z2, jy), 0)
		}
	}
	return dn
}

func (f *Fluid) clearDivergence() {
	dv := newSlice2D(len(f.d), len(f.d[0]))
	for x := range f.d {
		for y := range f.d[x] {
			dv[x][y] = (val(f.vx, x+1, y) - val(f.vx, x-1, y) + val(f.vy, x, y+1) - val(f.vy, x, y-1)) / 2.0
		}
	}
	p := newSlice2D(len(f.d), len(f.d[0]))
	for range f.iters {
		for x := range f.d {
			for y := range f.d[x] {
				p[x][y] = (val(p, x-1, y) + val(p, x+1, y) + val(p, x, y-1) + val(p, x, y+1) - dv[x][y]) / 4.0
			}
		}
	}
	for x := range f.d {
		for y := range f.d[x] {
			f.vx[x][y] -= (val(p, x+1, y) - val(p, x-1, y)) / 2.0
			f.vy[x][y] -= (val(p, x, y+1) - val(p, x, y-1)) / 2.0
		}
	}
}

func (f *Fluid) Update() {
	// TODO: fluid fade out
	f.d = diffuse(f.d, f.k, f.iters)
	f.vx = diffuse(f.vx, f.k, f.iters)
	f.vy = diffuse(f.vy, f.k, f.iters)
	f.d = f.advect()
	f.clearDivergence()
}

func (f *Fluid) Reset() {
	for x := range f.d {
		for y := range f.d[x] {
			f.d[x][y] = 0
			f.vx[x][y] = 0
			f.vy[x][y] = 0
		}
	}
}

func resizeSlice2D(s [][]float64, w, h int) [][]float64 {
	n := newSlice2D(w, h)
	for x := 0; x < len(s) && x < w; x++ {
		for y := 0; y < len(s[0]) && y < h; y++ {
			n[x][y] = s[x][y]
		}
	}
	return n
}

func (f *Fluid) Resize(w, h int) {
	f.d = resizeSlice2D(f.d, w, h)
	f.vx = resizeSlice2D(f.vx, w, h)
	f.vy = resizeSlice2D(f.vy, w, h)
}
