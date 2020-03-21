package id

import (
	"sync"
)

// Seq stands for sequence
type Seq struct {
	m   sync.Mutex
	seq uint64
}

func New() *Seq {
	return &Seq{}
}

func (i *Seq) Next() uint64 {
	i.m.Lock()
	defer i.m.Unlock()

	i.seq++
	return i.seq
}
