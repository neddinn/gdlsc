package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryanuber/go-license"
)

type response struct {
	License string
	File    string
}

func getLicense(input chan string, output chan response) {
	for file := range input {
		l, err := license.NewFromFile(file)
		splitFile := strings.Split(file, "/")
		newFile := splitFile[1 : len(splitFile)-1]
		file = strings.Join(newFile, "/")
		if err != nil {
			output <- response{File: file, License: err.Error()}
		} else {
			output <- response{File: file, License: l.Type}
		}
	}
}
func scan(path string) bool {
	path = strings.ToUpper(path)
	return strings.Contains(path, "LICENSE") || strings.Contains(path, "COPYING")
}

func main() {
	root := "vendor/"
	i := 0
	input := make(chan string)
	output := make(chan response)

	for a := 0; a < 5; a++ {
		go getLicense(input, output)
	}
	var files []string
	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("No vendor folder found. Aborting..")
			return err
		}
		if !f.IsDir() {
			splitPath := strings.Split(path, "/")
			rawPath := splitPath[len(splitPath)-1]
			if scan(rawPath) {
				files = append(files, path)
			}
		}
		return nil
	})

	if len(files) == 0 {
		return
	}

	go func() {
		for _, v := range files {
			input <- v
		}
	}()
	for out := range output {
		fmt.Printf("%v =======> %v\n", out.File, out.License)
		i++
		if i == len(files)-1 || len(files) == 1 {
			break
		}

	}
}
