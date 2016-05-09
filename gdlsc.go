package main

import (

	"path"
	//"github.com/ryanuber/go-license"
	"io/ioutil"
	"strings"
	"os"
	"fmt"
)

const vendorDir = "vendor"

type node struct {
	name string
	golang bool
	licenseTxt string
	licenseTree bool
	parent *node
	children []*node
	reduced bool
}

func (n *node) reduce() int {
	reduced := 0
	newChildren := make([]*node, 0, len(n.children))

	for i := 0; i < len(n.children); i++ {
		child := n.children[i]
		if len(child.children) == 0 && n.licenseTree == child.licenseTree {
			n.reduced = true
			reduced += 1
		} else {
			newChildren = append(newChildren, child)
			reduced += child.reduce()
		}

	}

	n.children = newChildren
	return reduced
}

func (n *node) generate(prefix string) []string {
	out := make([]string, 0)
	r := ""
	if n.reduced {
		r = "/..."
	}
	if n.golang || n.licenseTree {
		out = append(out, fmt.Sprintf("%v/%v%v [go:%v, lic:%v]", prefix, n.name, r, n.golang, n.licenseTree))
	}
	if n.licenseTxt != "" {
		lines := strings.Split(n.licenseTxt, "\n")
		for i := 0; i < 5; i++ {
			out = append(out, fmt.Sprintf("| %v", lines[i]))
		}
	}
	for _, child := range n.children {
		out = append(out, child.generate(path.Join(prefix, n.name))...)
	}

	return out
}

func buildNode(prefix string, parent *node, licenseTree bool) *node {
	n := &node{ name: path.Base(prefix), parent: parent, licenseTree: licenseTree, children: make([]*node, 0) }

	files, err := ioutil.ReadDir(prefix)
	if err != nil {
		panic(err);
	}

	for _, file := range files {
		if (file.Mode().IsRegular()) {
			if strings.HasSuffix(file.Name(), ".go") {
				n.golang = true
			}
			if (file.Name() == "LICENSE" || file.Name() == "COPYING" || file.Name() == "LICENSE.txt") {
				license, err := ioutil.ReadFile(path.Join(prefix, file.Name()))
				if err != nil {
					panic(err)
				}
				n.licenseTxt = string(license)
				n.licenseTree = true
			}
		} else if (file.Mode().IsDir() && !strings.HasPrefix(file.Name(), ".")) {
			n.children = append(n.children,
				buildNode(path.Join(prefix, file.Name()), n, licenseTree || n.licenseTree))
		}
	}

	return n
}

func main() {
	n := buildNode(path.Join(os.Args[1], vendorDir), nil, false)
	for (n.reduce() > 0) {}
	fmt.Println(strings.Join(n.generate(""), "\n"))
}
