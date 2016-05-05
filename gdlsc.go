package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"sort"
	"strconv"

	"github.com/ryanuber/go-license"
)

const root = "vendor/"

// Response contains license and file info
type response struct {
	License string
	File    string
}

// ByName type is used to implement sort interface to sort response by name
type ByName []response

func (s ByName) Len() int           { return len(s) }
func (s ByName) Less(i, j int) bool { return s[i].File < s[j].File }
func (s ByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (r response) String() string {
	return r.File + " ------- > " + r.License + "\n"
}

type request struct {
	Path        string
	LicenseFile string
}

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
	match := "everyone is permitted to copy and distribute"
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

func getLicense(input chan request, output chan response, size int) {
	for i := range input {
		file := strings.TrimPrefix(i.Path, root)
		if i.LicenseFile == "" {
			output <- response{File: file, License: "FILE HAS NO LICENSE"}
		} else {
			l, err := checkLicenseType(i.LicenseFile)
			if err != nil {
				data := formatLicense(i.LicenseFile, size)
				output <- response{File: file, License: data}
			} else {
				output <- response{File: file, License: l}
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

	size := 5
	if len(os.Args) > 1 {
		size, _ = strconv.Atoi(os.Args[1])
	}
	for a := 0; a < 5; a++ {
		go getLicense(input, output, size)
	}

	if len(packageFiles) == 0 {
		return
	}

	go func() {
		for k, v := range packageFiles {
			input <- request{
				Path:        k,
				LicenseFile: v,
			}
		}
	}()
	var out []response
	for i := 0; i < len(packageFiles); i++ {
		out = append(out, <-output)
	}

	sort.Sort(ByName(out))
	for _, v := range out {
		fmt.Println(v)
	}

}
