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
	d time.Duration
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
			unsortedDurations, durations, signals := inferSignals(ds)
			stdscr.MovePrintf(19, 0, "Durations %v", unsortedDurations)
			stdscr.MovePrintf(20, 0, "Durations %v", durations)
			stdscr.MovePrintf(21, 0, "Signals %v", signals)
		case gc.KEY_MOUSE:
			if md := gc.GetMouse(); md != nil {
				if md.State == gc.M_B1_PRESSED {
					newTime := time.Now()
					dr := newTime.Sub(lastTime)
					lastTime = newTime

					ds = append(ds, Element{dr, false})

					stdscr.MovePrintf(22, 0, "Mouse pressed = %3d/%c", key, key)
				} else if md.State == gc.M_B1_RELEASED {
					newTime := time.Now()
					dr := newTime.Sub(lastTime)
					lastTime = newTime

					ds = append(ds, Element{dr, true})

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
func inferSignals(ds []Element) ([]int, []int, string) {
	unsortedSignals := make([]int, 0)
	signals := make([]int, 0)
	for _, d := range ds {
		if d.s {
			unsortedSignals = append(unsortedSignals, int(int64(d.d)/1000000))
			signals = append(signals, int(int64(d.d)/1000000))
		}
	}
	sort.IntSlice(signals).Sort()
	dotLen := signals[0]
	dahLen := signals[len(signals)-1]

	lastDotMean := dotLen
	lastDahMean := dahLen
	for {
		border := 0
		for i, s := range signals {
			if abs(s-dotLen) > abs(s-dahLen) {
				border = i
				break
			}
		}
		dotMean := 0
		for i := 0; i < border; i++ {
			dotMean += signals[i]
		}
		dotMean /= border
		dahMean := 0
		for i := border; i < len(signals); i++ {
			dahMean += signals[i]
		}
		dahMean /= len(signals) - border
		if dotMean == lastDotMean && dahMean == lastDahMean {
			break
		}
		lastDotMean = dotMean
		lastDahMean = dahMean
	}

	res := make([]byte, len(unsortedSignals))
	for i := 0; i < len(unsortedSignals); i++ {
		if abs(unsortedSignals[i]-lastDotMean) < abs(unsortedSignals[i]-lastDahMean) {
			res[i] = '.'
		} else {
			res[i] = '-'
		}
	}
	return unsortedSignals, signals, string(res)
}
