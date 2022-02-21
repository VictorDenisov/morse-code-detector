package main

import (
	"log"
	"sort"
	"time"

	gc "github.com/rthornton128/goncurses"
)

const (
	HEIGHT = 10
	WIDTH  = 30
)

type Element struct {
	d int
	s bool
}

func main() {
	stdscr, err := gc.Init()
	if err != nil {
		log.Fatal("init:", err)
	}
	defer gc.End()

	gc.Raw(true)
	gc.Echo(false)
	gc.Cursor(0)
	stdscr.Keypad(true)

	_, err = gc.NewWindow(HEIGHT, WIDTH, 5, 2)
	if err != nil {
		log.Fatal("new_window:", err)
	}
	stdscr.MovePrint(0, 0, "Use mouse button to signal morse code or 'q' to quit")
	stdscr.Refresh()

	if gc.MouseOk() {
		stdscr.MovePrint(3, 0, "WARN: Mouse support not detected.")
	}

	gc.MouseInterval(0) // Allow arbitrary duration between press and release of mouse button.

	gc.MouseMask(gc.M_B1_PRESSED|gc.M_B1_RELEASED, nil) // only detect left mouse press and release

	ds := make([]Element, 0)

	lastTime := time.Now()

	var key gc.Key
	for key != 'q' {
		key = stdscr.GetChar()
		switch key {
		case 'c':
			ditMean, dahMean := inferSignals(ds)
			res := make([]byte, len(ds))
			for i := 0; i < len(ds); i++ {
				if ds[i].s {
					if abs(ds[i].d-ditMean) < abs(ds[i].d-dahMean) {
						res[i] = '.'
					} else {
						res[i] = '-'
					}
				} else {
					res[i] = ' '
				}
			}
			stdscr.MovePrintf(20, 0, "Durations %v", string(res))
		case gc.KEY_MOUSE:
			if md := gc.GetMouse(); md != nil {
				if md.State == gc.M_B1_PRESSED {
					newTime := time.Now()
					dr := newTime.Sub(lastTime)
					lastTime = newTime

					ds = append(ds, Element{int(dr / 1000000), false})

					stdscr.MovePrintf(22, 0, "Mouse pressed = %3d/%c", key, key)
				} else if md.State == gc.M_B1_RELEASED {
					newTime := time.Now()
					dr := newTime.Sub(lastTime)
					lastTime = newTime

					ds = append(ds, Element{int(dr / 1000000), true})

					stdscr.MovePrintf(22, 0, "Mouse released = %3d/%c", key, key)
					stdscr.MovePrintf(24, 0, "Durations = %v", ds)
				}
			}
			fallthrough
		default:
			stdscr.MovePrintf(23, 0, "Character pressed = %3d/%c", key, key)
			stdscr.ClearToEOL()
			stdscr.Refresh()
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// K-means for classifying dots and dashes
func inferSignals(ds []Element) (int, int) {
	unsortedSignals := make([]int, 0)
	signals := make([]int, 0)
	for _, d := range ds {
		if d.s {
			unsortedSignals = append(unsortedSignals, int(d.d))
			signals = append(signals, int(d.d))
		}
	}
	sort.IntSlice(signals).Sort()

	lastDotMean := signals[0]
	lastDahMean := signals[len(signals)-1]
	for {
		border := 0
		for i, s := range signals {
			if abs(s-lastDotMean) > abs(s-lastDahMean) {
				border = i
				break
			}
		}
		ditMean := 0
		for i := 0; i < border; i++ {
			ditMean += signals[i]
		}
		ditMean /= border
		dahMean := 0
		for i := border; i < len(signals); i++ {
			dahMean += signals[i]
		}
		dahMean /= len(signals) - border
		if ditMean == lastDotMean && dahMean == lastDahMean {
			break
		}
		lastDotMean = ditMean
		lastDahMean = dahMean
	}
	return lastDotMean, lastDahMean
}
