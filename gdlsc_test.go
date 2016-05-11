package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var package1 = &node{
	"gdlsc_test",
	"package1",
	false,
	"The MIT License (MIT)",
	[]*node{},
	false,
}
var package2 = &node{
	"gdlsc_test",
	"package2",
	true,
	"",
	[]*node{},
	false,
}
var package3 = &node{
	"gdlsc_test",
	"package3",
	false,
	"Apache license version 2.0, January 2004",
	[]*node{package4},
	false,
}
var package4 = &node{
	"gdlsc_test/package3",
	"package4",
	false,
	"eclipse public license - v 1.0",
	[]*node{},
	false,
}

var topDir = &node{
	".",
	"gdlsc_test",
	false,
	"",
	[]*node{package1, package2, package3},
	false,
}

func TestMakeTree(t *testing.T) {
	assert := assert.New(t)
	n := makeTree(path.Join("gdlsc_test/"))
	if assert.NotNil(n, "Should not be Nil") {
		assert.Equal(n, topDir, " Should be euqal")
		assert.Equal(n.children[0], package1, " Should be euqal")
		assert.Equal(n.children[1], package2, " Should be euqal")
		assert.Equal(n.children[2], package3, " Should be euqal")
	}
}

func TestExtract(t *testing.T) {
	assert := assert.New(t)
	n := makeTree(path.Join("gdlsc_test/"))
	h := &holder{nodes: make([]*node, 0)}
	for n.extract(h) {
	}
	assert.NotNil(h, "Should not be nil")
	assert.NotEqual(len(n.children), 3, "Lenght of n.children should reduce after extracting")
	assert.Equal(n.children[0], package2, "Should Be Equal")
	package1.reduced = true
	assert.Equal(len(h.nodes), 3, "Should be Equal")
	assert.Equal(h.nodes[0], package1, "Should Be Equal")
}

func TestAttachLicenceFile(t *testing.T) {
	n := makeTree(path.Join("gdlsc_test/"))
	h := &holder{nodes: make([]*node, 0)}
	n.extract(h)
	h.attachLicenseType()
	node := h.nodes[0]
	assert.NotEqual(t, node.licenseTxt, "The MIT License (MIT)", "License text should change to license type")
}

func TestReduce(t *testing.T) {
	package6 := &node{
		"gdlsc_test",
		"package6",
		true,
		"",
		[]*node{},
		false,
	}
	package5 := &node{
		"gdlsc_test",
		"package5",
		true,
		"",
		[]*node{package6},
		false,
	}
	n := makeTree(path.Join("gdlsc_test/"))
	n.children = append(n.children, package5)
	for n.reduce() {
	}
	subChild := n.children[3]
	assert.Empty(t, subChild.children, "Should be expty")
	assert.True(t, subChild.reduced, "Should be reduced")
}

func TestGetLicenseType(t *testing.T) {
	licenseType := getLicenseType(package3.licenseTxt)
	assert.NotNil(t, licenseType, "Should not be nil")
	assert.Equal(t, licenseType, "Apache-2.0", "Should get the correct license type")
}
