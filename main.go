package main

import (
	"log"
	"os"

	"github.com/z7zmey/php-parser/pkg/cfg"
	"github.com/z7zmey/php-parser/pkg/errors"
	"github.com/z7zmey/php-parser/pkg/parser"
	"github.com/z7zmey/php-parser/pkg/version"
	"github.com/z7zmey/php-parser/pkg/visitor/dumper"
)

func main() {
	// filePath := "D:\\src\\tugas-akhir\\sqli_datasets\\254507-v1.0.0\\src\\sample.php"
	// src, err := os.ReadFile(filePath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	src := []byte(`<? echo "Hello world";`)

	// Error handler
	var parserErrors []*errors.Error
	errorHandler := func(e *errors.Error) {
		parserErrors = append(parserErrors, e)
	}

	// Parse
	rootNode, err := parser.Parse(src, cfg.Config{
		Version:          &version.Version{Major: 5, Minor: 6},
		ErrorHandlerFunc: errorHandler,
	})

	if err != nil {
		log.Fatal("Error:" + err.Error())
	}

	// Dump
	goDumper := dumper.NewDumper(os.Stdout).
		WithTokens().
		WithPositions()

	rootNode.Accept(goDumper)
}
