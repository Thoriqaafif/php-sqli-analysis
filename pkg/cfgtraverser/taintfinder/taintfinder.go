package taintfinder

import (
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis/taintutil"
)

type TaintFinder struct {
	traverser.NullTraverser

	CurrScript     *cfg.Script
	CurrFunc       *cfg.OpFunc
	ArrVars        map[*cfg.OpExprArrayDimFetch]string
	UnresolvedArrs map[*cfg.OpExprArrayDimFetch]string
	CurrClass      *cfg.OpStmtClass
}

func NewTaintFinder() *TaintFinder {
	return &TaintFinder{}
}

func (t *TaintFinder) EnterScript(script *cfg.Script) {
	t.CurrScript = script
}

func (t *TaintFinder) EnterFunc(fn *cfg.OpFunc) {
	t.CurrFunc = fn
	t.ArrVars = make(map[*cfg.OpExprArrayDimFetch]string)
	t.UnresolvedArrs = make(map[*cfg.OpExprArrayDimFetch]string)
}

func (t *TaintFinder) LeaveFunc(fn *cfg.OpFunc) {
	t.CurrFunc = nil
	t.ArrVars = nil
	t.UnresolvedArrs = nil
}

func (t *TaintFinder) EnterOp(op cfg.Op, block *cfg.Block) {
	// if source, add to sources
	if taintutil.IsSource(op) {
		t.CurrFunc.Sources = append(t.CurrFunc.Sources, op)
	} else if opClass, ok := op.(*cfg.OpStmtClass); ok {
		// laravel source
		extendClass, extendClassStr := cfg.GetOperVal(opClass.Extends).(*cfg.OperString)
		className, classNameStr := cfg.GetOperVal(opClass.Name).(*cfg.OperString)

		// controller class
		if extendClassStr && classNameStr && strings.HasSuffix(extendClass.Val, "Controller") {
			for methodName, methodFn := range t.CurrScript.Funcs {
				if strings.HasPrefix(methodName, className.Val) {
					t.CurrScript.Main.Sources = append(t.CurrScript.Main.Sources, methodFn)
				}
			}
		}
	}

	// Resolve ArrayDimFetch
	switch opT := op.(type) {
	case *cfg.OpExprArrayDimFetch:
		arrDimStr := opT.ToString()
		for arr, arrStr := range t.ArrVars {
			if strings.HasPrefix(arrDimStr, arrStr) && opT != arr {
				arr.Result.AddUsage(opT)
			}
		}
		t.UnresolvedArrs[opT] = arrDimStr
	case *cfg.OpExprAssign:
		for _, left := range opT.Var.GetWriteOps() {
			if left != nil {
				leftDef, ok := left.(*cfg.OpExprArrayDimFetch)
				leftDefStr := ""
				if ok {
					leftDefStr = leftDef.ToString()
				}
				if leftDefStr != "" {
					t.ArrVars[leftDef] = leftDefStr
					for arr, arrStr := range t.UnresolvedArrs {
						if strings.HasPrefix(arrStr, leftDefStr) {
							leftDef.Result.AddUsage(arr)
						}
					}
				}
			}
		}
	}
}

func (t *TaintFinder) CallToTaintedFunc(call cfg.Op) bool {
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
