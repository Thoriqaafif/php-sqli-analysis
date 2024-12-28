package sourcefinder

import (
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis/taintutil"
)

type SourceFinder struct {
	traverser.NullTraverser

	CurrScript *cfg.Script
	CurrFunc   *cfg.OpFunc
}

func NewSourceFinder() *SourceFinder {
	return &SourceFinder{}
}

func (t *SourceFinder) EnterScript(script *cfg.Script) {
	t.CurrScript = script
}

func (t *SourceFinder) EnterFunc(fn *cfg.OpFunc) {
	t.CurrFunc = fn
}

func (t *SourceFinder) LeaveFunc(fn *cfg.OpFunc) {
	t.CurrFunc = nil
}

func (t *SourceFinder) EnterOp(op cfg.Op, block *cfg.Block) {
	// if source, add to sources
	if taintutil.IsSource(op) || t.CallToTaintedFunc(op) {
		t.CurrFunc.Sources = append(t.CurrFunc.Sources, op)
	}
}

func (t *SourceFinder) CallToTaintedFunc(call cfg.Op) bool {
	funcName := ""
	switch opT := call.(type) {
	case *cfg.OpExprFunctionCall:
		funcName = opT.GetName()
	case *cfg.OpExprMethodCall:
		funcName = opT.GetName()
	case *cfg.OpExprStaticCall:
		funcName = opT.GetName()
	}
	if fn, ok := t.CurrScript.Funcs[funcName]; ok {
		if fn.ContaintTainted {
			return true
		}
		for _, fnCall := range fn.Calls {
			return t.CallToTaintedFunc(fnCall)
		}
	}
	return false
}
