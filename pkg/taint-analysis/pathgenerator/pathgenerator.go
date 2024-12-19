package pathgenerator

import (
	"fmt"
	"log"
	"reflect"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg"
	"github.com/aclements/go-z3/z3"
)

type PathGenerator struct {
	Scripts    map[string]*cfg.Script
	CurrScript *cfg.Script
	CurrPath   *ExecPath
	CurrFunc   *cfg.OpFunc
	VarIds     map[cfg.Operand]int

	// TODO: check again
	CurrVar       int // helper to create name for variable used for z3 solver in the next
	FeasiblePaths []*ExecPath
}

// execution path
type ExecPath struct {
	Instructions []cfg.Op                    // first item for source, last item for sink
	Conds        []cfg.Operand               //
	Vars         map[cfg.Operand]struct{}    // set of var defined in this path contex, can be helper to choose phi value
	ReplacedVars map[cfg.Operand]cfg.Operand // choosen phi var or param function based on this path

	// TODO: check again if not used
	CurrReturnVal cfg.Operand // helper to handle function return
}

func NewExecPath() *ExecPath {
	return &ExecPath{
		Instructions: make([]cfg.Op, 0),
		Conds:        make([]cfg.Operand, 0),
		Vars:         make(map[cfg.Operand]struct{}),
		ReplacedVars: make(map[cfg.Operand]cfg.Operand),
	}
}

// add instruction with path specific variable
func (p *ExecPath) AddInstruction(instruction cfg.Op) {
	// if one of the used var replaced, create new instruction specific for this path
	useReplacedVar := false
	cloneIntr := instruction.Clone()
	for varName, vr := range cloneIntr.GetOpVars() {
		if replacedVar, ok := p.ReplacedVars[vr]; ok {
			useReplacedVar = true
			cloneIntr.ChangeOpVar(varName, replacedVar)
		}
	}
	for _, varList := range cloneIntr.GetOpListVars() {
		for idx, vr := range varList {
			if replacedVar, ok := p.ReplacedVars[vr]; ok {
				useReplacedVar = true
				varList[idx] = replacedVar
			}
		}
	}

	if useReplacedVar {

	} else {
		p.Instructions = append(p.Instructions, instruction)
	}
}

func (p *ExecPath) AddCondition(cond cfg.Operand) {
	p.Conds = append(p.Conds, cond)
}

func (p *ExecPath) AddVar(oper cfg.Operand) {
	p.Vars[oper] = struct{}{}
}

func (p *ExecPath) ReplaceVar(from, to cfg.Operand) {
	p.ReplacedVars[from] = to
}

func (p *ExecPath) GetVar(vr cfg.Operand) cfg.Operand {
	if replacedVr, ok := p.ReplacedVars[vr]; ok {
		return replacedVr
	}
	return vr
}

func (p *ExecPath) Clone() *ExecPath {
	copiedInstructions := make([]cfg.Op, len(p.Instructions))
	copy(copiedInstructions, p.Instructions)
	copiedConds := make([]cfg.Operand, len(p.Conds))
	copy(copiedConds, p.Conds)
	return &ExecPath{
		Instructions: copiedInstructions,
		Conds:        copiedConds,
		Vars:         p.Vars,
		ReplacedVars: p.ReplacedVars,
	}
}

func GenerateFeasiblePath(scripts map[string]*cfg.Script) []*ExecPath {
	generator := &PathGenerator{
		FeasiblePaths: make([]*ExecPath, 0),
		VarIds:        make(map[cfg.Operand]int),
	}

	// Generate path if script's Main contain tainted data
	for _, script := range scripts {
		if script.Main.ContaintTainted {
			// TODO: search sources as begining path
			// generate path with script.Main as entry block

			generator.CurrPath = NewExecPath()
			generator.CurrScript = script
			generator.CurrFunc = script.Main
			generator.TraverseBlock(script.Main.Cfg)
		}
	}

	return generator.FeasiblePaths
}

