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
				dx, dy := float64(x-prevMouseX), float64(y-prevMouseY)
				dist := math.Hypot(dx, dy)
				if dist < 1 {
					continue
				}
				dx, dy = dx/dist, dy/dist
				for i := range int(dist) {
					xx := int(float64(prevMouseX) + dx*float64(i))
					yy := int(float64(prevMouseY) + dy*float64(i))
					mut.Lock()
					f.Set(xx, yy*2, 15.0, dx*8.0, dy*8.0)
					mut.Unlock()
				}
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

func autoRun(s tcell.Screen, f *fluid.Fluid) {
	for {
		if !cfg.paused {
			w, h := s.Size()
			p1x, p1y := rand.Float64()*float64(w), rand.Float64()*float64(h)*2
			p2x, p2y := rand.Float64()*float64(w), rand.Float64()*float64(h)*2
			dx, dy := p2x-p1x, p2y-p1y
			dist := math.Hypot(dx, dy)
			if dist < 1 {
				continue
			}
			dx, dy = dx/dist, dy/dist
			delay := time.Duration(rand.IntN(2000) + 2000)
			for i := range int(dist) {
				xx := int(p1x + dx*float64(i))
				yy := int(p1y + dy*float64(i))
				mut.Lock()
				f.Set(xx, yy, 15.0, dx*8.0, dy*8.0)
				mut.Unlock()
				time.Sleep(time.Microsecond * delay)
				delay += 100
			}
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

	viscosity := flag.Float64("v", 0.1, "viscosity")
	decay := flag.Float64("d", 0.01, "decay")
	iters := flag.Int("i", 5, "iterations")
	flag.BoolVar(&cfg.autoRunDisabled, "a", false, "disable auto run")
	flag.StringVar(&cfg.fg, "fg", "#ff0000", "foreground color")
	flag.StringVar(&cfg.bg, "bg", "#00eeff", "background color")
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
