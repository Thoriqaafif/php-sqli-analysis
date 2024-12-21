package traverser

import "github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"

type NullTraverser struct {
}

func (t *NullTraverser) EnterScript(*cfg.Script) {}

func (t *NullTraverser) LeaveScript(*cfg.Script) {}

func (t *NullTraverser) EnterFunc(*cfg.OpFunc) {}

func (t *NullTraverser) LeaveFunc(*cfg.OpFunc) {}

func (t *NullTraverser) EnterBlock(*cfg.Block, *cfg.Block) {}

func (t *NullTraverser) LeaveBlock(*cfg.Block, *cfg.Block) {}

func (t *NullTraverser) SkipBlock(*cfg.Block, *cfg.Block) {}

func (t *NullTraverser) EnterOp(cfg.Op, *cfg.Block) {}

func (t *NullTraverser) LeaveOp(cfg.Op, *cfg.Block) {}
