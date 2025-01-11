package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis"
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taintanalysis/report"
)

type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	NumOfFiles    int       `json:"num_of_files"`
	DetectedVulns int       `json:"detected_vulns"`
	ScanTime      float64   `json:"scan_time"`
	CreatedAt     time.Time `json:"created_at"`
}

func main() {
	if len(os.Args) < 2 {
		errorCommand()
	}

	srcPath := os.Args[1]
	if srcPath == "--help" {
		errorCommand()
	}
	absPath, err := filepath.Abs(srcPath)
	if err != nil {
		log.Fatal("error: invalid project path")
	}

	host := ""
	outPath := "result.json"
	isLaravel := false
	for i := 2; i < len(os.Args); i++ {
		splittedOption := strings.Split(os.Args[i], "=")
		option := splittedOption[0]

		switch option {
		case "--host":
			val := splittedOption[1]
			host = val
		case "--out":
			val := splittedOption[1]
			outPath = val
		case "--laravel":
			isLaravel = true
		case "--help":
			errorCommand()
		default:
			errorCommand()
		}
	}

	// start
	start := time.Now()

	// get php files inside directory
	filePaths, err := getPhpFiles(srcPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Scan %d php files ...\n\n", len(filePaths))

	// scan result
	result := taintanalysis.Scan(srcPath, filePaths, isLaravel)
	// Calculate elapsed time
	elapsed := time.Since(start)
	runTime := elapsed.Seconds()
	fmt.Printf("Detected %d sqli vulnerabilities in %.2f second.\n", len(result.Results), runTime)
	fmt.Printf("Result reported in '%s'\n", outPath)

	// output file to .json
	f, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	jsonEncoder := json.NewEncoder(f)
	jsonEncoder.SetEscapeHTML(false)
	if err := jsonEncoder.Encode(result); err != nil {
		log.Fatalf("Failed to encode content: %s", err)
	}

	// send scan result to server
	if host != "" {
		var project struct {
			Data   Project           `json:"data"`
			Result report.ScanReport `json:"result"`
		}
		project.Data.Name = filepath.Base(absPath)
		project.Data.NumOfFiles = len(result.Paths.Scanned)
		project.Data.DetectedVulns = len(result.Results)
		project.Data.ScanTime = runTime
		project.Data.CreatedAt = time.Now()
		project.Result = *result

		// sent to host/api/project
		projectJson, err := json.Marshal(project)
		if err != nil {
			log.Fatal("error: fail to marshall project json")
		}
		err = sentToHost(host, projectJson)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}

func sentToHost(host string, data []byte) error {
	// create http request with the json data
	url := host + "/api/project"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return errors.New("fail sending request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("fail status code %v", resp.StatusCode)
	}

	return nil
}

func getPhpFiles(dirPath string) ([]string, error) {
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

func errorCommand() {
	fmt.Println("Specify the directory path")
	fmt.Println("Usage: sqli-scanner [directory path] [option]")
	fmt.Println("Options:")
	fmt.Println("    --host: specify web host url")
	fmt.Println("    --out: specify report file path")
	fmt.Println("    --laravel: using laravel taint rule")
	os.Exit(0)
}
