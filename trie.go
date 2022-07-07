/*
MIT License

Copyright (c) 2020 Loyalsoldier

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"errors"
	"strings"
)

type node struct {
	leaf     bool
	children map[string]*node
}

func newNode() *node {
	return &node{
		leaf:     false,
		children: make(map[string]*node),
	}
}

func (n *node) getChild(s string) *node {
	return n.children[s]
}

func (n *node) hasChild(s string) bool {
	return n.getChild(s) != nil
}

func (n *node) addChild(s string, child *node) {
	n.children[s] = child
}

func (n *node) isLeaf() bool {
	return n.leaf
}

type DomainTrie struct {
	root *node
}

func NewDomainTrie() *DomainTrie {
	return &DomainTrie{
		root: newNode(),
	}
}

func (t *DomainTrie) Insert(domain string) (bool, error) {
	if domain == "" {
		return false, errors.New("empty domain")
	}
	parts := strings.Split(domain, ".")

	node := t.root
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]

		if node.isLeaf() {
			return false, nil
		}
		if !node.hasChild(part) {
			node.addChild(part, newNode())
			if i == 0 {
				node.getChild(part).leaf = true
				return true, nil
			}
		}
		node = node.getChild(part)
	}
	return false, nil
}
