package main

import (
	"path"
	"io/ioutil"
	"strings"
	"os"
	"fmt"
)

const vendorDir = "vendor"

// node describes a directory in the vendor tree
type node struct {
	prefix string
	name string
	golang bool
	licenseTxt string
	children []*node
	reduced bool
}

// holder allows to append while keeping a reference to the slice
type holder struct {
	nodes []*node
}

// makeTree parses the vendor tree
func makeTree(p string) *node {
	n := &node{ prefix: path.Dir(p), name: path.Base(p), children: make([]*node, 0) }

	files, err := ioutil.ReadDir(p)
	if err != nil {
		panic(err);
	}

	for _, file := range files {
		if (file.Mode().IsRegular()) {
			if strings.HasSuffix(file.Name(), ".go") {
				n.golang = true
			} else if (isLicenseFileName(file.Name())) {
				license, err := ioutil.ReadFile(path.Join(p, file.Name()))
				if err != nil {
					panic(err)
				}
				n.licenseTxt = string(license)
			}
		} else if (file.Mode().IsDir() && !strings.HasPrefix(file.Name(), ".") && file.Name() != "ConnectCorp") {
			n.children = append(n.children,
				makeTree(path.Join(p, file.Name())))
		}
	}

	return n
}

// extract takes out some of the minimal licensed subtrees
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

	out := make([]string, 0)
	if n.golang || n.licenseTxt != "" {
		out = append(out, fmt.Sprintf("%v/%v%v [go:%v,lic:%v]", n.prefix, n.name, r, n.golang, n.licenseTxt != ""))
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

	for (n.extract(h)) {}
	for (n.reduce()) {}

	fmt.Println("LICENSED")
	for _, e := range h.nodes {
		fmt.Println(strings.Join(e.format(), "\n"))
	}

	fmt.Println("UNLICENSED")
	fmt.Println(strings.Join(n.format(), "\n"))
}
