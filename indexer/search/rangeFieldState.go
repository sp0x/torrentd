package search

import (
	"fmt"
	"strconv"
)

type RangeFieldState struct {
	start   string
	end     string
	current string
}

func NewRangeFieldState(start, end string) *RangeFieldState {
	return &RangeFieldState{
		start,
		end,
		"",
	}
}

func (r *RangeFieldState) Next() string {
	switch {
	case r.current == "":
		r.current = r.start
	case !r.HasNext():
		return r.current
	default:
		r.increment()
	}
	return r.current
}

func (r *RangeFieldState) String() string {
	return r.current
}

func (r *RangeFieldState) HasNext() bool {
	return r.current != r.end
}

func (r *RangeFieldState) increment() {
	length := len(r.current)
	num, _ := strconv.Atoi(r.current)
	num++
	r.current = fmt.Sprintf("%0"+strconv.Itoa(length)+"d", num)
}
