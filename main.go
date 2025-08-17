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

func lerpColor(c1, c2 tcell.Color, ratio float64) tcell.Color {
	r1, g1, b1 := c1.RGB()
	r2, g2, b2 := c2.RGB()
	r := int32(float64(r1) + (float64(r2-r1) * ratio))
	g := int32(float64(g1) + (float64(g2-g1) * ratio))
	b := int32(float64(b1) + (float64(b2-b1) * ratio))
	return tcell.NewRGBColor(r, g, b)
}

func drawScreen(s tcell.Screen, f *fluid.Fluid) {
	fg := tcell.GetColor(cfg.fg)
	bg := tcell.GetColor(cfg.bg)
	w, h := s.Size()
	for x := range w {
		for y := range h {
			y1, y2 := y*2, y*2+1
			mut.RLock()
			b1, b2 := math.Min(f.D(x, y1), 1), math.Min(f.D(x, y2), 1)
			mut.RUnlock()
			st := tcell.StyleDefault
			st = st.Background(lerpColor(bg, fg, b1))
			st = st.Foreground(lerpColor(bg, fg, b2))
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
		time.Sleep(time.Millisecond * 33)
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
		if !cfg.paused {
			w, h := s.Size()
			p1x, p1y := rand.Float64()*float64(w), rand.Float64()*float64(h)*2
			p2x, p2y := rand.Float64()*float64(w), rand.Float64()*float64(h)*2
			setLine(f, int(p1x), int(p1y), int(p2x), int(p2y), 8, time.Millisecond*5)
		}
		time.Sleep(time.Millisecond*time.Duration(rand.IntN(3000)) + 1000)
	}
}

type config struct {
	autoRunDisabled bool
	fg, bg          string
	paused          bool
}

var cfg config = config{}
var prevMouseX, prevMouseY int
var mut sync.RWMutex
var quit chan struct{}

func main() {

	viscosity := flag.Float64("v", 0.2, "viscosity")
	decay := flag.Float64("d", 0.02, "decay")
	iters := flag.Int("i", 7, "iterations")
	flag.BoolVar(&cfg.autoRunDisabled, "a", false, "disable auto run")
	flag.StringVar(&cfg.fg, "fg", "#00aaff", "foreground color")
	flag.StringVar(&cfg.bg, "bg", "#000000", "background color")
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
	if !cfg.autoRunDisabled {
		go autoRun(s, f)
	}
	go eventLoop(s, f)

	<-quit
	s.Fini()
}
