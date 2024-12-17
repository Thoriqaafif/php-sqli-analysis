package traverser

import (
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg"
)

type RemoveType int

const (
	REMOVE_BLOCK RemoveType = iota + 102
	REMOVE_OP
)

type BlockTraverser interface {
	EnterScript(*cfg.Script)
	LeaveScript(*cfg.Script)
	EnterFunc(*cfg.OpFunc)
	LeaveFunc(*cfg.OpFunc)
	EnterBlock(*cfg.Block, *cfg.Block)
	LeaveBlock(*cfg.Block, *cfg.Block)
	SkipBlock(*cfg.Block, *cfg.Block)
	EnterOp(cfg.Op, *cfg.Block)
	LeaveOp(cfg.Op, *cfg.Block)
}

type Traverser struct {
	BlockTravs []BlockTraverser

	seenBlock map[*cfg.Block]struct{}
}

func NewTraverser() *Traverser {
	return &Traverser{
		BlockTravs: make([]BlockTraverser, 0),
	}
}

func (t *Traverser) AddTraverser(traverser BlockTraverser) {
	t.BlockTravs = append(t.BlockTravs, traverser)
}

func (t *Traverser) Traverse(script *cfg.Script) {
	t.EnterScript(script)
	t.TraverseFunc(script.Main)
	for _, fn := range script.Funcs {
		t.TraverseFunc(fn)
	}
	t.LeaveScript(script)
}

func (t *Traverser) TraverseFunc(fn *cfg.OpFunc) {
	t.seenBlock = make(map[*cfg.Block]struct{})
	t.EnterFunc(fn)
	block := fn.Cfg
	if block != nil {
		t.TraverseBlock(block, nil)
		// TODO: add remove block functionality if needed
	}
	t.LeaveFunc(fn)
	t.seenBlock = nil
}

func (t *Traverser) TraverseBlock(block *cfg.Block, prior *cfg.Block) {
	if t.InSeenBlock(block) {
		t.SkipBlock(block, prior)
		return
	}
	t.AddSeenBlock(block)
	t.EnterBlock(block, prior)
	ops := block.Instructions

	for _, op := range ops {
		t.EnterOp(op, block)
		for _, subblock := range cfg.GetSubBlocks(op) {
			t.TraverseBlock(subblock, block)
		}
		t.LeaveOp(op, block)
	}

	t.LeaveBlock(block, prior)
}

func (t *Traverser) EnterScript(script *cfg.Script) {
	for _, trav := range t.BlockTravs {
		trav.EnterScript(script)
	}
}

func (t *Traverser) LeaveScript(script *cfg.Script) {
	for _, trav := range t.BlockTravs {
		trav.LeaveScript(script)
	}
}

func (t *Traverser) EnterBlock(block *cfg.Block, prior *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.EnterBlock(block, prior)
	}
}

func (t *Traverser) LeaveBlock(block *cfg.Block, prior *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.LeaveBlock(block, prior)
	}
}

func (t *Traverser) SkipBlock(block *cfg.Block, prior *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.SkipBlock(block, prior)
	}
}

func (t *Traverser) EnterFunc(fn *cfg.OpFunc) {
	for _, trav := range t.BlockTravs {
		trav.EnterFunc(fn)
	}
}

func (t *Traverser) LeaveFunc(fn *cfg.OpFunc) {
	for _, trav := range t.BlockTravs {
		trav.LeaveFunc(fn)
	}
}

func (t *Traverser) EnterOp(op cfg.Op, block *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.EnterOp(op, block)
	}
}

func (t *Traverser) LeaveOp(op cfg.Op, block *cfg.Block) {
	for _, trav := range t.BlockTravs {
		trav.LeaveOp(op, block)
	}
}

func (t *Traverser) AddSeenBlock(block *cfg.Block) {
	t.seenBlock[block] = struct{}{}
}

func (t *Traverser) InSeenBlock(block *cfg.Block) bool {
	if _, ok := t.seenBlock[block]; ok {
		return true
	}
	return false
}
