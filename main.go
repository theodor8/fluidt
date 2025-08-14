package main

import (
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
					paused = !paused
				case 'r':
					mut.Lock()
					f.Reset()
					mut.Unlock()
					drawScreen(s, f)
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			if paused || (prevMouseX == 0 && prevMouseY == 0) {
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
					f.Set(xx, yy*2, 11.0, dx*8.0, dy*8.0)
					mut.Unlock()
				}
			}
			prevMouseX, prevMouseY = x, y
		}
	}
}

func drawScreen(s tcell.Screen, f *fluid.Fluid) {
	w, h := s.Size()
	for x := range w {
		for y := range h {
			y1, y2 := y*2, y*2+1
			mut.RLock()
			b1, b2 := int32(math.Min(f.D(x, y1), 1)*255), int32(math.Min(f.D(x, y2), 1)*255)
			mut.RUnlock()
			st := tcell.StyleDefault
			st = st.Background(tcell.NewRGBColor(b1, 255-b1, 255-b1))
			st = st.Foreground(tcell.NewRGBColor(b2, 255-b2, 255-b2))
			s.SetContent(x, y, '▄', nil, st)
		}
	}
	s.Show()
}

func eventLoop(s tcell.Screen, f *fluid.Fluid) {
	for {
		if !paused {
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
		if !paused {
			w, h := s.Size()
			p1x, p1y := rand.Float64() * float64(w), rand.Float64() * float64(h) * 2
			p2x, p2y := rand.Float64() * float64(w), rand.Float64() * float64(h) * 2
			dx, dy := p2x - p1x, p2y - p1y
			dist := math.Hypot(dx, dy)
			if dist < 1 {
				continue
			}
			dx, dy = dx/dist, dy/dist
			delay := time.Duration(rand.IntN(10) + 10)
			for i := range int(dist) {
				xx := int(p1x + dx*float64(i))
				yy := int(p1y + dy*float64(i))
				mut.Lock()
				f.Set(xx, yy, 11.0, dx*8.0, dy*8.0)
				mut.Unlock()
				time.Sleep(time.Millisecond * delay)
			}
		}
		time.Sleep(time.Millisecond * time.Duration(rand.IntN(5000)) + 1000)
	}
}

var prevMouseX, prevMouseY int
var paused bool = false
var mut sync.RWMutex
var quit chan struct{}

func main() {

	// TODO: optimize performance and memory usage
	// TODO: flags

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
	f := fluid.NewFluid(w, h*2, 0.15, 10)

	quit = make(chan struct{})
	go pollEvents(s, f)
	go autoRun(s, f)
	go eventLoop(s, f)

	<-quit
	s.Fini()
}
