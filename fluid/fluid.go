package fluid

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

func (f *Fluid) D(x, y int) float64 {
	if x < 0 || y < 0 || x >= len(f.d) || y >= len(f.d[0]) {
		return 0
	}
	return f.d[x][y]
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
				sn := 0.0
				if c > 0 {
					sn += xn[c-1][r]
				}
				if c < len(xn)-1 {
					sn += xn[c+1][r]
				}
				if r > 0 {
					sn += xn[c][r-1]
				}
				if r < len(xn[c])-1 {
					sn += xn[c][r+1]
				}
				sn /= 4.0
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
			dn[x][y] = lerp(z1, z2, jy)
		}
	}
	return dn
}

func (f *Fluid) clearDivergence() {
	dv := newSlice2D(len(f.d), len(f.d[0]))
	for x := 1; x < len(f.d)-1; x++ {
		for y := 1; y < len(f.d[0])-1; y++ {
			dv[x][y] = (f.vx[x+1][y] - f.vx[x-1][y] + f.vy[x][y+1] - f.vy[x][y-1]) / 2.0
		}
	}
	p := newSlice2D(len(f.d), len(f.d[0]))
	for range f.iters {
		for x := 1; x < len(f.d)-1; x++ {
			for y := 1; y < len(f.d[0])-1; y++ {
				p[x][y] = (p[x-1][y] + p[x+1][y] + p[x][y-1] + p[x][y+1] - dv[x][y]) / 4.0
			}
		}
	}
	for x := 1; x < len(f.d)-1; x++ {
		for y := 1; y < len(f.d[0])-1; y++ {
			f.vx[x][y] -= (p[x+1][y] - p[x-1][y]) / 2.0
			f.vy[x][y] -= (p[x][y+1] - p[x][y-1]) / 2.0
		}
	}
}

func (f *Fluid) Update() {
	f.d = diffuse(f.d, f.k, f.iters)
	f.vx = diffuse(f.vx, f.k, f.iters)
	f.vy = diffuse(f.vy, f.k, f.iters)
	f.d = f.advect()
	f.clearDivergence()
}
