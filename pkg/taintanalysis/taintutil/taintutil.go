package taintutil

import (
	"fmt"
	"log"
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
			if len(opT.Args) == 1 {
				return true
			} else {
				filter := opT.Args[1].GetWriteOp()
				switch filterOp := filter.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperName(filterOp.Name)
					if err != nil {
						log.Fatalf("error in IsSource: %v", err)
					}
					if len(opT.Args) == 2 {
						switch constName {
						case "FILTER_SANITIZE_NUMBER_INT":
							fallthrough
						case "FILTER_SANITIZE_NUMBER_FLOAT":
							return false
						default:
							return true
						}
					}
				case *cfg.OpExprArray:
					for _, arg := range filterOp.Vals {
						arrFilter := arg.GetWriteOp()
						switch filterOp := arrFilter.(type) {
						case *cfg.OpExprConstFetch:
							constName, err := cfg.GetOperName(filterOp.Name)
							if err != nil {
								log.Fatalf("error in IsSource: %v", err)
							}
							if len(opT.Args) == 2 {
								switch constName {
								case "FILTER_SANITIZE_NUMBER_INT":
									fallthrough
								case "FILTER_SANITIZE_NUMBER_FLOAT":
									return false
								}
							}
						}
					}
					return true
				}
			}
		case "filter_input":
			if len(opT.Args) <= 2 {
				return true
			} else {
				filter := opT.Args[2].GetWriteOp()
				switch filterOp := filter.(type) {
				case *cfg.OpExprConstFetch:
					constName, err := cfg.GetOperName(filterOp.Name)
					if err != nil {
						log.Fatalf("error in IsSource: %v", err)
					}
					if len(opT.Args) == 3 {
						switch constName {
						case "FILTER_SANITIZE_NUMBER_INT":
							fallthrough
						case "FILTER_SANITIZE_NUMBER_FLOAT":
							return false
						default:
							return true
						}
					}
				case *cfg.OpExprArray:
					for _, arg := range filterOp.Vals {
						arrFilter := arg.GetWriteOp()
						switch filterOp := arrFilter.(type) {
						case *cfg.OpExprConstFetch:
							constName, err := cfg.GetOperName(filterOp.Name)
							if err != nil {
								log.Fatalf("error in IsSource: %v", err)
							}
							if len(opT.Args) == 3 {
								switch constName {
								case "FILTER_SANITIZE_NUMBER_INT":
									fallthrough
								case "FILTER_SANITIZE_NUMBER_FLOAT":
									return false
								}
							}
						}
					}
					return true
				}
			}
		case "apache_request_headers":
			fallthrough
		case "getallheaders":
			return true
		}
	case *cfg.OpReset:
		return false
	case *cfg.OpExprArrayDimFetch:
		if right, ok := opT.Var.(*cfg.OperSymbolic); ok {
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
		} else if varName, ok := cfg.GetOperVal(opT.Var).(*cfg.OperString); ok {
			if !ok {
				return false
			}
			switch varName.Val {
			case "$_POST":
				fallthrough
			case "$_GET":
				fallthrough
			case "$_REQUEST":
				fallthrough
			case "$_FILES":
				fallthrough
			case "$_COOKIE":
				fallthrough
			case "$_SERVERS":
				return true
			}
		}
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

	switch opT := op.(type) {
	case *cfg.OpExprStaticCall:
		className, strClass := cfg.GetOperVal(opT.Class).(*cfg.OperString)
		methodName, strMethod := cfg.GetOperVal(opT.Name).(*cfg.OperString)
		if strClass && strMethod {
			if strings.HasPrefix(className.Val, "Route") {
				switch methodName.Val {
				case "get":
					fallthrough
				case "post":
					fallthrough
				case "put":
					fallthrough
				case "patch":
					fallthrough
				case "delete":
					return true
				}
			}
		}
	}
	return false
}

