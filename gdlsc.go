package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/ryanuber/go-license"
)

const vendorDir = "vendor"

// node describes a directory in the vendor tree
type node struct {
	prefix     string
	name       string
	golang     bool
	licenseTxt string
	children   []*node
	reduced    bool
}

// holder allows to append while keeping a reference to the slice
type holder struct {
	nodes []*node
}

// makeTree parses the vendor tree
func makeTree(p string) *node {
	n := &node{prefix: path.Dir(p), name: path.Base(p), children: make([]*node, 0)}

	files, err := ioutil.ReadDir(p)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			if strings.HasSuffix(file.Name(), ".go") {
				n.golang = true
			} else if isLicenseFileName(file.Name()) {
				license, err := ioutil.ReadFile(path.Join(p, file.Name()))
				if err != nil {
					panic(err)
				}
				n.licenseTxt = string(license)
			}
		} else if file.Mode().IsDir() && !strings.HasPrefix(file.Name(), ".") && file.Name() != "ConnectCorp" {
			n.children = append(n.children,
				makeTree(path.Join(p, file.Name())))
		}
	}

	return n
}

// Extract takes out some of the minimal licensed subtrees
func (n *node) extract(h *holder) bool {
	extracted := false
	newChildren := make([]*node, 0, len(n.children))

	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		license := findLicense(child.children)
		if child.licenseTxt != "" && !license {
			child.reduced = true
			child.children = []*node{}
			h.nodes = append(h.nodes, child)
			extracted = true
		} else {
			newChildren = append(newChildren, child)
			if child.extract(h) {
				extracted = true
			}
		}
	}

	n.children = newChildren
	return extracted
}

// AttachLicenseType finds the appropriate licence types and attaches it
func (n *holder) attachLicenseType() {
	for i := 0; i < len(n.nodes); i++ {
		child := n.nodes[i]
		if child.licenseTxt != "" {
			child.licenseTxt = getLicenseType(child.licenseTxt)
		}
	}
}

func (n *node) reduce() bool {
	reduced := false

	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		if !n.golang && child.golang && len(child.children) > 0 {
			child.reduced = true
			child.children = []*node{}
			reduced = true
		} else {
			if child.reduce() {
				reduced = true
			}
		}
	}

	return reduced
}

// format the given tree for output
func (n *node) format() []string {
	r := ""
	if n.reduced {
		r = "/..."
	}
	if n.licenseTxt != "" {
		r += " ======>"
	}
	out := make([]string, 0)
	if n.golang || n.licenseTxt != "" {
		out = append(out, fmt.Sprintf("%v/%v%v %v", n.prefix, n.name, r, n.licenseTxt))
	}
	for _, child := range n.children {
		out = append(out, child.format()...)
	}
	return out
}

func isLicenseFileName(name string) bool {
	name = strings.ToUpper(name)
	return strings.Contains(name, "LICENSE") || strings.Contains(name, "COPYING")
}

func isWtfLicense(text string) bool {
	match := "do what the fuck you want to public license version 2"
	rawText := strings.ToLower(text)
	formattedText := regexp.MustCompile("\\s{2,}").ReplaceAllLiteralString(rawText, " ")
	return strings.Contains(formattedText, match)
}

func formatLicenseText(text string) string {
	data := strings.Split(text, "\n")
	if len(data) >= 10 {
		data = data[0:10]
	}
	formatted := "UNKOWN LICENSE\n|\t" + strings.Join(data, "\n|\t") + "\n"
	return formatted
}

func getLicenseType(text string) string {
	l := new(license.License)
	l.Text = text
	if err := l.GuessType(); err != nil {
		if isWtfLicense(text) {
			return "WTFPL-2.0"
		}
		return formatLicenseText(text)
	}

	return l.Type
}

// finds a license in a forest
func findLicense(nodes []*node) bool {
	for i := 0; i < len(nodes); i++ {
		if nodes[i].licenseTxt != "" {
			return true
		}
		if findLicense(nodes[i].children) {
			return true
		}
	}
	return false
}

func main() {
	n := makeTree(path.Join(os.Args[1], vendorDir))
	h := &holder{nodes: make([]*node, 0)}

	for n.extract(h) {
	}
	h.attachLicenseType()
	for n.reduce() {
	}

	fmt.Println("LICENSED")
	for _, e := range h.nodes {
		fmt.Println(strings.Join(e.format(), "\n"))
	}

	fmt.Println("UNLICENSED")
	fmt.Println(strings.Join(n.format(), "\n"))
}
