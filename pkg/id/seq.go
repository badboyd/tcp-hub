package id

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Seq stands for sequence
type Seq struct {
	m   sync.Mutex
	seq uint64
}

// New retunrs new Sequence
func New() *Seq {
	return &Seq{}
}

// Next returns the next seq
func (i *Seq) Next() uint64 {
	i.m.Lock()
	defer i.m.Unlock()

	i.seq++
	return i.seq
}

// ConvertFromStringToArray helpers for translate from list of id separated by comma to a id array
func ConvertFromStringToArray(s string) ([]uint64, error) {
	receivers := []uint64{}
	for _, word := range strings.Split(s, ",") {
		id, err := strconv.ParseUint(strings.TrimSpace(word), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Unknown ID format")
		}

		receivers = append(receivers, id)
	}
	return receivers, nil
}

// JoinIDArray joins a id array to a string separated by delim
func JoinIDArray(ids []uint64, delim string) string {
	tmpArr := []string{}
	for _, id := range ids {
		tmpArr = append(tmpArr, fmt.Sprint(id))
	}
	return strings.Join(tmpArr, delim)
}
