package optimizer

import (
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
)

type Optimizer struct {
	traverser.NullTraverser

	FilePath string
}

func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

func (t *Optimizer) LeaveBlock(block *cfg.Block, _ *cfg.Block) {
	// set block for each op
	for _, op := range block.Instructions {
		op.SetBlock(block)
	}
}

func (t *Optimizer) EnterScript(script *cfg.Script) {
	t.FilePath = script.FilePath
}

func (t *Optimizer) EnterOp(op cfg.Op, block *cfg.Block) {
	op.SetFilePath(t.FilePath)
	switch o := op.(type) {
	case *cfg.OpExprBinaryDiv:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(leftVal.Val / rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryMinus:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(leftVal.Val - rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryPlus:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(leftVal.Val + rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryMul:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(leftVal.Val * rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryMod:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(float64(int(leftVal.Val) % int(rightVal.Val)))
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryEqual:
		// left and right must have same type
		switch leftVal := cfg.GetOperVal(o.Left).(type) {
		case *cfg.OperBool:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperBool)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val == rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperNumber:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val == rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperString:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperString)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val == rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		}
	case *cfg.OpExprBinaryNotEqual:
		// left and right must have same type
		switch leftVal := cfg.GetOperVal(o.Left).(type) {
		case *cfg.OperBool:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperBool)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val != rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperNumber:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val != rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperString:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperString)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val != rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		}
	case *cfg.OpExprBinaryIdentical:
		// left and right must have same type
		switch leftVal := cfg.GetOperVal(o.Left).(type) {
		case *cfg.OperBool:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperBool)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val == rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperNumber:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val == rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperString:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperString)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val == rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		}
	case *cfg.OpExprBinaryNotIdentical:
		// left and right must have same type
		switch leftVal := cfg.GetOperVal(o.Left).(type) {
		case *cfg.OperBool:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperBool)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val != rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperNumber:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val != rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		case *cfg.OperString:
			rightVal, ok := cfg.GetOperVal(o.Right).(*cfg.OperString)
			if ok {
				newOp := cfg.NewOperBool(leftVal.Val != rightVal.Val)
				ReplaceOpVar(o.Result, newOp)
				o.Result = newOp
			}
		}
	case *cfg.OpExprBinaryGreater:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperBool(leftVal.Val > rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryGreaterOrEqual:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperBool(leftVal.Val >= rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinarySmaller:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperBool(leftVal.Val < rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinarySmallerOrEqual:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperBool(leftVal.Val <= rightVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryShiftLeft:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(float64(int(leftVal.Val) << int(rightVal.Val)))
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBinaryShiftRight:
		// both left and right must be a number
		leftVal, isNumLeft := cfg.GetOperVal(o.Left).(*cfg.OperNumber)
		rightVal, isNumRight := cfg.GetOperVal(o.Right).(*cfg.OperNumber)
		if isNumLeft && isNumRight {
			newOp := cfg.NewOperNumber(float64(int(leftVal.Val) >> int(rightVal.Val)))
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprUnaryMinus:
		exprVal, isNum := cfg.GetOperVal(o.Expr).(*cfg.OperNumber)
		if isNum {
			newOp := cfg.NewOperNumber(-1 * exprVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprUnaryPlus:
		exprVal, isNum := cfg.GetOperVal(o.Expr).(*cfg.OperNumber)
		if isNum {
			newOp := cfg.NewOperNumber(exprVal.Val)
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprBitwiseNot:
		exprVal, isNum := cfg.GetOperVal(o.Expr).(*cfg.OperNumber)
		if isNum {
			newOp := cfg.NewOperNumber(float64(^int(exprVal.Val)))
			ReplaceOpVar(o.Result, newOp)
			o.Result = newOp
		}
	case *cfg.OpExprAssign:
		if cfg.IsScalarOper(o.Expr) {
			ReplaceOpVar(o.Result, o.Expr)
			o.Result = o.Expr
			cfg.SetOperVal(o.Var, o.Expr)
		}
	case *cfg.OpExprBinaryConcat:
		// TODO: concat string if the value of left and right are scalar
	case *cfg.OpExprConcatList:
		// TODO: concat string if the value of each operand in lists are scalar
	case *cfg.OpExprBinaryBitwiseAnd, *cfg.OpExprBinaryBitwiseOr, *cfg.OpExprBinaryBitwiseXor:
		// TODO:
	case *cfg.OpExprBinaryPow:
		// TODO:
	case *cfg.OpExprBinaryLogicalAnd, *cfg.OpExprBinaryLogicalOr, *cfg.OpExprBinaryLogicalXor:
		// TODO:
	case *cfg.OpExprBinaryCoalesce, *cfg.OpExprBinarySpaceship:
		// DO Nothing
	}
}

func ReplaceOpVar(from, to cfg.Operand) {
	for _, usage := range from.GetUsage() {
		if negationUsage, ok := usage.(*cfg.OpExprBooleanNot); ok {
			if boolTo, ok := to.(*cfg.OperBool); ok {
				negationUsage.Expr = boolTo
				newOp := cfg.NewOperBool(!boolTo.Val)
				ReplaceOpVar(negationUsage.Result, newOp)
				negationUsage.Result = newOp
			}
			continue
		} else if phiUsage, ok := usage.(*cfg.OpPhi); ok {
			phiUsage.RemoveOperand(from)
			phiUsage.AddOperand(to)
		}
		for vrName, vr := range usage.GetOpVars() {
			if vr == from {
				usage.ChangeOpVar(vrName, to)
				to.AddUsage(usage)
				from.RemoveUsage(usage)
			}
		}
		for _, varList := range usage.GetOpListVars() {
			for i, vr := range varList {
				if vr == from {
					varList[i] = to
					to.AddUsage(usage)
					from.RemoveUsage(usage)
				}
			}
		}
	}
	for _, block := range from.GetCondUsages() {
		for i, cond := range block.Conds {
			if cond == from {
				block.Conds[i] = to
			}
		}
	}
}
