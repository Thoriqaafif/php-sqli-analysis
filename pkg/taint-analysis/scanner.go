package taintanalysis

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/cfg"
)

type Composer struct {
	Autoload struct {
		Psr4 map[string]string `json:"psr-4"`
	} `json:"autoload"`
}

func Scan(path string) {
	// get psr-4 autoload configuration
	composerPath := path + "/composer.json"
	composer, _ := ParseAutoLoaderConfig(composerPath)

	autoloadConfig := make(map[string]string)
	if composer != nil && composer.Autoload.Psr4 != nil {
		for from, to := range composer.Autoload.Psr4 {
			autoloadConfig[from] = to
		}
	}

	// id := NewIDGenerator()

	// // get file source code
	// src, err := os.ReadFile(filepath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Error handler
	// var parserErrors []*errors.Error
	// errorHandler := func(e *errors.Error) {
	// 	parserErrors = append(parserErrors, e)
	// }

	// // Parse
	// rootNode, err := parser.Parse(src, conf.Config{
	// 	Version:          &version.Version{Major: 5, Minor: 6},
	// 	ErrorHandlerFunc: errorHandler,
	// })

	// if err != nil {
	// 	log.Fatal("Error:" + err.Error())
	// }

	// // Create Symbol Table
	// root, _ := rootNode.(*ast.Root)
	// for _, st := range root.Stmts {
	// 	switch s := st.(type) {
	// 	case *ast.ExprAssign:
	// 	}
	// }

	// // Generate SSA
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
