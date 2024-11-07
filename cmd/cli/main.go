package main

import (
	"log"
	"os"

	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	"github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/VKCOM/php-parser/pkg/visitor/dumper"
)

func main() {
	src := []byte(`<?php echo "Hello world";`)

	// Error handler

	var parserErrors []*errors.Error
	errorHandler := func(e *errors.Error) {
		parserErrors = append(parserErrors, e)
	}

	// Parse

	rootNode, err := parser.Parse(src, conf.Config{
		Version:          &version.Version{Major: 8, Minor: 0},
		ErrorHandlerFunc: errorHandler,
	})

	if err != nil {
		log.Fatal("Error:" + err.Error())
	}

	if len(parserErrors) > 0 {
		for _, e := range parserErrors {
			log.Println(e.String())
		}
		os.Exit(1)
	}

	// Dump

	goDumper := dumper.NewDumper(os.Stdout).
		WithTokens().
		WithPositions()

	rootNode.Accept(goDumper)
}
