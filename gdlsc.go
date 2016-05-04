package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ryanuber/go-license"
)

const root = "vendor/"

var files []string
var rootPath string
var noLicense = map[string]bool{}

// Response contains license and file info
type Response struct {
	License string
	File    string
}

func getLicense(input chan string, output chan Response) {
	for file := range input {
		l, err := license.NewFromFile(file)
		splitFile := strings.Split(file, "/")
		splitFile = splitFile[1 : len(splitFile)-1]
		file := path.Join(splitFile...)
		if err != nil {
			output <- Response{File: file, License: err.Error()}
		} else {
			output <- Response{File: file, License: l.Type}
		}
	}
}
func checkLicense(path string) bool {
	path = strings.ToUpper(path)
	return strings.Contains(path, "LICENSE") || strings.Contains(path, "COPYING")
}

func walker(filePath string, f os.FileInfo, err error) error {
	if err != nil {
		fmt.Println("No vendor folder found. Aborting..")
		return err
	}

	if !f.IsDir() {
		// checks that a path is the root path and assigns it to rootpath.
		// Determines root path by checking the directory of the first file seen after a change in the previous root path
		if currentDir := path.Dir(filePath); !strings.Contains(currentDir, ".git") && (!strings.HasPrefix(currentDir, rootPath) || rootPath == "") {
			rootPath = currentDir
			noLicense[rootPath] = true
		}
		rawPath := path.Base(filePath)
		if checkLicense(rawPath) {
			//removes current rootpath from no license list once a license is seen in one of the rootpaths subdirectories
			delete(noLicense, rootPath)
			files = append(files, filePath)
		}
	}
	return nil
}

func main() {
	input := make(chan string)
	output := make(chan Response)

	filepath.Walk(root, walker)

	for a := 0; a < 5; a++ {
		go getLicense(input, output)
	}

	if len(files) == 0 { // checks that a license file was gotten
		return
	}

	go func() {
		for _, v := range files {
			input <- v
		}
	}()

	for i := 0; i < len(files); i++ {
		out := <-output
		fmt.Printf("%v =======> %v\n", out.File, out.License)
	}

	for i := range noLicense {
		fmt.Printf("Packages without a License file: %v\n", i)
	}
}
