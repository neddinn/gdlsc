package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"strconv"

	"path"

	"sort"

	"github.com/ryanuber/go-license"
)

const root = "vendor/"

func formatLicense(file string, size int) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	data := strings.Split(string(f), "\n")
	if size <= len(data) {
		data = data[0:size]
	}
	formatted := "UNKOWN LICENSE\n\n|\t" + strings.Join(data, "\n|\t") + "\n"
	return formatted
}

func isWtf(file string) bool {
	match := "do what the fuck you want to public license version 2"

	f, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	license := string(f)
	rawText := strings.ToLower(license)
	text := regexp.MustCompile("\\s{2,}").ReplaceAllLiteralString(rawText, " ")
	return strings.Contains(text, match)
}

func checkLicenseType(file string) (string, error) {
	l, err := license.NewFromFile(file)
	if err != nil {
		if isWtf(file) {
			return "WTFPL-2.0", nil
		}
		return "", err
	}
	return l.Type, nil
}

func getLicenseType(file string, size int) string {
	l, err := checkLicenseType(file)
	if err != nil {
		return formatLicense(file, size)
	}
	return l
}

func walkPath(root string) []string {
	var contents []string
	filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if strings.Contains(path, ".git") { //ignore git folders
			return nil
		}
		if f.IsDir() {
			contents = append(contents, path)
		}
		return nil
	})
	return contents
}

func isPackage(folder string) bool {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		panic(err)
	}
	for _, v := range files {
		if strings.HasSuffix(v.Name(), ".go") {
			return true
		}
	}
	return false
}

func checkLicense(folder string) string {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		panic(err)
	}
	for _, v := range files {
		file := strings.ToUpper(v.Name())
		if (v.Mode().IsRegular()) && (strings.Contains(file, "LICENSE") || strings.Contains(file, "COPYING")) {
			return path.Join(folder, v.Name())
		}
	}
	return ""
}
func hasLicense(folder string) bool {
	if checkLicense(folder) == "" {
		return false
	}
	return true
}

func filter(folders []string) []string {
	var filtered []string
	for _, v := range folders {
		if isPackage(v) || hasLicense(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func generateOutput(folders []string, size int) string {
	// Walk dir printing the Packages and keepijng track of the root
	// If license is found, store the license
	// If another license is found, print the stored license and store the new one
	// If root path changes, print the stored license if it exists otherwise "No license" and reset license
	// var root string
	// var license string
	var noLicenseError = "No License Found"
	var dirRoot, output string
	license := noLicenseError
	for _, v := range folders {

		if dirRoot == "" {
			dirRoot = v
		} else if !strings.Contains(v, dirRoot) {
			output += "\t" + license + "\n"
			dirRoot = v
			license = noLicenseError
		}

		if licenseFile := checkLicense(v); licenseFile != "" {
			if license != noLicenseError {
				output += "\t" + license + "\n"
			}
			license = getLicenseType(licenseFile, size)
		}

		output += strings.TrimPrefix(v, root) + "\n"
	}
	output += "\t" + license
	return output
}

func main() {
	size := 5
	if len(os.Args) > 1 {
		size, _ = strconv.Atoi(os.Args[1])
	}

	files := walkPath(root)
	if len(files) == 0 {
		fmt.Println("Vendor Folder not found. Aborting..")
		return
	}
	filtered := filter(files)
	sort.Strings(filtered)

	output := generateOutput(filtered, size)
	fmt.Println(output)

}