func (pg *PathGenerator) TraverseBlock(block *cfg.Block) {
	if block == nil || len(block.Instructions) <= 0 || block.Visited {
		return
	}
	block.Visited = true
	for i := 0; i < len(block.Instructions)-1; i++ {
		// create
		pg.CurrPath.AddInstruction(block.Instructions[i])
		// find if there is a func call
		switch intr := block.Instructions[i].(type) {
		case *cfg.OpExprFunctionCall:
			// go to the function's blocks
			funcName := intr.GetName()
			fn := pg.GetFunc(funcName)
			// TODO: check the argument and parameter
			idx := 0
			for ; idx < len(intr.Args); idx++ {
				arg := intr.Args[idx]
				param := fn.Params[idx]
				pg.CurrPath.ReplaceVar(param.Result, arg)
			}
			for ; idx < len(fn.Params); idx++ {
				param := fn.Params[idx]
				if param.DefaultVar != nil {
					pg.CurrPath.ReplaceVar(param.Result, param.DefaultVar)
				} else {
					pg.CurrPath.ReplaceVar(param.Result, cfg.NewOperNull())
				}
			}

			tempFunc := pg.CurrFunc
			pg.CurrFunc = fn
			pg.TraverseBlock(fn.Cfg)
			pg.CurrFunc = tempFunc

			// add return value
			if pg.CurrPath.CurrReturnVal != nil {
				pg.CurrPath.ReplaceVar(intr.Result, pg.CurrPath.CurrReturnVal)
				pg.CurrPath.CurrReturnVal = nil
			}
		case *cfg.OpExprMethodCall:
			// go to the method's blocks
			funcName := intr.GetName()
			fn := pg.GetFunc(funcName)
			// TODO: check the argument and parameter
			idx := 0
			for ; idx < len(intr.Args); idx++ {
				arg := intr.Args[idx]
				param := fn.Params[idx]
				pg.CurrPath.ReplaceVar(param.Result, arg)
			}
			for ; idx < len(fn.Params); idx++ {
				param := fn.Params[idx]
				if param.DefaultVar != nil {
					pg.CurrPath.ReplaceVar(param.Result, param.DefaultVar)
				} else {
					pg.CurrPath.ReplaceVar(param.Result, cfg.NewOperNull())
				}
			}

			tempFunc := pg.CurrFunc
			pg.CurrFunc = fn
			pg.TraverseBlock(fn.Cfg)
			pg.CurrFunc = tempFunc

			// add return value
			if pg.CurrPath.CurrReturnVal != nil {
				pg.CurrPath.ReplaceVar(intr.Result, pg.CurrPath.CurrReturnVal)
				pg.CurrPath.CurrReturnVal = nil
			}
		case *cfg.OpExprStaticCall:
			// go to the static method's blocks
			funcName := intr.GetName()
			fn := pg.GetFunc(funcName)
			// TODO: check the argument and parameter
			idx := 0
			for ; idx < len(intr.Args); idx++ {
				arg := intr.Args[idx]
				param := fn.Params[idx]
				pg.CurrPath.ReplaceVar(param.Result, arg)
			}
			for ; idx < len(fn.Params); idx++ {
				param := fn.Params[idx]
				if param.DefaultVar != nil {
					pg.CurrPath.ReplaceVar(param.Result, param.DefaultVar)
				} else {
					pg.CurrPath.ReplaceVar(param.Result, cfg.NewOperNull())
				}
			}

			tempFunc := pg.CurrFunc
			pg.CurrFunc = fn
			pg.TraverseBlock(fn.Cfg)
			pg.CurrFunc = tempFunc

			// add return value
			if pg.CurrPath.CurrReturnVal != nil {
				pg.CurrPath.ReplaceVar(intr.Result, pg.CurrPath.CurrReturnVal)
				pg.CurrPath.CurrReturnVal = nil
			}
		case *cfg.OpExprAssign:
			// define variable to the path context
			pg.CurrPath.AddVar(intr.Var)
		case *cfg.OpExprAssignRef:
			pg.CurrPath.AddVar(intr.Var)
		case *cfg.OpPhi:
			// find var which defined in the current path
			for vr := range intr.Vars {
				if _, ok := pg.CurrPath.Vars[vr]; ok {
					// replace phi result to var
					pg.CurrPath.ReplaceVar(intr.Result, vr)
				}
			}
		}
	}
	lastInstruction := block.Instructions[len(block.Instructions)-1]
	switch intr := lastInstruction.(type) {
	case *cfg.OpStmtJumpIf:
		cond := intr.Cond
		negatedCond := cfg.NewOpExprBooleanNot(cond, nil).Result
		condVal := cfg.GetOperVal(intr.Cond)

		if cvBool, ok := condVal.(*cfg.OperBool); ok {
			// condition is boolean, traverse one of the block
			newPath := pg.CurrPath.Clone()
			tmp := pg.CurrPath
			pg.CurrPath = newPath
			if cvBool.Val {
				// traverse if block
				newPath.AddCondition(cond)
				pg.TraverseBlock(intr.If)
			} else {
				// traverse else block
				newPath.AddCondition(negatedCond)
				pg.TraverseBlock(intr.Else)
			}
			pg.CurrPath = tmp
		} else {
			tmp := pg.CurrPath
			// else, traverse both
			ifPath := pg.CurrPath.Clone()
			elsePath := pg.CurrPath.Clone()

			// traverse if block first
			ifPath.AddCondition(cond)
			pg.CurrPath = ifPath
			pg.TraverseBlock(intr.If)

			// then, traverse else block
			elsePath.AddCondition(negatedCond)
			pg.CurrPath = elsePath
			pg.TraverseBlock(intr.Else)

			pg.CurrPath = tmp
		}
	case *cfg.OpStmtSwitch:
		// go to conditional block
		for i, cs := range intr.Cases {
			tmp := pg.CurrPath
			// traverse to each condition block
			cond := cfg.NewOpExprBinaryEqual(intr.Cond, cs, nil).Result
			newPath := pg.CurrPath.Clone()
			pg.CurrPath = newPath
			newPath.AddCondition(cond)
			pg.TraverseBlock(intr.Targets[i])
			pg.CurrPath = tmp
		}
	case *cfg.OpStmtJump:
		pg.TraverseBlock(intr.Target)
	case *cfg.OpReturn:
		// if return in main script, exit block
		// else, just back to the caller
		if pg.CurrFunc.GetScopedName() == "{main}" {
			pg.AddCurrPath()
		} else {
			// TODO: handle function return value
			pg.CurrPath.CurrReturnVal = intr.Expr
		}
	default:
		log.Fatalf("Error: invalid instruction '%v' as last instruction", reflect.TypeOf(intr))
	}
	block.Visited = false
}

