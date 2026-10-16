package main

import (
	"math/rand"
	"sync"
)

type randomItemGenerator struct {
	items   []generatedItem
	index   int
	mtx     *sync.Mutex
	shuffle *sync.Once
}

func (r *randomItemGenerator) reset() {
	r.mtx = &sync.Mutex{}
	r.shuffle = &sync.Once{}

	r.items = []generatedItem{
		{title: "Nano", description: "You're strange"},
		{title: "Vim", description: "Bro think is Linus"},
		{title: "Emacs", description: "Come on man"},
		{title: "Sublime", description: "????"},
		{title: "Micro", description: "Goat choice"},
		{title: "Vscode", description: "Normie"},
	}

	r.shuffle.Do(func() {
		rand.Shuffle(len(r.items), func(i, j int) {
			r.items[i], r.items[j] = r.items[j], r.items[i]
		})
	})
}

func (r *randomItemGenerator) next() item {
	if r.mtx == nil {
		r.reset()
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	generated := r.items[r.index]

	r.index++
	if r.index >= len(r.items) {
		r.index = 0
	}

	return item{
		title:       generated.title,
		description: generated.description,
	}
}

type generatedItem struct {
	title       string
	description string
}
