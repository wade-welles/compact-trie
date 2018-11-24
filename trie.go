// Package trie implements a trie for rune slices
package trie

import (
	"fmt"
	"sort"

	"github.com/disiqueira/gotree"
)

type wordArray struct {
	words []string
}

func (w *wordArray) add(word string) {
	w.words = append(w.words, word)
}

// Trie defines a trie with an optional name
type Trie struct {
	Root *Node
	Name string
}

// New creates a trie with name specified. If no name is specified then "Trie" is used
func New(name string) *Trie {
	if name == "" {
		name = "Trie"
	}
	return &Trie{
		Root: &Node{
			children: make(ChildNodeMap),
			isRoot:   true,
		},
		Name: name,
	}
}

// Find check if the trie has the word and return the terminating node of the word
func (t *Trie) Find(word string) (*Node, error) {
	if len(word) == 0 {
		return nil, fmt.Errorf("no string to find")
	}

	runes := []rune(word)

	termNode, err := t.findAtNode(t.Root, runes, 0)
	if err != nil {
		return nil, fmt.Errorf("word %s not found: %s", word, err)
	}

	return termNode, nil
}

// findAtNode gets the node beginning from specified node where the runes terminate
func (t *Trie) findAtNode(n *Node, runes []rune, pos int) (*Node, error) {
	r := runes[pos]
	cNode, ok := n.Children()[r]
	if !ok {
		var prefix string
		if pos > 0 {
			prefix = string(runes[0 : pos-1])
		}
		return nil, fmt.Errorf("string %s not found, longest prefix found: %s", string(runes), prefix)
	}

	if pos < (len(runes) - 1) {
		pos = pos + 1
	} else {
		// This was the last character, check if the node is terminating
		if !cNode.IsTerm() {
			return nil, fmt.Errorf("string %s not found but exists as a non-terminated path", string(runes))
		}
		return cNode, nil
	}

	return t.findAtNode(cNode, runes, pos)
}

// Remove removes the word from the trie. An error is returned is the word is not in the trie
func (t *Trie) Remove(word string) error {
	termNode, err := t.Find(word)
	if err != nil {
		return fmt.Errorf("could not find word %s in trie: %s", word, err)
	}

	termNode.isTerm = false

	curNode := termNode
	for !curNode.isTerm && !curNode.isRoot && !curNode.HasChildren() {
		delete(curNode.parent.children, curNode.value)
		curNode = curNode.parent
	}

	return nil
}

// Add adds a word to the trie and returns the terminating node. If the word already
// exists in the trie an error is returned
func (t *Trie) Add(word string, data interface{}) (*Node, error) {
	if len(word) == 0 {
		return nil, fmt.Errorf("no string to add")
	}

	runes := []rune(word)

	return t.addAtNode(t.Root, runes, data)
}

// addAtNode adds runes starting at node specified and returns the terminating node
func (t *Trie) addAtNode(n *Node, runes []rune, data interface{}) (*Node, error) {
	r := runes[0]
	nResult := n.AddChild(r)

	var cRunes []rune
	if len(runes) > 1 {
		cRunes = runes[1:]
	} else {
		// This was the last character so we should check if this is a terminator
		if nResult.result == nodeFound {
			if nResult.Node.IsTerm() {
				return nil, fmt.Errorf("word already exists in trie")
			}
			nResult.Node.MakeTerm()
			nResult.Node.SetData(data)
		}
		nResult.Node.SetData(data)
		return nResult.Node, nil
	}

	return t.addAtNode(nResult.Node, cRunes, data)
}

// Words returns an array of words in the trie
func (t *Trie) Words() []string {
	words := &wordArray{
		words: []string{},
	}

	t.wordsAtNode(t.Root, "", words)

	return words.words
}

// wordsAtNode returns all words that occur after the node specified
func (t *Trie) wordsAtNode(n *Node, tillThis string, words *wordArray) {
	if n.IsTerm() {
		words.add(tillThis)
	}

	for r, cNode := range n.Children() {
		t.wordsAtNode(cNode, tillThis+string(r), words)
	}
}

// Tree gives a goTree for the trie
func (t *Trie) Tree() gotree.Tree {
	tree := gotree.New(t.Name)

	t.treeAtNode(t.Root, tree)

	return tree
}

// treeAtNode gives the tree beginning from the node specified
func (t *Trie) treeAtNode(n *Node, tree gotree.Tree) {
	// Sort child runes so that the trie viz is consistent
	runes := make(runeSlice, len(n.children))
	i := 0
	for r := range n.children {
		runes[i] = r
		i++
	}
	sort.Sort(runeSlice(runes))

	for _, r := range runes {
		leaf := tree.Add(string(r))
		t.treeAtNode(n.children[r], leaf)
	}
}

// String returns the tree as a string
func (t *Trie) String() string {
	return t.Tree().Print()
}

// Sadly you have to implement a sort interface for a rune :-(
type runeSlice []rune

func (p runeSlice) Len() int           { return len(p) }
func (p runeSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p runeSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
