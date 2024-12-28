package stack

import "github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"

type OpStack []cfg.Op

func NewOpStack() *OpStack {
	st := make([]cfg.Op, 0)
	return (*OpStack)(&st)
}

func (st *OpStack) Push(op cfg.Op) {
	*st = append(*st, op)
}

func (st *OpStack) Pop() (cfg.Op, bool) {
	if len(*st) == 0 {
		return nil, false
	}
	lastItem := (*st)[len(*st)-1]
	*st = (*st)[:(len(*st) - 1)]
	return lastItem, true
}

func (st *OpStack) IsEmpty() bool {
	return len(*st) == 0
}
