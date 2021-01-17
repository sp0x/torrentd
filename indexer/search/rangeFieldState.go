package search

import (
	"fmt"
	"strconv"
)

type rangeFieldState struct {
	start   string
	end     string
	current string
}

func (r *rangeFieldState) Next() string {
	if r.current == "" {
		r.current = r.start
	} else if !r.HasNext() {
		return r.current
	} else {
		r.increment()
	}
	return r.current
}

func (r *rangeFieldState) String() string {
	return r.current
}

func (r *rangeFieldState) HasNext() bool {
	return r.current != r.end
}

func (r *rangeFieldState) increment() {
	length := len(r.current)
	num, _ := strconv.Atoi(r.current)
	num++
	r.current = fmt.Sprintf("%0"+strconv.Itoa(length)+"d", num)
}
