package taintanalysis

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg/traverser"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/optimizer"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/simplifier"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfgtraverser/taintfinder"
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

func Scan(dirPath string, filePaths []string, isLaravel bool) *report.ScanReport {
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
	relPaths := make([]string, 0)
	for _, filePath := range filePaths {
		// fullPath := dirPath + "\\" + filePath
		src, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal(err)
		}
		relPath, err := filepath.Rel(dirPath, filePath)
		if err != nil {
			log.Fatal(err)
		}
		relPaths = append(relPaths, relPath)

		script := cfg.BuildCFG(src, filePath, make(map[string]string))

		// simplify and optimize SSA
		cfgTraverser := traverser.NewTraverser()
		trivPhiRemover := simplifier.NewSimplifier()
		optimizer := optimizer.NewOptimizer()
		sourceFinder := taintfinder.NewTaintFinder()
		cfgTraverser.AddTraverser(trivPhiRemover)
		cfgTraverser.Traverse(script)

		cfgTraverser = traverser.NewTraverser()
		cfgTraverser.AddTraverser(optimizer)
		cfgTraverser.AddTraverser(sourceFinder)
		cfgTraverser.Traverse(script)

		// path generator
		scripts[filePath] = script
	}

	var paths [][]cfg.Op
	var err error
	if isLaravel {
		paths, err = pathgenerator.GeneratePaths(scripts, pathgenerator.LARAVEL)
	} else {
		paths, err = pathgenerator.GeneratePaths(scripts, pathgenerator.NATIVE)
	}
	if err != nil {
		log.Fatal(err)
	}

	newReport := report.NewScanReport(relPaths)
	for _, path := range paths {
		var source *report.Node
		var sink *report.Node

		traces := make([]*report.Node, 0)
		for i := 0; i < len(path)-1; i++ {
			if len(traces) == 0 {
				// source
				switch path[i].(type) {
				case *cfg.OpExprAssign, *cfg.OpExprArrayDimFetch, *cfg.OpExprParam:
					if path[i].GetPosition() != nil {
						intermVar, err := OptoReportNode(dirPath, path[i])
						if err != nil {
							log.Fatalf("Error converting intermediate var: %v", err)
						}
						traces = append(traces, intermVar)
					}
				}
			} else {
				switch path[i].(type) {
				case *cfg.OpExprAssign, *cfg.OpExprFunctionCall, *cfg.OpExprMethodCall, *cfg.OpExprStaticCall:
					if path[i].GetPosition() != nil {
						intermVar, err := OptoReportNode(dirPath, path[i])
						if err != nil {
							log.Fatalf("Error converting intermediate var: %v", err)
						}
						traces = append(traces, intermVar)
					}
				}
			}
		}
		if len(traces) > 0 {
			source = traces[0]
			sink = traces[len(traces)-1]
			result := report.NewResult(sink.Location.Start, sink.Location.End, sink.Location.Path)
			result.SetSource(*source)
			result.SetSink(*sink)
			for i := 1; i < len(traces)-1; i++ {
				result.AddIntermediateVar(*traces[i])
			}
			result.SetMessage("SQLi vulnerability")
			newReport.AddResult(*result)
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

func OptoReportNode(dirPath string, op cfg.Op) (*report.Node, error) {
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
	relPath, err := filepath.Rel(dirPath, filePath)
	if err != nil {
		return nil, err
	}

	startLoc := report.NewLoc(opPos.StartLine, opPos.StartPos)
	endLoc := report.NewLoc(opPos.EndLine, opPos.EndPos)
	return report.NewCodeNode(string(content), relPath, startLoc, endLoc), nil
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
