package main

import (
	"fmt"
	"log"
	"os"
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
			ditMean, dahMean := classifySignals(ds)
			ditGap := ditMean
			charGap := dahMean
			wordGap := 7 * ditGap
			res := make([]byte, len(ds))
			for i := 0; i < len(ds); i++ {
				if ds[i].s {
					if abs(ds[i].d-ditMean) < abs(ds[i].d-dahMean) {
						res[i] = '.'
					} else {
						res[i] = '-'
					}
				} else {
					if abs(ds[i].d-ditGap) < abs(ds[i].d-charGap) && abs(ds[i].d-ditGap) < abs(ds[i].d-wordGap) {
						res[i] = ' '
					}
					if abs(ds[i].d-charGap) < abs(ds[i].d-ditGap) && abs(ds[i].d-charGap) < abs(ds[i].d-wordGap) {
						res[i] = '|'
					}
					if abs(ds[i].d-wordGap) < abs(ds[i].d-ditGap) && abs(ds[i].d-wordGap) < abs(ds[i].d-charGap) {
						res[i] = '>'
					}
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

func classifyGaps(ds []Element) (ditGap int, charGap int, wordGap int) {
	f, err := os.Create("classifyGaps.log")
	if err != nil {
		os.Exit(1)
	}
	defer f.Close()
	unsortedGaps := make([]int, 0)
	gaps := make([]int, 0)
	for _, d := range ds {
		if !d.s {
			unsortedGaps = append(unsortedGaps, int(d.d))
			gaps = append(gaps, int(d.d))
		}
	}
	sort.IntSlice(gaps).Sort()

	lastDitGap := gaps[0]
	lastCharGap := lastDitGap * 3
	lastWordGap := gaps[len(gaps)-1]

	fmt.Fprintf(f, "Gaps: %v\n", gaps)

	for {
		fmt.Fprintf(f, "DitGap: %v, CharGap: %v, WordGap: %v\n", lastDitGap, lastCharGap, lastWordGap)
		border1 := 0
		border2 := 0
		for i, s := range gaps {
			if abs(s-lastDitGap) > abs(s-lastCharGap) {
				border1 = i
				break
			}
		}
		for i, s := range gaps {
			if abs(s-lastCharGap) > abs(s-lastWordGap) {
				border2 = i
				break
			}
		}
		fmt.Fprintf(f, "border1: %v, border2: %v\n", border1, border2)
		ditGapMean := 0
		for i := 0; i < border1; i++ {
			ditGapMean += gaps[i]
		}
		ditGapMean /= border1

		charGapMean := 0
		for i := border1; i < border2; i++ {
			charGapMean += gaps[i]
		}
		charGapMean /= border2 - border1

		wordGapMean := 0
		for i := border2; i < len(gaps); i++ {
			wordGapMean += gaps[i]
		}
		wordGapMean /= len(gaps) - border2

		if ditGapMean == lastDitGap && charGapMean == lastCharGap && wordGapMean == lastWordGap {
			break
		}
		lastDitGap = ditGapMean
		lastCharGap = charGapMean
		lastWordGap = wordGapMean
	}

	return lastDitGap, lastCharGap, lastWordGap
}

// K-means for classifying dots and dashes
func classifySignals(ds []Element) (int, int) {
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