func (pg *PathGenerator) AddCurrPath() {
	ctx := z3.NewContext(nil)
	solver := z3.NewSolver(ctx)
	// check conditions using z3 solver
	fmt.Println("Check condition")
	for i, cond := range pg.CurrPath.Conds {
		constraint, _ := pg.ExtractConstraints(ctx, cond)
		solver.Assert(constraint)
		isSatisfiable, err := solver.Check()
		if err != nil {
			log.Fatal("error: fail to execute z3 solver")
		}
		// if not satisfy, path not added
		if !isSatisfiable {
			fmt.Println("End condition")
			return
		}
		fmt.Println(i)
	}
	fmt.Println("End condition")
	// all condition satisfiable,
	pg.FeasiblePaths = append(pg.FeasiblePaths, pg.CurrPath)
}

func (pg *PathGenerator) ExtractConstraints(ctx *z3.Context, oper cfg.Operand) (z3.Bool, bool) {
	// get the var definition specific to the current path
	oper = pg.CurrPath.GetVar(oper)

	// check if operand is scalar
	switch operT := oper.(type) {
	case *cfg.OperNumber:
		if operT.Val == 0 {
			return ctx.FromBool(false), true
		} else {
			return ctx.FromBool(true), true
		}
	case *cfg.OperBool:
		return ctx.FromBool(operT.Val), true
	case *cfg.OperString:
		if operT.Val == "" || operT.Val == "0" {
			return ctx.FromBool(false), true
		} else {
			return ctx.FromBool(true), true
		}
	}
	if len(oper.GetWriteOp()) > 0 {
		switch op := oper.GetWriteOp()[0].(type) {
		case *cfg.OpExprBinaryEqual:
			// for now, handle equal similar to identical
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			switch leftVal := cfg.GetOperVal(left).(type) {
			case *cfg.OperBool:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					if rightVal.Val == leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromBool(leftVal.Val)
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.BoolSort()).(z3.Bool)
					return leftSort.Eq(rightSort), true
				}
			case *cfg.OperNumber:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperNumber:
					if rightVal.Val == leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromFloat64(leftVal.Val, ctx.FloatSort(11, 53))
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.FloatSort(11, 53)).(z3.Float)
					return leftSort.Eq(rightSort), true
				}
			case *cfg.OperString:
				if rightVal, ok := cfg.GetOperVal(right).(*cfg.OperString); ok {
					if rightVal.Val == leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				}
			case *cfg.OperSymbolic:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.BoolSort()).(z3.Bool)
					rightSort := ctx.FromBool(rightVal.Val)
					return leftSort.Eq(rightSort), true
				case *cfg.OperNumber:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.FloatSort(11, 53)).(z3.Float)
					rightSort := ctx.FromFloat64(rightVal.Val, ctx.FloatSort(11, 53))
					return leftSort.Eq(rightSort), true
				}
			}
			// evaluate arithmetic
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.Eq(rightArith), true
			}
		case *cfg.OpExprBinaryNotEqual:
			// for now, handle equal similar to identical
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			switch leftVal := cfg.GetOperVal(left).(type) {
			case *cfg.OperBool:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					if rightVal.Val != leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromBool(leftVal.Val)
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.BoolSort()).(z3.Bool)
					return leftSort.Eq(rightSort), true
				}
			case *cfg.OperNumber:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperNumber:
					if rightVal.Val != leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromFloat64(leftVal.Val, ctx.FloatSort(11, 53))
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.FloatSort(11, 53)).(z3.Float)
					return leftSort.NE(rightSort), true
				}
			case *cfg.OperString:
				if rightVal, ok := cfg.GetOperVal(right).(*cfg.OperString); ok {
					if rightVal.Val != leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				}
			case *cfg.OperSymbolic:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.BoolSort()).(z3.Bool)
					rightSort := ctx.FromBool(rightVal.Val)
					return leftSort.NE(rightSort), true
				case *cfg.OperNumber:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.FloatSort(11, 53)).(z3.Float)
					rightSort := ctx.FromFloat64(rightVal.Val, ctx.FloatSort(11, 53))
					return leftSort.NE(rightSort), true
				}
			}
			// evaluate arithmetic
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.NE(rightArith), true
			}
		case *cfg.OpExprBinaryIdentical:
			// for now, handle equal similar to identical
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			switch leftVal := cfg.GetOperVal(left).(type) {
			case *cfg.OperBool:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					if rightVal.Val == leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromBool(leftVal.Val)
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.BoolSort()).(z3.Bool)
					return leftSort.Eq(rightSort), true
				case *cfg.OperString, *cfg.OperNumber:
					return ctx.FromBool(false), true
				}
			case *cfg.OperNumber:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperNumber:
					if rightVal.Val == leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromFloat64(leftVal.Val, ctx.FloatSort(11, 53))
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.FloatSort(11, 53)).(z3.Float)
					return leftSort.Eq(rightSort), true
				case *cfg.OperString, *cfg.OperBool:
					return ctx.FromBool(false), true
				}
			case *cfg.OperString:
				switch rightVal := right.(type) {
				case *cfg.OperString:
					if rightVal.Val == leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperBool, *cfg.OperNumber:
					return ctx.FromBool(false), true
				}
			case *cfg.OperSymbolic:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.BoolSort()).(z3.Bool)
					rightSort := ctx.FromBool(rightVal.Val)
					return leftSort.Eq(rightSort), true
				case *cfg.OperNumber:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.FloatSort(11, 53)).(z3.Float)
					rightSort := ctx.FromFloat64(rightVal.Val, ctx.FloatSort(11, 53))
					return leftSort.Eq(rightSort), true
				}
			}
			// evaluate arithmetic
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.Eq(rightArith), true
			} else if (isLeftArith && !isRightArith) || (!isLeftArith && isRightArith) {
				return ctx.FromBool(false), true
			}
		case *cfg.OpExprBinaryNotIdentical:
			// for now, handle equal similar to identical
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			switch leftVal := cfg.GetOperVal(left).(type) {
			case *cfg.OperBool:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					if rightVal.Val != leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromBool(leftVal.Val)
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.BoolSort()).(z3.Bool)
					return leftSort.NE(rightSort), true
				case *cfg.OperString, *cfg.OperNumber:
					return ctx.FromBool(true), true
				}
			case *cfg.OperNumber:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperNumber:
					if rightVal.Val != leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperSymbolic:
					leftSort := ctx.FromFloat64(leftVal.Val, ctx.FloatSort(11, 53))
					rightName := pg.GetVarName(right)
					rightSort := ctx.Const(rightName, ctx.FloatSort(11, 53)).(z3.Float)
					return leftSort.NE(rightSort), true
				case *cfg.OperString, *cfg.OperBool:
					return ctx.FromBool(true), true
				}
			case *cfg.OperString:
				switch rightVal := right.(type) {
				case *cfg.OperString:
					if rightVal.Val != leftVal.Val {
						return ctx.FromBool(true), true
					} else {
						return ctx.FromBool(false), true
					}
				case *cfg.OperBool, *cfg.OperNumber:
					return ctx.FromBool(true), true
				}
			case *cfg.OperSymbolic:
				switch rightVal := cfg.GetOperVal(right).(type) {
				case *cfg.OperBool:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.BoolSort()).(z3.Bool)
					rightSort := ctx.FromBool(rightVal.Val)
					return leftSort.NE(rightSort), true
				case *cfg.OperNumber:
					leftName := pg.GetVarName(left)
					leftSort := ctx.Const(leftName, ctx.FloatSort(11, 53)).(z3.Float)
					rightSort := ctx.FromFloat64(rightVal.Val, ctx.FloatSort(11, 53))
					return leftSort.NE(rightSort), true
				}
			}
			// evaluate arithmetic
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.NE(rightArith), true
			} else if (isLeftArith && !isRightArith) || (!isLeftArith && isRightArith) {
				return ctx.FromBool(true), true
			}
		case *cfg.OpExprBinaryGreater:
			// for now, just handle numeric
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.GT(rightArith), true
			}
		case *cfg.OpExprBinaryGreaterOrEqual:
			// for now, just handle numeric
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.GE(rightArith), true
			}
		case *cfg.OpExprBinarySmaller:
			// for now, just handle numeric
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.LT(rightArith), true
			}
		case *cfg.OpExprBinarySmallerOrEqual:
			// for now, just handle numeric
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftArith, isLeftArith := pg.EvaluateArithmetic(ctx, left)
			rightArith, isRightArith := pg.EvaluateArithmetic(ctx, right)
			if isLeftArith && isRightArith {
				return leftArith.LE(rightArith), true
			}
		case *cfg.OpExprBinaryLogicalAnd:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftConstraint, _ := pg.ExtractConstraints(ctx, left)
			rightConstraint, _ := pg.ExtractConstraints(ctx, right)
			return leftConstraint.And(rightConstraint), true
		case *cfg.OpExprBinaryLogicalOr:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftConstraint, _ := pg.ExtractConstraints(ctx, left)
			rightConstraint, _ := pg.ExtractConstraints(ctx, right)
			return leftConstraint.Or(rightConstraint), true
		case *cfg.OpExprBinaryLogicalXor:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftConstraint, _ := pg.ExtractConstraints(ctx, left)
			rightConstraint, _ := pg.ExtractConstraints(ctx, right)
			return leftConstraint.Xor(rightConstraint), true
		case *cfg.OpExprBooleanNot:
			expr := pg.CurrPath.GetVar(op.Expr)
			exprConstraint, isDef := pg.ExtractConstraints(ctx, expr)
			if isDef {
				return exprConstraint.Not(), true
			} else {
				return ctx.FromBool(true), false
			}
		}
	}

	// for other expression, just return true
	return ctx.FromBool(true), false
}

