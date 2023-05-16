package main

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"github.com/gordonklaus/greenhill/model"
	"golang.org/x/exp/slices"
)

type TreeView struct {
	core *Core
	tree *model.Tree

	parent   *TreeView
	text     *widget.Editor
	oldText  string
	children []*TreeView

	collapsed             bool
	requestFocus, focused bool
}

func NewTreeView(core *Core, tree *model.Tree, parent *TreeView) *TreeView {
	v := newTreeView(core, tree, parent)
	v.children = make([]*TreeView, len(tree.Children))
	for i, t := range tree.Children {
		v.children[i] = NewTreeView(v.core, t, v)
	}
	return v
}

func newTreeView(core *Core, tree *model.Tree, parent *TreeView) *TreeView {
	v := &TreeView{
		core:   core,
		tree:   tree,
		parent: parent,
		text:   &widget.Editor{Submit: true},
	}
	v.text.SetText(tree.Text)
	return v
}

func (v *TreeView) Focus() {
	if v != nil {
		v.requestFocus = true
	}
}

func (v *TreeView) NextChild(c *TreeView, next bool) *TreeView {
	if v == nil {
		return nil
	}

	i := slices.Index(v.children, c) - 1
	if next {
		i += 2
	}
	if i < 0 || i >= len(v.children) {
		c := v.parent.NextChild(v, next)
		if c != nil && len(c.children) > 0 {
			if i < 0 {
				return c.children[len(c.children)-1]
			}
			return c.children[0]
		}
		return c
	}
	return v.children[i]
}

func (v *TreeView) MoveChild(c *TreeView, down bool) {
	if v == nil {
		return
	}

	j := slices.Index(v.children, c)
	i := j - 1
	if down {
		i++
		j++
	}
	if i >= 0 && j < len(v.children) {
		do := func() {
			v.children[i], v.children[j] = v.children[j], v.children[i]
			v.tree.Children[i], v.tree.Children[j] = v.tree.Children[j], v.tree.Children[i]
		}
		v.core.Do(do, do)
	}
}

func (v *TreeView) NewParent() {
	t := *v.tree
	c := newTreeView(v.core, &t, v)
	v.core.Do(
		func() {
			c.children = v.children
			for _, cc := range c.children {
				cc.parent = c
			}
			v.tree.Text = ""
			v.tree.Children = []*model.Tree{c.tree}
			v.text.SetText("")
			v.children = []*TreeView{c}
			v.Focus()
		},
		func() {
			*v.tree = *c.tree
			v.text.SetText(v.tree.Text)
			v.children = c.children
			for _, c := range v.children {
				c.parent = v
			}
			v.Focus()
		},
	)
}

func copySlice[T any](s []T) []T {
	ss := make([]T, len(s))
	for i := range s {
		ss[i] = s[i]
	}
	return ss
}

func (v *TreeView) NewChild(after bool, c *TreeView) {
	if v == nil {
		return
	}

	i := slices.Index(v.children, c)
	if i < 0 {
		i = 0
		if after {
			i = len(v.children)
		}
	} else if after {
		i++
	}

	c2 := NewTreeView(v.core, &model.Tree{}, v)
	v.core.Do(
		func() {
			v.tree.Children = slices.Insert(v.tree.Children, i, c2.tree)
			v.children = slices.Insert(v.children, i, c2)
			c2.Focus()
		},
		func() {
			v.tree.Children = slices.Delete(v.tree.Children, i, i+1)
			v.children = slices.Delete(v.children, i, i+1)
			if c != nil {
				c.Focus()
			} else {
				v.Focus()
			}
		},
	)
	c2.text.Focus()
	v.collapsed = false
}

func (v *TreeView) DeleteRoot() {
	t := *v.tree
	c := v.children[0]
	v.core.Do(
		func() {
			*v.tree = *c.tree
			v.text.SetText(v.tree.Text)
			v.children = c.children
			for _, c := range v.children {
				c.parent = v
			}
			v.Focus()
		},
		func() {
			*v.tree = t
			v.text.SetText(v.tree.Text)
			v.children = []*TreeView{c}
			for _, cc := range c.children {
				cc.parent = c
			}
			v.Focus()
		},
	)
}

