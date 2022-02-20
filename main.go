package main

import (
	"log"

	gc "github.com/rthornton128/goncurses"
)

const (
	HEIGHT = 10
	WIDTH  = 30
)

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

	var key gc.Key
	for key != 'q' {
		key = stdscr.GetChar()
		switch key {
		case gc.KEY_MOUSE:
			if md := gc.GetMouse(); md != nil {
				if md.State == gc.M_B1_PRESSED {
					stdscr.MovePrintf(22, 0, "Mouse pressed = %3d/%c", key, key)
				} else if md.State == gc.M_B1_RELEASED {
					stdscr.MovePrintf(22, 0, "Mouse released = %3d/%c", key, key)
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
