package taintanalysis

// func Scan(filepath string) {
// 	id := NewIDGenerator()

// 	// get file source code
// 	src, err := os.ReadFile(filepath)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Error handler
// 	var parserErrors []*errors.Error
// 	errorHandler := func(e *errors.Error) {
// 		parserErrors = append(parserErrors, e)
// 	}

// 	// Parse
// 	rootNode, err := parser.Parse(src, conf.Config{
// 		Version:          &version.Version{Major: 5, Minor: 6},
// 		ErrorHandlerFunc: errorHandler,
// 	})

// 	if err != nil {
// 		log.Fatal("Error:" + err.Error())
// 	}

// 	// Create Symbol Table
// 	root, _ := rootNode.(*ast.Root)
// 	for _, st := range root.Stmts {
// 		switch s := st.(type) {
// 		case *ast.ExprAssign:
// 		}
// 	}

// 	// Generate SSA
// }
