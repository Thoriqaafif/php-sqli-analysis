package taintanalysis

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/optimizer"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/simplifier"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/pathgenerator"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis/report"
)

type Composer struct {
	Autoload struct {
		Psr4 map[string]string `json:"psr-4"`
	} `json:"autoload"`
}

type Scanner struct {
	Sources     map[cfg.Op]report.Node
	VarNodes    map[cfg.Op]report.Node
	TaintedNext map[cfg.Op]cfg.Op
	TaintedPrev map[cfg.Op]cfg.Op
}

func NewScanner() *Scanner {
	return &Scanner{
		Sources:     make(map[cfg.Op]report.Node),
		VarNodes:    make(map[cfg.Op]report.Node),
		TaintedNext: make(map[cfg.Op]cfg.Op),
		TaintedPrev: make(map[cfg.Op]cfg.Op),
	}
}

func Scan(dirPath string, filePaths []string) *report.ScanReport {
	// get psr-4 autoload configuration
	composerPath := dirPath + "/composer.json"
	composer, _ := ParseAutoLoaderConfig(composerPath)

	autoloadConfig := make(map[string]string)
	if composer != nil && composer.Autoload.Psr4 != nil {
		for from, to := range composer.Autoload.Psr4 {
			autoloadConfig[from] = to
		}
	}

	// build ssa form cfg for each file
	scripts := make(map[string]*cfg.Script)
	for _, filePath := range filePaths {
		// fullPath := dirPath + "\\" + filePath
		fmt.Println(filePath)
		src, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}

		script := cfg.BuildCFG(src, filePath, make(map[string]string))

		// simplify and optimize SSA
		cfgTraverser := traverser.NewTraverser()
		trivPhiRemover := simplifier.NewSimplifier()
		optimizer := optimizer.NewOptimizer(script.FilePath)
		cfgTraverser.AddTraverser(trivPhiRemover)
		cfgTraverser.Traverse(script)

		cfgTraverser = traverser.NewTraverser()
		cfgTraverser.AddTraverser(optimizer)
		cfgTraverser.Traverse(script)

		// path generator
		scripts[filePath] = script
	}
	paths := pathgenerator.GenerateFeasiblePath(scripts)

	// do taint analysis for each feasible path
	sources := make(map[cfg.Op]report.Node)
	sinks := make(map[cfg.Op]report.Node)
	taintedVarGraph := make(map[cfg.Op]map[cfg.Op]struct{})
	for _, path := range paths {
		currTaintedVars := make(map[cfg.Operand]struct{})
		for _, intr := range path.Instructions {
			if IsSource(path, intr) {
				// set source as tainted var
				assignOp := intr.(*cfg.OpExprAssign)
				currTaintedVars[assignOp.Result] = struct{}{}
				currTaintedVars[assignOp.Var] = struct{}{}
				// add source
				var reportNode report.Node
				rightPos := assignOp.ExprPos
				if right, ok := assignOp.Expr.(*cfg.OperSymbolic); ok {
					endLoc := report.NewLoc(rightPos.EndLine, rightPos.EndPos)
					startLoc := report.NewLoc(rightPos.StartLine, rightPos.StartPos)
					reportNode = report.NewCodeNode(right.Val, intr.GetFilePath(), endLoc, startLoc)
				} else if assignOp.Expr.GetWriteOp() != nil {
					reportNode = OptoReportNode(assignOp.Expr.GetWriteOp())
				} else {
					intrPos := intr.GetPosition()
					sourceContent := GetFileContent(intr.GetFilePath(), intrPos.EndPos, intrPos.StartPos)
					log.Fatalf("Error: Unknown source '%s'", sourceContent)
				}
				sources[intr] = reportNode
			} else if IsSink(path, currTaintedVars, intr) {
				useTainted := false
				// sink is a function/method/static call
				for _, args := range intr.GetOpListVars() {
					for _, arg := range args {
						arg = path.GetVar(arg)
						if _, ok := currTaintedVars[arg]; ok {
							useTainted = true
							// intr (Op) use arg which is tainted var (Operand)
							argDef := arg.GetWriteOp()
							if _, ok := taintedVarGraph[intr]; !ok {
								taintedVarGraph[intr] = make(map[cfg.Op]struct{})
							}
							taintedVarGraph[intr][argDef] = struct{}{}
						}
					}
				}
				// sink use tainted var
				if useTainted {
					// add sink
					sinkNode := OptoReportNode(intr)
					sinks[intr] = sinkNode
				}
			} else {
				taintedVars := PropagateTaintedData(path, currTaintedVars, intr)
				// if use tainted var, add to usedTaintedVars
				if len(taintedVars) > 0 {
					for _, taintedVar := range taintedVars {
						// intr op use taintedVar
						varDef := taintedVar.GetWriteOp()
						if _, ok := taintedVarGraph[intr]; !ok {
							taintedVarGraph[intr] = make(map[cfg.Op]struct{})
						}
						taintedVarGraph[intr][varDef] = struct{}{}
					}
				}
			}
		}
	}

	// add each tainted dataflow to report
	newReport := report.NewScanReport(filePaths)
	for sink, sinkNode := range sinks {
		// dfs until source
		dataflow := NewOpStack()
		st := NewOpStack()
		for taintedVar := range taintedVarGraph[sink] {
			st.Push(taintedVar)
		}
		for !st.IsEmpty() {
			currVar, _ := st.Pop()
			dataflow.Push(currVar)
			if sourceNode, isSource := sources[currVar]; isSource {
				result := report.NewResult(sinkNode.Location.Start, sinkNode.Location.End, sinkNode.Location.Path)
				result.SetSink(sinkNode)
				result.SetSource(sourceNode)
				// trace the tainted data
				for i := len(*dataflow) - 1; i >= 0; i-- {
					reportNode := OptoReportNode((*dataflow)[i])
					result.AddIntermediateVar(reportNode)
				}
				newReport.AddResult(result)
			} else {
				for taintedVar := range taintedVarGraph[currVar] {
					st.Push(taintedVar)
				}
			}
			dataflow.Pop()
		}
	}

	return newReport
}

