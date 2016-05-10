package main

import (
	"path"
	"testing"
)

var package1 = &Node{
	"gdlsc_test",
	"package1",
	false,
	"The MIT License (MIT)",
	[]*Node{},
	false,
}
var package2 = &Node{
	"gdlsc_test",
	"package2",
	true,
	"",
	[]*Node{},
	false,
}

var topDir = &Node{
	".",
	"gdlsc_test",
	false,
	"",
	[]*Node{package1, package2},
	false,
}

func (n *Node) equal(o *Node) bool {
	if (n.name != o.name) || (n.prefix != o.prefix) || (n.golang != o.golang) || (n.licenseTxt != o.licenseTxt) || (n.reduced != o.reduced) {
		return false
	}
	if len(n.children) != len(o.children) {
		return false
	}
	return true
}

func TestMakeTree(t *testing.T) {
	n := MakeTree(path.Join("gdlsc_test/"))
	if n == nil {
		t.Fail()
	}
	if !n.equal(topDir) {
		t.Fail()
	}
	child1 := n.children[0]
	if !child1.equal(package1) {
		t.Fail()
	}
	child2 := n.children[1]
	if !child2.equal(package2) {
		t.Fail()
	}
}

func TestExtract(t *testing.T) {
	n := MakeTree(path.Join("gdlsc_test/"))
	h := &Holder{nodes: make([]*Node, 0)}
	if !n.Extract(h) {
		t.Fail()
	}
	if h == nil {
		t.Fail()
	}
	if len(n.children) == 2 {
		t.Fail()
	}
	child1 := n.children[0]
	if !child1.equal(package2) {
		t.Fail()
	}
	node := h.nodes[0]
	if package1.reduced = true; !node.equal(package1) {
		t.Fail()
	}
}

func TestAttachLicenceFile(t *testing.T) {
	n := MakeTree(path.Join("gdlsc_test/"))
	h := &Holder{nodes: make([]*Node, 0)}
	n.Extract(h)
	h.AttachLicenseType()
	if node := h.nodes[0]; node.licenseTxt == "The MIT License (MIT)" {
		t.Fail()
	}
}
