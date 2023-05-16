package main

import (
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
)

func main() {
	go Main()
	app.Main()
}

func Main() {
	w := app.NewWindow(app.Title("Green Hill"), app.Decorated(false))
	w.Perform(system.ActionMaximize)

	core := NewCore()
	ui := NewUI(core)

	var ops op.Ops
	for e := range w.Events() {
		switch e := e.(type) {
		case system.DestroyEvent:
			os.Exit(0)
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			ui.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
