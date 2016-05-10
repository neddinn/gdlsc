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

// Node describes a directory in the vendor tree
type Node struct {
	prefix     string
	name       string
	golang     bool
	licenseTxt string
	children   []*Node
	reduced    bool
}

// Holder allows to append while keeping a reference to the slice
type Holder struct {
	nodes []*Node
}

// MakeTree parses the vendor tree
func MakeTree(p string) *Node {
	n := &Node{prefix: path.Dir(p), name: path.Base(p), children: make([]*Node, 0)}

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
				MakeTree(path.Join(p, file.Name())))
		}
	}

	return n
}

// Extract takes out some of the minimal licensed subtrees
func (n *Node) Extract(h *Holder) bool {
	extracted := false
	newChildren := make([]*Node, 0, len(n.children))

	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		license := findLicense(child.children)
		if child.licenseTxt != "" && !license {
			child.reduced = true
			child.children = []*Node{}
			h.nodes = append(h.nodes, child)
			extracted = true
		} else {
			newChildren = append(newChildren, child)
			if child.Extract(h) {
				extracted = true
			}
		}
	}

	n.children = newChildren
	return extracted
}

// AttachLicenseType finds the appropriate licence types and attaches it
func (n *Holder) AttachLicenseType() {
	for i := 0; i < len(n.nodes); i++ {
		child := n.nodes[i]
		if child.licenseTxt != "" {
			child.licenseTxt = getLicenseType(child.licenseTxt)
		}
	}
}

func (n *Node) reduce() bool {
	reduced := false

	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		if !n.golang && child.golang && len(child.children) > 0 {
			child.reduced = true
			child.children = []*Node{}
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
func (n *Node) format() []string {
	r := ""
	if n.licenseTxt != "" {
		r = "======>"
	}
	out := make([]string, 0)
	if n.golang || n.licenseTxt != "" {
		out = append(out, fmt.Sprintf("%v/%v %v %v", n.prefix, n.name, r, n.licenseTxt))
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
func findLicense(nodes []*Node) bool {
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
	n := MakeTree(path.Join(os.Args[1], vendorDir))
	h := &Holder{nodes: make([]*Node, 0)}

	for n.Extract(h) {
	}
	h.AttachLicenseType()
	for n.reduce() {
	}

	fmt.Println("LICENSED")
	for _, e := range h.nodes {
		fmt.Println(strings.Join(e.format(), "\n"))
	}

	fmt.Println("UNLICENSED")
	fmt.Println(strings.Join(n.format(), "\n"))
}