func (v *TreeView) DeleteChild(c *TreeView, reparentChildren, focusForward bool) {
	if v == nil {
		if len(c.children) == 1 && reparentChildren {
			c.DeleteRoot()
		}
		return
	}

	i := slices.Index(v.children, c)
	v.core.Do(
		func() {
			v.children = slices.Delete(v.children, i, i+1)
			v.tree.Children = slices.Delete(v.tree.Children, i, i+1)
			if reparentChildren {
				v.tree.Children = slices.Insert(v.tree.Children, i, c.tree.Children...)
				v.children = slices.Insert(v.children, i, c.children...)
				for _, c := range c.children {
					c.parent = v
				}
				if len(c.children) > 0 {
					c.children[(len(c.children)-1)/2].Focus()
				} else {
					v.Focus()
				}
				return
			}
			i := i
			if !focusForward && i > 0 || i == len(v.children) {
				i--
			}
			if len(v.children) > 0 {
				v.children[i].Focus()
			} else {
				v.Focus()
			}
		},
		func() {
			if reparentChildren {
				v.children = slices.Delete(v.children, i, i+len(c.children))
				v.tree.Children = slices.Delete(v.tree.Children, i, i+len(c.children))
				for _, cc := range c.children {
					cc.parent = c
				}
			}
			v.tree.Children = slices.Insert(v.tree.Children, i, c.tree)
			v.children = slices.Insert(v.children, i, c)
			c.Focus()
		},
	)
}

func (v *TreeView) Layout(gtx C) D {
	for _, e := range gtx.Events(v) {
		switch e := e.(type) {
		case key.FocusEvent:
			v.focused = e.Focus
		case key.Event:
			if e.State == key.Press {
				switch e.Name {
				case key.NameSpace:
					v.collapsed = !v.collapsed && len(v.children) > 0
				case key.NameLeftArrow:
					v.parent.Focus()
				case key.NameRightArrow:
					if len(v.children) > 0 && !v.collapsed {
						v.children[(len(v.children)-1)/2].Focus()
					}
				case key.NameUpArrow, key.NameDownArrow:
					if e.Modifiers.Contain(key.ModShortcut) {
						v.parent.MoveChild(v, e.Name == key.NameDownArrow)
					} else {
						shallowestCollapsed(v.parent.NextChild(v, e.Name == key.NameDownArrow)).Focus()
					}
				case key.NameDeleteBackward, key.NameDeleteForward:
					if e.Modifiers.Contain(key.ModShortcut) {
						v.parent.DeleteChild(v, e.Modifiers.Contain(key.ModAlt), e.Name == key.NameDeleteForward || e.Modifiers.Contain(key.ModShift))
					}
				case key.NameEnter, key.NameReturn:
					switch {
					case e.Modifiers&^key.ModShift == key.ModShortcut:
						if v.parent != nil {
							v.parent.NewChild(!e.Modifiers.Contain(key.ModShift), v)
						} else {
							v.NewChild(true, nil)
						}
					case e.Modifiers == key.ModShortcut|key.ModAlt:
						v.NewChild(true, nil)
					case e.Modifiers == key.ModShortcut|key.ModAlt|key.ModShift:
						v.NewParent()
					default:
						v.oldText = v.text.Text()
						v.text.Focus()
					}
				}
			}
		}
	}

	if v.requestFocus {
		v.requestFocus = false
		key.FocusOp{Tag: v}.Add(gtx.Ops)
	}
	key.InputOp{
		Tag: v,
		Keys: keySet(
			key.NameSpace,
			"(Shift)-"+key.NameTab,
			key.NameLeftArrow, key.NameRightArrow,
			"(Short)-"+key.NameUpArrow,
			"(Short)-"+key.NameDownArrow,
			"Short-(Alt)-(Shift)-"+key.NameDeleteBackward,
			"Short-(Alt)-(Shift)-"+key.NameDeleteForward,
			"(Short)-(Alt)-(Shift)-"+key.NameReturn,
			"(Short)-(Alt)-(Shift)-"+key.NameEnter,
		),
	}.Add(gtx.Ops)

	if v.collapsed {
		text := Record(gtx, v.layoutText)
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(text.Layout),
			layout.Rigid(func(gtx C) D {
				sz := image.Pt(gtx.Dp(30), text.Dims.Size.Y)
				for i := range v.children {
					var p clip.Path
					p.Begin(gtx.Ops)
					p1 := f32.Pt(0, float32(sz.Y)/2)
					p2 := f32.Pt(float32(sz.X), float32(sz.Y)*(float32(i)+.5)/float32(len(v.children)))
					d := f32.Pt(p2.X/2, 0)
					p.MoveTo(p1)
					p.CubeTo(p1.Add(d), p2.Sub(d), p2)
					st := clip.Stroke{
						Path:  p.End(),
						Width: 1,
					}.Op().Push(gtx.Ops)
					paint.LinearGradientOp{
						Stop1:  p1,
						Color1: th.Fg,
						Stop2:  p2,
						Color2: th.Bg,
					}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					st.Pop()
				}
				return D{Size: sz}
			}),
		)
	}

	childHeights := make([]int, len(v.children))
	children := Record(gtx, func(gtx C) D {
		children := make([]layout.FlexChild, len(v.children))
		for i, t := range v.children {
			i, t := i, t
			children[i] = layout.Rigid(func(gtx C) D {
				d := layout.Inset{Top: 2, Bottom: 2}.Layout(gtx, t.Layout)
				childHeights[i] = d.Size.Y
				return d
			})
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})

	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(v.layoutText),
		layout.Rigid(func(gtx C) D {
			if len(v.children) == 0 {
				return D{}
			}
			w := 3 * unit.Dp(math.Pow(float64(gtx.Metric.PxToDp(children.Dims.Size.Y)), 2./3))
			sz := image.Pt(gtx.Dp(w), children.Dims.Size.Y)
			y := 0
			for _, h := range childHeights {
				var p clip.Path
				p.Begin(gtx.Ops)
				p1 := f32.Pt(0, float32(sz.Y)/2)
				p2 := f32.Pt(float32(sz.X), float32(y+h/2))
				d := f32.Pt(float32(sz.X)/2, 0)
				p.MoveTo(p1)
				p.CubeTo(p1.Add(d), p2.Sub(d), p2)
				paint.FillShape(gtx.Ops, th.Fg, clip.Stroke{
					Path:  p.End(),
					Width: 1,
				}.Op())
				y += h
			}
			return D{Size: sz}
		}),
		layout.Rigid(children.Layout),
	)
}

