package fluid

type Fluid struct {
	d, vx, vy [][]float64
	k         float64
	iters     int
}

func NewFluid(w, h int, k float64, iters int) *Fluid {
	d := make([][]float64, w)
	vx := make([][]float64, w)
	vy := make([][]float64, w)
	for i := range w {
		d[i] = make([]float64, h)
		vx[i] = make([]float64, h)
		vy[i] = make([]float64, h)
	}
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
	xn := make([][]float64, len(x))
	for i := range xn {
		xn[i] = make([]float64, len(x[0]))
	}
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

// func (f *Fluid) advect() {
// 	lerp := func(a, b, k float64) float64 {
// 		return a + k*(b-a)
// 	}
// 	d := func(x, y int) float64 {
// 		return f.D(x, y)
// 	}
// 	dn := make([][]float64, len(f.cells))
// 	for i := range dn {
// 		dn[i] = make([]float64, len(f.cells[0]))
// 	}
// 	for x := range f.cells {
// 		for y := range f.cells[x] {
// 			fx, fy := float64(x)-f.cells[x][y].vx, float64(y)-f.cells[x][y].vy
// 			ix, iy := int(fx), int(fy)
// 			jx, jy := fx-float64(ix), fy-float64(iy)
// 			z1 := lerp(d(ix, iy), d(ix+1, iy), jx)
// 			z2 := lerp(d(ix, iy+1), d(ix+1, iy+1), jx)
// 			dn[x][y] = lerp(z1, z2, jy)
// 		}
// 	}
// 	for x := range dn {
// 		for y := range dn[x] {
// 			f.cells[x][y].d = dn[x][y]
// 		}
// 	}
// }

func (f *Fluid) Update() {

}
