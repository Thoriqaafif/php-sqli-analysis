package optimizer

import (
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg/traverser"
)

type Optimizer struct {
	traverser.NullTraverser
}

func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

func (t *Optimizer) EnterOp(op cfg.Op, block *cfg.Block) {
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
	case *cfg.OpExprBinaryPow:
		// TODO:
	case *cfg.OpExprBinaryLogicalAnd, *cfg.OpExprBinaryLogicalOr, *cfg.OpExprBinaryLogicalXor:
		// TODO:
	case *cfg.OpExprBinaryBooleanAnd, *cfg.OpExprBinaryBooleanOr:
		// TODO:
	case *cfg.OpExprBinaryBitwiseAnd, *cfg.OpExprBinaryBitwiseOr, *cfg.OpExprBinaryBitwiseXor:
		// TODO:
	case *cfg.OpExprBinaryCoalesce, *cfg.OpExprBinarySpaceship, *cfg.OpExprBinaryConcat:
		// DO Nothing
	}
}

func ReplaceOpVar(from, to cfg.Operand) {
	for _, usage := range from.GetUsage() {
		for vrName, vr := range usage.GetOpVars() {
			if vr == from {
				usage.ChangeOpVar(vrName, to)
			}
		}
	}
}