func IsPropagated(op cfg.Op, taintedVar cfg.Operand) bool {
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
	// data type isn't string
	case *cfg.OpExprCastBool, *cfg.OpExprCastDouble, *cfg.OpExprCastInt, *cfg.OpExprCastUnset, *cfg.OpUnset:
		return false
	case *cfg.OpExprAssertion:
		switch assert := opT.Assertion.(type) {
		case *cfg.TypeAssertion:
			isNot := assert.IsNegated
			if typeVal, ok := assert.Val.(*cfg.OperString); ok {
				switch typeVal.Val {
				case "int", "float", "bool", "null", "numeric", "alpha", "alnum", "cntrl":
					if !isNot {
						return false
					}
				}
			}
		}
	// array index
	case *cfg.OpExprArrayDimFetch:
		if opT.Dim == taintedVar {
			return false
		}
	// for programming failure
	case *cfg.OpStmtJumpIf, *cfg.OpExprBinaryPlus, *cfg.OpExprBinaryMinus, *cfg.OpExprBinaryMod,
		*cfg.OpExprBinaryDiv, *cfg.OpExprBinaryBitwiseAnd, *cfg.OpExprBinaryBitwiseOr, *cfg.OpExprBinaryBitwiseXor,
		*cfg.OpExprEmpty, *cfg.OpExprBinaryEqual, *cfg.OpExprBinaryGreater, *cfg.OpExprBinaryGreaterOrEqual,
		*cfg.OpExprBinaryIdentical, *cfg.OpExprBinaryLogicalAnd, *cfg.OpExprBinaryLogicalOr, *cfg.OpExprBinaryLogicalXor,
		*cfg.OpExprBinaryMul, *cfg.OpExprBinaryNotEqual, *cfg.OpExprBinaryNotIdentical, *cfg.OpExprBinaryPow,
		*cfg.OpExprBinaryShiftLeft, *cfg.OpExprBinaryShiftRight, *cfg.OpExprBinarySmaller, *cfg.OpExprBinarySmallerOrEqual,
		*cfg.OpExprBitwiseNot, *cfg.OpReset, *cfg.OpExit, *cfg.OpExprIsset, *cfg.OpStmtSwitch, *cfg.OpEcho:
		return false
	}

	// other function call which not related
	if fnCall, ok := op.(*cfg.OpExprFunctionCall); ok {
		fnName, err := cfg.GetOperName(fnCall.Name)
		if err != nil {
			log.Fatalf("error in IsPropagated: %v", err)
		}
		switch fnName {
		case "count_chars", "crc32", "sizeof", "count", "strlen", "strpos", "stripos", "strrpos",
			"strripos", "ord", "substr_count", "bindec", "strspn", "hexdec":
			return false
		case "str_word_count":
			if len(fnCall.Args) == 1 {
				return false
			}
			formatArg, ok := fnCall.Args[1].(*cfg.OperNumber)
			if ok && formatArg.Val == 0 {
				return false
			}
		case "filter_var":
			filter := fnCall.Args[1].GetWriteOp()
			switch filterOp := filter.(type) {
			case *cfg.OpExprConstFetch:
				constName, err := cfg.GetOperName(filterOp.Name)
				if err != nil {
					log.Fatalf("error in IsSource: %v", err)
				}
				if len(fnCall.Args) == 2 {
					switch constName {
					case "FILTER_SANITIZE_NUMBER_INT":
						fallthrough
					case "FILTER_SANITIZE_NUMBER_FLOAT":
						return false
					}
				}
			}
		case "intval":
			fallthrough
		case "floatval":
			fallthrough
		case "boolval":
			fallthrough
		case "doubleval":
			return false
		case "sha1":
			arg2, ok := fnCall.Args[1].(*cfg.OperBool)
			if ok && !arg2.Val {
				return false
			}
		case "bin2hex", "hash", "metaphone", "hash_hmac", "gzdeflate", "soundex", "zlib_encode",
			"base64_encode", "md5", "gzcompress", "mhash", "password_hash", "crypt", "gzencode":
			return false
		}
	}

	return true
}

