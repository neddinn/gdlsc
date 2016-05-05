package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ryanuber/go-license"
)

const root = "vendor/"

// Response contains license and file info
type response struct {
	License string
	File    string
}

type request struct {
	path        string
	licenseFile string
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
func getLicense(input chan request, output chan response, size int) {
	for i := range input {
		file := strings.TrimPrefix(i.path, root)
		if i.licenseFile == "" {
			output <- response{File: file, License: "FILE HAS NO LICENSE"}
		} else {
			l, err := license.NewFromFile(i.licenseFile)
			if err != nil {
				f, err := os.Open(i.licenseFile)
				catch(err)
				dat, err := bufio.NewReader(f).Peek(size)
				catch(err)
				output <- response{File: file, License: "#" + string(dat)}
			} else {
				output <- response{File: file, License: l.Type}
			}
		}
	}
}

func isLicense(fileName string) bool {
	fileName = strings.ToUpper(fileName)
	return strings.Contains(fileName, "LICENSE") || strings.Contains(fileName, "COPYING")
}

func isPackage(filePath string) bool {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return false
	}
	for _, file := range files {
		if file.Name() == ".git" {
			return true
		}
	}
	return false
}

func getLicenseFile(filePath string) string {
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		return ""
	}
	for _, file := range files {
		if isLicense(file.Name()) {
			return path.Join(filePath, file.Name())
		}
	}
	return ""
}

func getPackageLicenses() map[string]string {
	var licenseInfo = map[string]string{}

	filepath.Walk(root, func(filePath string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("No vendor folder found. Aborting..")
			return err
		}

		if isPackage(filePath) {
			licenseInfo[filePath] = getLicenseFile(filePath)
		} else if isLicense(filePath) && f.Mode().IsRegular() {
			licenseInfo[path.Dir(filePath)] = filePath
		}

		return nil
	})
	return licenseInfo
}

func main() {
	input := make(chan request)
	output := make(chan response)

	packageFiles := getPackageLicenses()

	size := 100
	if len(os.Args) > 1 {
		size, _ = strconv.Atoi(os.Args[1])
	}
	for a := 0; a < 5; a++ {
		go getLicense(input, output, size)
	}

	if len(packageFiles) == 0 { // checks that a license file was gotten
		return
	}

	go func() {
		for k, v := range packageFiles {
			input <- request{
				path:        k,
				licenseFile: v,
			}
		}
	}()
	w := tabwriter.NewWriter(os.Stdout, 1, 4, 2, ' ', 0)
	for i := 0; i < len(packageFiles); i++ {
		out := <-output
		_, err := w.Write([]byte(out.File + "\t" + out.License + "\n"))
		if err != nil {
			panic(err)
		}
	}
	w.Flush()

}
