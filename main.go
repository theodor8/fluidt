package main

import (
	"flag"
	"log"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"fluidt/fluid"

	"github.com/gdamore/tcell/v2"
)

func pollEvents(s tcell.Screen, f *fluid.Fluid) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			w, h := ev.Size()
			mut.Lock()
			f.Resize(w, h*2)
			mut.Unlock()
			s.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC, tcell.KeyEscape:
				close(quit)
				return
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					close(quit)
					return
				case ' ':
					cfg.paused = !cfg.paused
				case 'r':
					mut.Lock()
					f.Reset()
					mut.Unlock()
					drawScreen(s, f)
				case 's':
					mut.Lock()
					f.Swirl(float64(prevMouseX), float64(prevMouseY*2), 15, 1.5)
					mut.Unlock()
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			if cfg.paused || (prevMouseX == 0 && prevMouseY == 0) {
				prevMouseX, prevMouseY = x, y
				continue
			}
			switch ev.Buttons() {
			case tcell.Button1, tcell.Button2:
				setLine(f, prevMouseX, prevMouseY*2, x, y*2, 8, 0)
			}
			prevMouseX, prevMouseY = x, y
		}
	}
}

// hsvToRGB converts HSV (h in [0,360), s and v in [0,1]) to a tcell color.
func hsvToRGB(h, s, v float64) tcell.Color {
	h = math.Mod(h, 360)
	i := int(h / 60)
	f := h/60 - float64(i)
	p, q, t := v*(1-s), v*(1-s*f), v*(1-s*(1-f))
	var r, g, b float64
	switch i {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	return tcell.NewRGBColor(int32(r*255), int32(g*255), int32(b*255))
}

func colorRGB(c tcell.Color) (int32, int32, int32) {
	h := c.Hex()
	return (h >> 16) & 0xff, (h >> 8) & 0xff, h & 0xff
}

func lerpColor(a, b tcell.Color, t float64) tcell.Color {
	ar, ag, ab := colorRGB(a)
	br, bg2, bb := colorRGB(b)
	r := int32(float64(ar)*(1-t) + float64(br)*t)
	g := int32(float64(ag)*(1-t) + float64(bg2)*t)
	bv := int32(float64(ab)*(1-t) + float64(bb)*t)
	return tcell.NewRGBColor(r, g, bv)
}

func fluidColor(bg tcell.Color, density, speed float64) tcell.Color {
	if density <= 0 {
		return bg
	}
	// cyan (slow) → pink (fast), blend from bg at low density
	hue := 180 + math.Min(speed/2.0, 1.0)*140
	return lerpColor(bg, hsvToRGB(hue, 1.0, 1.0), math.Min(density, 1.0))
}

func drawScreen(s tcell.Screen, f *fluid.Fluid) {
	bg := tcell.GetColor(cfg.bg)
	w, h := s.Size()
	for x := range w {
		for y := range h {
			y1, y2 := y*2, y*2+1
			mut.RLock()
			d1, d2 := f.D(x, y1), f.D(x, y2)
			sp1, sp2 := f.AvgSpeed(x, y1, 4), f.AvgSpeed(x, y2, 4)
			mut.RUnlock()
			st := tcell.StyleDefault
			st = st.Background(fluidColor(bg, d1, sp1))
			st = st.Foreground(fluidColor(bg, d2, sp2))
			s.SetContent(x, y, '▄', nil, st)
		}
	}
	s.Show()
}

func eventLoop(s tcell.Screen, f *fluid.Fluid) {
	for {
		if !cfg.paused {
			mut.Lock()
			f.Update()
			mut.Unlock()
			drawScreen(s, f)
		}
		time.Sleep(time.Duration(float64(time.Millisecond*33) / cfg.speed))
	}
}

func setLine(f *fluid.Fluid, x1, y1, x2, y2 int, v float64, delay time.Duration) {
	dx, dy := float64(x2-x1), float64(y2-y1)
	dist := math.Hypot(dx, dy)
	if dist < 1 {
		return
	}
	dirx, diry := dx/dist, dy/dist
	for i := range int(dist) {
		xx := int(float64(x1) + dirx*float64(i))
		yy := int(float64(y1) + diry*float64(i))
		mut.Lock()
		f.Set(xx, yy, 12.0, dirx*v, diry*v)
		mut.Unlock()
		time.Sleep(delay)
	}
}

func autoRun(s tcell.Screen, f *fluid.Fluid) {
	for {

		if cfg.paused {
			continue
		}

		w, h := s.Size()
		px, py := rand.Float64()*float64(w), rand.Float64()*float64(h)*2
		angle := rand.Float64() * 2 * math.Pi
		vx, vy := math.Cos(angle), math.Sin(angle)
		dist := rand.Float64()*50 + 50
		for range int(dist) {
			mut.Lock()
			f.Set(int(px), int(py), 12.0, vx*8, vy*8)
			mut.Unlock()
			px += vx
			py += vy
			if px < 0 || px >= float64(w) {
				vx = -vx
				px = math.Max(0, math.Min(px, float64(w-1)))
			}
			if py < 0 || py >= float64(h*2) {
				vy = -vy
				py = math.Max(0, math.Min(py, float64(h*2-1)))
			}

			time.Sleep(time.Millisecond * 5)
		}

		time.Sleep(time.Millisecond*time.Duration(rand.IntN(3000) + 1500))
	}
}

type config struct {
	autoRun bool
	bg      string
	paused  bool
	speed   float64
}

var cfg config = config{}
var prevMouseX, prevMouseY int
var mut sync.RWMutex
var quit chan struct{}

func main() {

	viscosity := flag.Float64("v", 0.2, "viscosity")
	decay := flag.Float64("d", 0.02, "decay")
	iters := flag.Int("i", 7, "iterations")
	flag.Float64Var(&cfg.speed, "s", 1.0, "speed multiplier")
	flag.BoolVar(&cfg.autoRun, "a", false, "auto run")
	flag.StringVar(&cfg.bg, "bg", "#ddd5c8", "background color")
	flag.Parse()

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.Init(); err != nil {
		log.Fatal(err)
	}

	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	w, h := s.Size()
	f := fluid.NewFluid(w, h*2, *viscosity, *decay, *iters)

	quit = make(chan struct{})
	go pollEvents(s, f)
	if cfg.autoRun {
		go autoRun(s, f)
	}
	go eventLoop(s, f)

	<-quit
	s.Fini()
}