func IsSink(op cfg.Op, taintedVar cfg.Operand) bool {
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
		case "mysqli_prepare":
			fallthrough
		case "pg_query":
			fallthrough
		case "pg_send_query":
			fallthrough
		case "pg_prepare":
			fallthrough
		case "pg_send_prepare":
			return true
		case "mysqli_execute_query":
			if opT.Args[0] == taintedVar {
				return true
			}
		case "pg_query_params":
			fallthrough
		case "pg_send_query_params":
			if opT.Args[1] == taintedVar {
				return true
			}
		}
	case *cfg.OpExprMethodCall:
		methodNameStr, _ := cfg.GetOperName(opT.Name)
		switch methodNameStr {
		case "query":
			fallthrough
		case "multi_query":
			fallthrough
		case "real_query":
			fallthrough
		case "prepare":
			return true
		case "execute_query":
			if opT.Args[0] == taintedVar {
				return true
			}
		}
		if strings.HasSuffix(methodNameStr, "query") {
			return true
		}
	case *cfg.OpExprStaticCall:
		methodNameStr, _ := cfg.GetOperName(opT.Name)
		switch methodNameStr {
		case "query":
			fallthrough
		case "multi_query":
			fallthrough
		case "real_query":
			fallthrough
		case "prepare":
			return true
		case "execute_query":
			if opT.Args[0] == taintedVar {
				return true
			}
		}
		if strings.HasSuffix(methodNameStr, "query") {
			return true
		}
	}

	// laravel sink
	switch opT := op.(type) {
	case *cfg.OpExprMethodCall:
		methodName, isString := cfg.GetOperVal(opT.Name).(*cfg.OperString)

		if isString {
			switch methodName.Val {
			case "raw":
				fallthrough
			case "selectRaw":
				fallthrough
			case "whereRaw":
				fallthrough
			case "orWhereRaw":
				fallthrough
			case "havingRaw":
				fallthrough
			case "orHavingRaw":
				fallthrough
			case "orderByRaw":
				fallthrough
			case "groupByRaw":
				return true
			case "select":
				fallthrough
			case "where":
				fallthrough
			case "orWhere":
				fallthrough
			case "having":
				fallthrough
			case "orderBy":
				fallthrough
			case "groupBy":
				if opT.Args[0] == taintedVar {
					return true
				}
			}
		}
	case *cfg.OpExprStaticCall:
		methodName, isString := cfg.GetOperVal(opT.Name).(*cfg.OperString)

		if isString {
			switch methodName.Val {
			case "raw":
				fallthrough
			case "selectRaw":
				fallthrough
			case "whereRaw":
				fallthrough
			case "orWhereRaw":
				fallthrough
			case "havingRaw":
				fallthrough
			case "orHavingRaw":
				fallthrough
			case "orderByRaw":
				fallthrough
			case "groupByRaw":
				return true
			case "select":
				fallthrough
			case "where":
				fallthrough
			case "orWhere":
				fallthrough
			case "having":
				fallthrough
			case "orderBy":
				fallthrough
			case "groupBy":
				if opT.Args[0] == taintedVar {
					return true
				}
			}
		}
	}
	return false
}

func GetTaintedVar(op cfg.Op) (cfg.Operand, error) {
	if assignOp, ok := op.(*cfg.OpExprAssign); ok {
		if assignOp.Var != nil {
			return assignOp.Var, nil
		} else {
			return assignOp.Result, nil
		}
	}
	if fnCall, ok := op.(*cfg.OpExprFunctionCall); ok {
		fnName, ok := cfg.GetOperVal(fnCall.Name).(*cfg.OperString)
		if ok {
			switch fnName.Val {
			case "parse_str":
				if len(fnCall.Args) >= 2 {
					return fnCall.Args[1], nil
				}
			}
		}
	}
	if result, ok := op.GetOpVars()["Result"]; ok {
		if result != nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("wrong tainted var '%v'", reflect.TypeOf(op))
}
