package pathgenerator

import (
	"log"
	"reflect"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/condition"
)

type PathGenerator struct {
	Scripts     map[string]*cfg.Script
	CurrScript  *cfg.Script
	CurrPath    *PathCtx
	CurrFunc    *cfg.OpFunc
	CurrTainted map[cfg.Operand]struct{}

	CurrVar      int                  // helper to create name for variable used for z3 solver in the next
	PatWeightSet map[float64]struct{} // helper set to avoid duplicate path
}

// execution path
type PathCtx struct {
	Instructions []cfg.Op // first item for source, last item for sink
	Conds        []cfg.Operand
	Vars         map[cfg.Operand]struct{} // set of var defined in this path contex, can be helper to choose phi value

	Snb int // sum of normal block's unique number
	Cb  int // number of conditional blocks
	Lo  int // number of logical operations
}

func NewPathCtx() *PathCtx {
	return &PathCtx{
		Instructions: make([]cfg.Op, 0),
		Conds:        make([]cfg.Operand, 0),
		Snb:          0,
		Cb:           0,
		Lo:           0,
	}
}

func (p *PathCtx) AddInstruction(instruction cfg.Op) {
	p.Instructions = append(p.Instructions, instruction)
}

func (p *PathCtx) AddCondition(cond cfg.Operand) {
	p.Conds = append(p.Conds, cond)
}

func (p *PathCtx) Clone() *PathCtx {
	copiedInstructions := make([]cfg.Op, len(p.Instructions))
	copy(copiedInstructions, p.Instructions)
	copiedConds := make([]cfg.Operand, len(p.Conds))
	copy(copiedConds, p.Conds)
	return &PathCtx{
		Instructions: copiedInstructions,
		Conds:        copiedConds,
		Snb:          p.Snb,
		Cb:           p.Cb,
		Lo:           p.Lo,
	}
}

func GeneratePath(scripts map[string]*cfg.Script) []*PathCtx {
	generator := &PathGenerator{
		PatWeightSet: make(map[float64]struct{}),
	}

	// Generate path if script's Main contain tainted data
	for _, script := range scripts {
		if script.Main.ContaintTainted {
			// TODO: search sources as begining path
			// generate path with script.Main as entry block
			// generator.CurrPath = NewPath()

			//
			generator.CurrScript = script
			generator.CurrFunc = script.Main
			generator.TraverseBlock(script.Main.Cfg)
		}
	}

	return generator.Path
}

func (pg *PathGenerator) TraverseBlock(block *cfg.Block) {
	if block == nil || len(block.Instructions) <= 0 || block.Visited {
		return
	}
	block.Visited = true
	for i := 0; i < len(block.Instructions)-1; i++ {
		// find if there is a func call
		switch instruction := block.Instructions[i].(type) {
		case *cfg.OpExprFunctionCall:
			pg.CurrPath.AddInstruction(block.Instructions[i])
			// go to the function's blocks
			funcName := instruction.GetName()
			fn := pg.GetFunc(funcName)
			// TODO: check the argument and parameter
			tempFunc := pg.CurrFunc
			pg.CurrFunc = fn
			pg.TraverseBlock(fn.Cfg)
			pg.CurrFunc = tempFunc
		case *cfg.OpExprMethodCall:
			pg.CurrPath.AddInstruction(block.Instructions[i])
			// go to the method's blocks
			funcName := instruction.GetName()
			fn := pg.GetFunc(funcName)
			// TODO: check the argument and parameter
			tempFunc := pg.CurrFunc
			pg.CurrFunc = fn
			pg.TraverseBlock(fn.Cfg)
			pg.CurrFunc = tempFunc
		case *cfg.OpExprStaticCall:
			pg.CurrPath.AddInstruction(block.Instructions[i])
			// go to the static method's blocks
			funcName := instruction.GetName()
			fn := pg.GetFunc(funcName)
			// TODO: check the argument and parameter
			tempFunc := pg.CurrFunc
			pg.CurrFunc = fn
			pg.TraverseBlock(fn.Cfg)
			pg.CurrFunc = tempFunc
		case *cfg.OpExprAssign, *cfg.OpExprAssignRef:
			pg.CurrPath.AddInstruction(block.Instructions[i])
		}
	}
	lastInstruction := block.Instructions[len(block.Instructions)-1]
	switch i := lastInstruction.(type) {
	case *cfg.OpStmtJumpIf:
		condition := NewCondi
		condOp := i.Cond.GetWriteOp()[0]

		// go to conditional block
		ifPath := pg.CurrPath.Clone()
		elsePath := pg.CurrPath.Clone()
	case *cfg.OpStmtSwitch:
		// go to conditional block
	case *cfg.OpStmtJump:
		pg.TraverseBlock(i.Target)
	case *cfg.OpReturn:
		// if return in main script, exit block
		// else, just back to the caller
		if pg.CurrFunc.GetScopedName() == "{main}" {
			pg.Path = append(pg.Path, pg.CurrPath)
		}
	default:
		log.Fatalf("Error: invalid instruction '%v' as last instruction", reflect.TypeOf(i))
	}
	// case
	block.Visited = false
}

func (pg *PathGenerator) AddPath(path PathCtx) {
	// check conditions using z3 solver
	pg.
}

func (pg *PathGenerator) GetFunc(name string) *cfg.OpFunc {
	// find in the current script
	if fn, ok := pg.CurrScript.Funcs[name]; ok {
		return fn
	}

	// find in the script's included file
	for _, includedFile := range pg.CurrScript.IncludedFiles {
		if script, ok := pg.Scripts[includedFile]; ok {
			if fn, ok := script.Funcs[name]; ok {
				return fn
			}
		}
	}

	// function not founded
	return nil
}

func OpToCondition(op cfg.Op) condition.Condition {

}

func IsSource(op cfg.Op) bool {
	// php source
	if assignOp, ok := op.(*cfg.OpExprAssign); ok {
		// symbolic interpreter ($_POST, $_GET, $_REQUEST, $_FILES, $_COOKIE, $_SERVERS)
		if result, ok := assignOp.Result.(*cfg.OperSymbolic); ok {
			switch result.Val {
			case "postsymbolic":
				fallthrough
			case "getsymbolic":
				fallthrough
			case "requestsymbolic":
				fallthrough
			case "filessymbolic":
				fallthrough
			case "cookiesymbolic":
				fallthrough
			case "serverssymbolic":
				return true
			}
		}
		// filter_input(), apache_request_headers(), getallheaders()
		if assignOp.Expr.IsWritten() {
			if right, ok := assignOp.Expr.GetWriteOp()[0].(*cfg.OpExprFunctionCall); ok {
				funcNameStr := cfg.GetOperName(right.Name)
				switch funcNameStr {
				case "filter_input":
					// TODO: check again the arguments
					return true
				case "apache_request_headers":
					fallthrough
				case "getallheaders":
					return true
				}
			}
		}
	}

	// TODO: laravel source

	return false
}
