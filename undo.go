package main

type Undo struct {
	items []undoItem
	i     int
}

type undoItem struct {
	do, undo func()
}

func (u *Undo) Do(do, undo func()) {
	do()
	u.items = append(u.items[:u.i], undoItem{do: do, undo: undo})
	u.i++
}

func (u *Undo) Undo() {
	if u.i > 0 {
		u.i--
		u.items[u.i].undo()
	}
}

func (u *Undo) Redo() {
	if u.i < len(u.items) {
		u.items[u.i].do()
		u.i++
	}
}
