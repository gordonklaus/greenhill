package main

import (
	"image"
	"strings"
	"time"

	"gioui.org/font/gofont"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type C = layout.Context
type D = layout.Dimensions

var th = material.NewTheme(gofont.Collection())

func init() {
	th.Palette = component.SwapGrounds(th.Palette)
}

type UI struct {
	core *Core
	tree *TreeView

	hideCursorTime time.Time
	scroll         struct{ x, y gesture.Scroll }
	offset         image.Point
}

func NewUI(core *Core) *UI {
	ui := &UI{
		core: core,
		tree: NewTreeView(core, core.Tree, nil),
	}
	ui.tree.Focus()
	return ui
}

func (ui *UI) Layout(gtx C) D {
	for _, e := range gtx.Events(ui) {
		switch e := e.(type) {
		case key.Event:
			if e.State == key.Press {
				switch e.Name {
				case "Z":
					if e.Modifiers == key.ModShortcut {
						ui.core.Undo()
					} else if e.Modifiers == key.ModShortcut|key.ModShift {
						ui.core.Redo()
					}
				}
			}
		case pointer.Event:
			pointer.CursorDefault.Add(gtx.Ops)
			ui.hideCursorTime = gtx.Now.Add(time.Second)
		}
	}

	if gtx.Now.After(ui.hideCursorTime) {
		pointer.CursorNone.Add(gtx.Ops)
	}

	key.InputOp{
		Tag: ui,
		Keys: keySet(
			"Short-(Shift)-Z",
		),
	}.Add(gtx.Ops)

	pointer.InputOp{Tag: ui, Types: pointer.Move}.Add(gtx.Ops)

	ui.scroll.x.Add(gtx.Ops, image.Rect(-1e9, 0, 1e9, 0))
	ui.scroll.y.Add(gtx.Ops, image.Rect(0, -1e9, 0, 1e9))

	ui.offset = ui.offset.Sub(image.Pt(
		ui.scroll.x.Scroll(gtx.Metric, gtx, gtx.Now, gesture.Horizontal),
		ui.scroll.y.Scroll(gtx.Metric, gtx, gtx.Now, gesture.Vertical),
	))

	dims := D{Size: gtx.Constraints.Max}

	paint.FillShape(gtx.Ops, th.Bg, clip.Rect{Max: gtx.Constraints.Max}.Op())

	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max = image.Pt(1e9, 1e9)
	tree := Record(gtx, ui.tree.Layout)

	d := dims.Size.Sub(tree.Dims.Size)
	if d.X > 0 {
		ui.offset.X = d.X / 2
	} else if ui.offset.X > 0 {
		ui.offset.X = 0
	} else if ui.offset.X < d.X {
		ui.offset.X = d.X
	}
	if d.Y > 0 {
		ui.offset.Y = d.Y / 2
	} else if ui.offset.Y > 0 {
		ui.offset.Y = 0
	} else if ui.offset.Y < d.Y {
		ui.offset.Y = d.Y
	}
	defer op.Offset(ui.offset).Push(gtx.Ops).Pop()
	tree.Layout(gtx)

	return dims
}

func keySet(s ...string) key.Set {
	return key.Set(strings.Join(s, "|"))
}