func ParseAutoLoaderConfig(filePath string) (*Composer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var composer Composer
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &composer, nil
}

func IsSource(path *pathgenerator.ExecPath, op cfg.Op) bool {
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
			if right, ok := assignOp.Expr.GetWriteOp().(*cfg.OpExprFunctionCall); ok {
				funcNameStr := cfg.GetOperName(right.Name)
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
			}
		}
	}

	// TODO: laravel source

	return false
}

func PropagateTaintedData(path *pathgenerator.ExecPath, taintedVarSet map[cfg.Operand]struct{}, op cfg.Op) []cfg.Operand {
	// for sanitizer, data cannot be tainted
	switch opT := op.(type) {
	case *cfg.OpExprFunctionCall:
		funcNameStr := cfg.GetOperName(opT.Name)
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
			return nil
		case "preg_match":
			arg0 := path.GetVar(opT.Args[0])
			if arg0Str, ok := arg0.(*cfg.OperString); ok && arg0Str.Val == "/^[0-9]*$/" {
				return nil
			}
		}
	case *cfg.OpExprMethodCall:
		funcNameStr := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "escape_string":
			fallthrough
		case "quote":
			return nil
		}
	case *cfg.OpExprStaticCall:
		funcNameStr := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "escape_string":
			fallthrough
		case "quote":
			return nil
		}
	case *cfg.OpExprCastBool, *cfg.OpExprCastDouble, *cfg.OpExprCastInt, *cfg.OpExprCastUnset, *cfg.OpUnset:
		return nil
	case *cfg.OpExprAssertion:
		switch assert := opT.Assertion.(type) {
		case *cfg.TypeAssertion:
			if typeVal, ok := assert.Val.(*cfg.OperString); ok {
				switch typeVal.Val {
				case "int", "float", "bool", "null":
					return nil
				}
			}
		}
	}

	usedTainted := make([]cfg.Operand, 0)
	var resultVr cfg.Operand
	for _, vrList := range op.GetOpListVars() {
		for _, vr := range vrList {
			vr = path.GetVar(vr)
			if _, tainted := taintedVarSet[vr]; tainted {
				usedTainted = append(usedTainted, vr)
			}
		}
	}
	for vrName, vr := range op.GetOpVars() {
		vr = path.GetVar(vr)
		if _, tainted := taintedVarSet[vr]; vrName != "Result" && tainted {
			usedTainted = append(usedTainted, vr)
		}
		if vrName == "Result" {
			resultVr = vr
		}
	}
	if resultVr != nil && len(usedTainted) > 0 {
		// add the op's result to tainted variables
		taintedVarSet[resultVr] = struct{}{}
		// for assign Op, add also the Var Operand to tainted variables
		if assignOp, isAssign := op.(*cfg.OpExprAssign); isAssign {
			taintedVarSet[assignOp.Var] = struct{}{}
		}
		return usedTainted
	}

	return nil
}

func IsSink(path *pathgenerator.ExecPath, taintedVar map[cfg.Operand]struct{}, op cfg.Op) bool {
	// php sink
	switch opT := op.(type) {
	case *cfg.OpExprFunctionCall:
		funcNameStr := cfg.GetOperName(opT.Name)
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
		methodNameStr := cfg.GetOperName(opT.Name)
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
	case *cfg.OpExprStaticCall:
		methodNameStr := cfg.GetOperName(opT.Name)
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
	}

	// TODO: laravel sink

	return false
}

func TraceTaintedVars(path *pathgenerator.ExecPath, op cfg.Op) []report.Result {
	return nil
}

func OptoReportNode(op cfg.Op) report.Node {
	// read the content based on op position
	filePath := op.GetFilePath()
	if filePath == "" {
		log.Fatal("Error: cannot convert Op to Node, Op don't have filepath")
	}
	opPos := op.GetPosition()
	content := GetFileContent(filePath, opPos.EndPos, opPos.StartPos)

	startLoc := report.NewLoc(opPos.StartLine, opPos.StartPos)
	endLoc := report.NewLoc(opPos.EndLine, opPos.EndPos)
	return report.NewCodeNode(string(content), filePath, startLoc, endLoc)
}

func GetFileContent(filePath string, endPos, startPos int) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Cannot open file: %s", err)
	}
	_, err = file.Seek(int64(startPos), 0)
	if err != nil {
		log.Fatal(err)
	}
	length := endPos - startPos + 1
	buffer := make([]byte, length)
	n, err := file.Read(buffer)
	if err != nil {
		log.Fatalf("Cannot get file content: %s", err)
	}

	return string(buffer[:n])
}

type OpStack []cfg.Op

func NewOpStack() *OpStack {
	st := make([]cfg.Op, 0)
	return (*OpStack)(&st)
}

func (st *OpStack) Push(op cfg.Op) {
	*st = append(*st, op)
}

func (st *OpStack) Pop() (cfg.Op, bool) {
	if len(*st) == 0 {
		return nil, false
	}
	lastItem := (*st)[len(*st)-1]
	*st = (*st)[:(len(*st) - 1)]
	return lastItem, true
}

func (st *OpStack) IsEmpty() bool {
	return len(*st) == 0
}
