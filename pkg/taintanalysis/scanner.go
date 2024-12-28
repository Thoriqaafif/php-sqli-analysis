package taintanalysis

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/optimizer"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/simplifier"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/sourcefinder"
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
		optimizer := optimizer.NewOptimizer()
		sourceFinder := sourcefinder.NewSourceFinder()
		cfgTraverser.AddTraverser(trivPhiRemover)
		cfgTraverser.Traverse(script)

		cfgTraverser = traverser.NewTraverser()
		cfgTraverser.AddTraverser(optimizer)
		cfgTraverser.AddTraverser(sourceFinder)
		cfgTraverser.Traverse(script)

		// path generator
		scripts[filePath] = script
	}
	paths, err := pathgenerator.GeneratePaths(scripts, pathgenerator.NATIVE)
	if err != nil {
		log.Fatal(err)
	}

	newReport := report.NewScanReport(filePaths)
	for _, path := range paths {
		source, err := OptoReportNode(path[0])
		if err != nil {
			log.Fatalf("Error converting source: %v", err)
		}
		sink, err := OptoReportNode(path[len(path)-1])
		if err != nil {
			log.Fatalf("Error converting sink: %v", err)
		}
		fmt.Println(sink.Location.Path)
		result := report.NewResult(sink.Location.Start, sink.Location.End, sink.Location.Path)
		result.SetSource(*source)
		result.SetSink(*sink)

		result.AddIntermediateVar(*source)
		for i := 1; i < len(path)-1; i++ {
			switch path[i].(type) {
			case *cfg.OpExprAssign, *cfg.OpExprFunctionCall, *cfg.OpExprMethodCall, *cfg.OpExprStaticCall:
				intermVar, err := OptoReportNode(path[i])
				if err != nil {
					log.Fatalf("Error converting intermediate var: %v", err)
				}
				result.AddIntermediateVar(*intermVar)
			}
		}
		result.AddIntermediateVar(*sink)
		result.SetMessage("SQLi vulnerability")
		newReport.AddResult(*result)
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

func PropagateTaintedData(path *pathgenerator.ExecPath, taintedVarSet map[cfg.Operand]struct{}, op cfg.Op) []cfg.Operand {
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
			return nil
		case "preg_match":
			arg0 := path.GetVar(opT.Args[0])
			if arg0Str, ok := arg0.(*cfg.OperString); ok && arg0Str.Val == "/^[0-9]*$/" {
				return nil
			}
		}
	case *cfg.OpExprMethodCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
		switch funcNameStr {
		case "escape_string":
			fallthrough
		case "quote":
			return nil
		}
	case *cfg.OpExprStaticCall:
		funcNameStr, _ := cfg.GetOperName(opT.Name)
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
	}

	// TODO: laravel sink

	return false
}

func TraceTaintedVars(path *pathgenerator.ExecPath, op cfg.Op) []report.Result {
	return nil
}

func OptoReportNode(op cfg.Op) (*report.Node, error) {
	// read the content based on op position
	filePath := op.GetFilePath()
	if filePath == "" {
		return nil, fmt.Errorf("cannot convert Op to Node, Op '%v' don't have filepath", reflect.TypeOf(op))
	}
	opPos := op.GetPosition()
	if opPos == nil {
		return nil, fmt.Errorf("cannot convert Op to Node, Op '%v' have nil position", reflect.TypeOf(op))
	}
	content := GetFileContent(filePath, opPos.EndPos, opPos.StartPos)

	startLoc := report.NewLoc(opPos.StartLine, opPos.StartPos)
	endLoc := report.NewLoc(opPos.EndLine, opPos.EndPos)
	return report.NewCodeNode(string(content), filePath, startLoc, endLoc), nil
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
