package taintutil

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
)

func IsSource(op cfg.Op) bool {
	// php source
	switch opT := op.(type) {
	case *cfg.OpExprAssign:
		if right, ok := opT.Expr.(*cfg.OperSymbolic); ok {
			switch right.Val {
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
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "filter_input_array":
			// TODO: check again the arguments
			return true
		case "filter_input":
			// TODO: check again the arguments
			return true
		case "apache_request_headers":
			fallthrough
		case "getallheaders":
			return true
		}
	case *cfg.OpReset:
		return false
	default:
		for _, vr := range op.GetOpVars() {
			if vr, ok := vr.(*cfg.OperSymbolic); ok {
				switch vr.Val {
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
		}
	}

	// TODO: laravel source

	return false
}

func IsPropagated(op cfg.Op) bool {
	// for sanitizer, data cannot be tainted
	switch opT := op.(type) {
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "mysql_real_escape_string":
			fallthrough
		case "mysql_escape_string":
			fallthrough
		case "mysqli_real_escape_string":
			fallthrough
		case "pg_escape_string":
			fallthrough
		case "pg_escape_literal":
			fallthrough
		case "pg_escape_identifier":
			fallthrough
		case "intval":
			fallthrough
		case "floatval":
			fallthrough
		case "boolval":
			fallthrough
		case "doubleval":
			return false
		case "preg_match":
			arg0 := opT.Args[0]
			if arg0Str, ok := arg0.(*cfg.OperString); ok && arg0Str.Val == "/^[0-9]*$/" {
				return false
			}
		}
	case *cfg.OpExprMethodCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "escape_string":
			fallthrough
		case "quote":
			return false
		}
	case *cfg.OpExprStaticCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "escape_string":
			fallthrough
		case "quote":
			return false
		}
	case *cfg.OpExprCastBool, *cfg.OpExprCastDouble, *cfg.OpExprCastInt, *cfg.OpExprCastUnset, *cfg.OpUnset:
		return false
	case *cfg.OpExprAssertion:
		switch assert := opT.Assertion.(type) {
		case *cfg.TypeAssertion:
			if typeVal, ok := assert.Val.(*cfg.OperString); ok {
				switch typeVal.Val {
				case "int", "float", "bool", "null":
					return false
				}
			}
		}
	case *cfg.OpStmtJumpIf, *cfg.OpExprBinaryPlus, *cfg.OpExprBinaryMinus, *cfg.OpExprBinaryMod,
		*cfg.OpExprBinaryDiv, *cfg.OpExprBinaryBitwiseAnd, *cfg.OpExprBinaryBitwiseOr, *cfg.OpExprBinaryBitwiseXor,
		*cfg.OpExprEmpty, *cfg.OpExprBinaryEqual, *cfg.OpExprBinaryGreater, *cfg.OpExprBinaryGreaterOrEqual,
		*cfg.OpExprBinaryIdentical, *cfg.OpExprBinaryLogicalAnd, *cfg.OpExprBinaryLogicalOr, *cfg.OpExprBinaryLogicalXor,
		*cfg.OpExprBinaryMul, *cfg.OpExprBinaryNotEqual, *cfg.OpExprBinaryNotIdentical, *cfg.OpExprBinaryPow,
		*cfg.OpExprBinaryShiftLeft, *cfg.OpExprBinaryShiftRight, *cfg.OpExprBinarySmaller, *cfg.OpExprBinarySmallerOrEqual,
		*cfg.OpExprBitwiseNot, *cfg.OpReset, *cfg.OpExit, *cfg.OpExprIsset, *cfg.OpStmtSwitch, *cfg.OpEcho:
		return false
	}

	return true
}

func IsSink(op cfg.Op) bool {
	// php sink
	switch opT := op.(type) {
	case *cfg.OpExprFunctionCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "mysql_query":
			fallthrough
		case "mysql_db_query":
			fallthrough
		case "mysqli_query":
			fallthrough
		case "mysqli_multi_query":
			fallthrough
		case "mysqli_real_query":
			fallthrough
		case "mysqli_execute":
			fallthrough
		case "mysqli_prepare":
			fallthrough
		case "pg_query":
			fallthrough
		case "pg_send_query":
			return true
		}
	case *cfg.OpExprMethodCall:
		methodNameStr, _ := cfg.GetOperName(opT.Name)
		switch methodNameStr {
		case "direct_query":
			fallthrough
		case "query":
			fallthrough
		case "multi_query":
			fallthrough
		case "prepare":
			return true
		}
		if strings.HasSuffix(methodNameStr, "query") {
			return true
		}
	case *cfg.OpExprStaticCall:
		methodNameStr, _ := cfg.GetOperName(opT.Name)
		switch methodNameStr {
		case "direct_query":
			fallthrough
		case "query":
			fallthrough
		case "multi_query":
			fallthrough
		case "prepare":
			return true
		}
		if strings.HasSuffix(methodNameStr, "query") {
			return true
		}
	}

	// TODO: laravel sink

	return false
}

func GetTaintedVar(op cfg.Op) (cfg.Operand, error) {
	if assignOp, ok := op.(*cfg.OpExprAssign); ok {
		if assignOp.Var != nil {
			return assignOp.Var, nil
		} else {
			return assignOp.Result, nil
		}
	} else if result, ok := op.GetOpVars()["Result"]; ok {
		if result != nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("wrong tainted var '%v'", reflect.TypeOf(op))
}