func shallowestCollapsed(v *TreeView) *TreeView {
	for v2 := v; v2 != nil; v2 = v2.parent {
		if v2.collapsed {
			v = v2
		}
	}
	return v
}

func (v *TreeView) layoutText(gtx C) D {
	for _, e := range v.text.Events() {
		switch e.(type) {
		case widget.ChangeEvent:
			v.tree.Text = v.text.Text()
		case widget.SubmitEvent:
			if old, txt := v.oldText, v.text.Text(); txt != old {
				v.core.Do(
					func() { v.text = &widget.Editor{Submit: true}; v.text.SetText(txt) },
					func() { v.text = &widget.Editor{Submit: true}; v.text.SetText(old) },
				)
			}
			v.Focus()
		}
	}

	textTag := &v.text
	for _, e := range gtx.Events(textTag) {
		switch e := e.(type) {
		case key.Event:
			if e.State != key.Press {
				continue
			}
			switch e.Name {
			case key.NameEscape:
				if old, txt := v.oldText, v.text.Text(); txt != old {
					v.core.Do(
						func() { v.text = &widget.Editor{Submit: true}; v.text.SetText(txt) },
						func() { v.text = &widget.Editor{Submit: true}; v.text.SetText(old) },
					)
				}
				v.Focus()
			}
		}
	}

	if v.text.Focused() {
		key.InputOp{
			Tag: textTag,
			Keys: keySet(
				key.NameEscape,
				"(Shift)-Tab",
				key.NameLeftArrow, key.NameRightArrow, key.NameUpArrow, key.NameDownArrow,
			),
		}.Add(gtx.Ops)
	}

	const radius = unit.Dp(4)
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx C) D {
			if !v.focused {
				return D{}
			}
			return component.Rect{
				Size:  gtx.Constraints.Min,
				Color: component.WithAlpha(th.Fg, 96),
				Radii: gtx.Dp(radius),
			}.Layout(gtx)
		}),
		layout.Stacked(func(gtx C) D {
			return widget.Border{
				Color:        th.Fg,
				CornerRadius: radius,
				Width:        1,
			}.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Dp(5)
				gtx.Constraints.Max.X = gtx.Dp(300)
				return layout.UniformInset(radius).Layout(gtx, material.Editor(th, v.text, "").Layout)
			})
		}),
	)
}
