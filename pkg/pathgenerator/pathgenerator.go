package pathgenerator

import (
	"errors"
	"log"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis/taintutil"
)

type PathType int

const (
	NATIVE PathType = iota
	LARAVEL
)

type PathGenerator struct {
	paths    [][]cfg.Op
	currPath []cfg.Op
	vis      map[cfg.Op]map[cfg.Operand]struct{}
}

func NewPathGenerator() *PathGenerator {
	return &PathGenerator{
		paths: make([][]cfg.Op, 0),
	}
}

func GeneratePaths(scripts map[string]*cfg.Script, pathType PathType) ([][]cfg.Op, error) {
	switch pathType {
	case NATIVE:
		return phpGeneratePaths(scripts), nil
	case LARAVEL:
		return laravelGeneratePaths(scripts), nil
	}
	return nil, errors.New("invalid path type")
}

func phpGeneratePaths(scripts map[string]*cfg.Script) [][]cfg.Op {
	pg := NewPathGenerator()
	for _, script := range scripts {
		pg.vis = make(map[cfg.Op]map[cfg.Operand]struct{})
		pg.traverseFunc(*script.Main)
		for _, fn := range script.Funcs {
			pg.traverseFunc(*fn)
		}
	}
	return pg.paths
}

func laravelGeneratePaths(scripts map[string]*cfg.Script) [][]cfg.Op {
	// Get handler function
	pg := NewPathGenerator()
	handlerFuncs := make([]cfg.OpFunc, 0)
	for _, script := range scripts {
		for _, source := range script.Main.Sources {
			// ROUTE::
			if opRoute, ok := source.(*cfg.OpExprStaticCall); ok {
				handlerOper := opRoute.Args[1].GetWriteOp()
				// get handler closure
				if opClosure, ok := handlerOper.(*cfg.OpExprClosure); handlerOper != nil && ok {
					fn := opClosure.Func

					// set all parameter as tainted variable
					for _, stmt := range fn.Cfg.Instructions {
						if _, ok := stmt.(*cfg.OpExprParam); ok {
							fn.Sources = append(fn.Sources, stmt)
						}
					}
					handlerFuncs = append(handlerFuncs, *fn)
				}
			} else if fn, ok := source.(*cfg.OpFunc); ok {
				// method inside controller
				// set all parameter as tainted variable
				for _, stmt := range fn.Cfg.Instructions {
					if _, ok := stmt.(*cfg.OpExprParam); ok {
						fn.Sources = append(fn.Sources, stmt)
					}
				}
				handlerFuncs = append(handlerFuncs, *fn)
			}
		}
	}

	// traverse func with all parameter as source
	for _, fn := range handlerFuncs {
		pg.vis = make(map[cfg.Op]map[cfg.Operand]struct{})
		pg.traverseFunc(fn)
	}
	return pg.paths
}

func (pg *PathGenerator) traverseFunc(fn cfg.OpFunc) {
	for _, source := range fn.Sources {
		// dfs from source
		pg.currPath = []cfg.Op{source}
		sourceVar, err := taintutil.GetTaintedVar(source)
		if err != nil {
			continue
		}
		for _, sourceUser := range sourceVar.GetUsage() {
			err := pg.traceTaintedVar(sourceUser, sourceVar)
			if err != nil {
				log.Fatalf("Error generate php feasible path in '%s': %v", fn.FilePath, err)
			}
		}
	}
}

func (pg *PathGenerator) traceTaintedVar(node cfg.Op, taintedVar cfg.Operand) error {
	// stop if find sink, not propagated, or have been visited
	if taintutil.IsSink(node, taintedVar) {
		newPath := make([]cfg.Op, len(pg.currPath))
		copy(newPath, pg.currPath)
		newPath = append(newPath, node)

		// check if newPath feasible
		if pg.IsFeasiblePath(newPath) {
			pg.paths = append(pg.paths, newPath)
		}

		return nil
	} else if !taintutil.IsPropagated(node, taintedVar) {
		return nil
	} else if _, ok := pg.vis[node]; ok {
		if _, ok := pg.vis[node][taintedVar]; ok {
			return nil
		}
	}
	if _, ok := pg.vis[node]; !ok {
		pg.vis[node] = make(map[cfg.Operand]struct{})
	}
	pg.vis[node][taintedVar] = struct{}{}
	nodeVar, err := taintutil.GetTaintedVar(node)
	if err != nil {
		return nil
	}

	// get all operations using this tainted var
	for _, taintedUsage := range nodeVar.GetUsage() {
		// if _, ok := pg.vis[nodeVar][taintedUsage]; !ok {
		newPath := make([]cfg.Op, len(pg.currPath))
		copy(newPath, pg.currPath)
		newPath = append(newPath, taintedUsage)

		tmp := pg.currPath
		pg.currPath = newPath
		err := pg.traceTaintedVar(taintedUsage, nodeVar)
		if err != nil {
			return err
		}
		pg.currPath = tmp
		newPath = nil
		// pg.vis[nodeVar][taintedUsage] = struct{}{}
		// }
	}

	return nil
}

func (pg *PathGenerator) IsFeasiblePath(path []cfg.Op) bool {
	conds := make(map[cfg.Operand]struct{})
	for _, op := range path {
		if op.GetBlock() == nil {
			continue
		}
		for _, cond := range op.GetBlock().Conds {
			conds[cond] = struct{}{}
		}
	}

	// check all conditions
	for cond := range conds {
		condVal := cfg.GetOperVal(cond)
		switch cv := condVal.(type) {
		case *cfg.OperBool:
			if !cv.Val {
				return false
			}
		case *cfg.OperNumber:
			if cv.Val == 0 {
				return false
			}
		case *cfg.OperString:
			if cv.Val == "" || cv.Val == "0" {
				return false
			}
		}
	}

	return true
}