func (pg *PathGenerator) EvaluateArithmetic(ctx *z3.Context, oper cfg.Operand) (z3.Float, bool) {
	floatZero := ctx.FloatZero(ctx.FloatSort(11, 53), true)
	floatSort := ctx.FloatSort(11, 53)
	// check if operand's value is scalar
	oper = pg.CurrPath.GetVar(oper)
	operValue := cfg.GetOperVal(oper)
	switch o := operValue.(type) {
	case *cfg.OperNumber:
		return ctx.FromFloat64(o.Val, floatSort), true
	case *cfg.OperBool, *cfg.OperString:
		return floatZero, false
	}
	if len(oper.GetWriteOp()) > 0 {
		op := oper.GetWriteOp()[0]
		switch op := op.(type) {
		case *cfg.OpExprBinaryPlus:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftVal, isLeftNum := pg.EvaluateArithmetic(ctx, left)
			rightVal, isRightNum := pg.EvaluateArithmetic(ctx, right)
			if !isLeftNum {
				leftVal = ctx.Const(pg.GetVarName(left), floatSort).(z3.Float)
			}
			if !isRightNum {
				rightVal = ctx.Const(pg.GetVarName(right), floatSort).(z3.Float)
			}
			return leftVal.Add(rightVal), true
		case *cfg.OpExprBinaryMinus:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftVal, isLeftNum := pg.EvaluateArithmetic(ctx, left)
			rightVal, isRightNum := pg.EvaluateArithmetic(ctx, right)
			if !isLeftNum {
				leftVal = ctx.Const(pg.GetVarName(left), floatSort).(z3.Float)
			}
			if !isRightNum {
				rightVal = ctx.Const(pg.GetVarName(right), floatSort).(z3.Float)
			}
			return leftVal.Sub(rightVal), true
		case *cfg.OpExprBinaryDiv:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftVal, isLeftNum := pg.EvaluateArithmetic(ctx, left)
			rightVal, isRightNum := pg.EvaluateArithmetic(ctx, right)
			if !isLeftNum {
				leftVal = ctx.Const(pg.GetVarName(left), floatSort).(z3.Float)
			}
			if !isRightNum {
				rightVal = ctx.Const(pg.GetVarName(right), floatSort).(z3.Float)
			}
			return leftVal.Div(rightVal), true
		case *cfg.OpExprBinaryMul:
			left := pg.CurrPath.GetVar(op.Left)
			right := pg.CurrPath.GetVar(op.Right)
			leftVal, isLeftNum := pg.EvaluateArithmetic(ctx, left)
			rightVal, isRightNum := pg.EvaluateArithmetic(ctx, right)
			if !isLeftNum {
				leftVal = ctx.Const(pg.GetVarName(left), floatSort).(z3.Float)
			}
			if !isRightNum {
				rightVal = ctx.Const(pg.GetVarName(right), floatSort).(z3.Float)
			}
			return leftVal.Mul(rightVal), true
		}
	}

	return ctx.Const(pg.GetVarName(oper), floatSort).(z3.Float), true
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

func (pg *PathGenerator) GetVarName(oper cfg.Operand) string {
	if id, ok := pg.VarIds[oper]; ok {
		return fmt.Sprintf("v%d", id)
	}
	pg.VarIds[oper] = pg.CurrVar
	pg.CurrVar += 1
	return fmt.Sprintf("v%d", pg.VarIds[oper])
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
