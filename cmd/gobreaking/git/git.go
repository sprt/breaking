// Package git is intended for internal use only.
// Do not import it into your own project as breaking changes may occur.
package git

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
)

type kind int

const (
	blob kind = iota
	tree
)

type treeEntry struct {
	filename string
	kind     kind
}

type Tree struct {
	treeish string
	entries []treeEntry
}

func LsTree(treeish string) (*Tree, error) {
	out, err := exec.Command("git", "ls-tree", treeish).Output()
	if err != nil {
		return nil, err
	}

	var entries []treeEntry
	for _, line := range strings.Split(string(out)[:len(out)-1], "\n") {
		fields := strings.Fields(line)
		var k kind
		if fields[1] == "tree" {
			k = tree
		} else {
			k = blob
		}
		entries = append(entries, treeEntry{filename: fields[3], kind: k})
	}

	return &Tree{treeish, entries}, nil
}

func (t *Tree) GoFiles() (map[string]io.Reader, error) {
	files := make(map[string]io.Reader)
	for _, entry := range t.entries {
		if entry.kind != blob || !strings.HasSuffix(entry.filename, ".go") {
			continue
		}
		b, err := exec.Command("git", "show", t.treeish+":"+entry.filename).Output()
		if err != nil {
			return nil, err
		}
		files[entry.filename] = bytes.NewReader(b)
	}
	return files, nil
}
