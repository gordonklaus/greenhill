package main

import (
	"gioui.org/layout"
	"gioui.org/op"
)

type Recording struct {
	Call op.CallOp
	Dims D
}

func Record(gtx C, w layout.Widget) Recording {
	m := op.Record(gtx.Ops)
	return Recording{
		Dims: w(gtx),
		Call: m.Stop(),
	}
}

func (r Recording) Layout(gtx C) D {
	r.Call.Add(gtx.Ops)
	return r.Dims
}
