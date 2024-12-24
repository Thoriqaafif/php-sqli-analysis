package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Specify the directory path")
		fmt.Println("Usage: sqli-scanner [directory path]")
		os.Exit(0)
	}

	srcPath := os.Args[1]
	outPath := "result.json"
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	}

	// start
	start := time.Now()

	// get php files inside directory
	filePaths, err := GetPhpFiles(srcPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Scan %d php files ...\n\n", len(filePaths))

	// scan result
	result := taintanalysis.Scan(srcPath, filePaths)
	// Calculate elapsed time
	elapsed := time.Since(start)
	fmt.Printf("Detected %d sqli vulnerabilities in %.2f second.\n", len(result.Results), elapsed.Seconds())
	fmt.Printf("Result reported in '%s'\n", outPath)
	resultJson, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile(outPath, resultJson, os.FileMode(os.O_RDWR))
}

func GetPhpFiles(dirPath string) ([]string, error) {
	var files []string
	extension := ".php"

	// Walk through the directory
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip vendor directory
		if d.IsDir() && d.Name() == "vendor" {
			return filepath.SkipDir
		}

		// Check if it's a file and matches the extension
		if !d.IsDir() && strings.HasSuffix(d.Name(), extension) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
