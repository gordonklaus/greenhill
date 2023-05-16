package main

import (
	"encoding/gob"
	"errors"
	"log"
	"os"

	"github.com/gordonklaus/greenhill/model"
)

type Core struct {
	filename string
	Tree     *model.Tree
	undo     Undo
}

func NewCore() *Core {
	c := &Core{}

	if len(os.Args) != 2 {
		log.Fatal("Expected 1 argument")
	}
	c.filename = os.Args[1]

	var err error
	c.Tree, err = c.loadTree()
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func (c *Core) Do(do, undo func()) {
	c.undo.Do(
		func() { do(); c.Save() },
		func() { undo(); c.Save() },
	)
}
func (c *Core) Undo() { c.undo.Undo() }
func (c *Core) Redo() { c.undo.Redo() }

func (c *Core) Save() {
	f, err := os.Create(c.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(c.Tree); err != nil {
		log.Fatal(err)
	}
}

func (c *Core) loadTree() (*model.Tree, error) {
	t := &model.Tree{}

	f, err := os.Open(c.filename)
	if errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(c.filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if err := gob.NewEncoder(f).Encode(t); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		defer f.Close()
		if err := gob.NewDecoder(f).Decode(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}
