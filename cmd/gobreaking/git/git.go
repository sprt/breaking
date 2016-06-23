package git

import (
	"os/exec"
	"strings"
)

type ObjectKind int

const (
	Blob ObjectKind = iota
	Tree
)

type LsTreeEntry struct {
	Filename string
	Kind     ObjectKind
}

func LsTree(treeish string) ([]*LsTreeEntry, error) {
	out, err := exec.Command("git", "ls-tree", treeish).Output()
	if err != nil {
		return nil, err
	}
	var entries []*LsTreeEntry
	for _, line := range strings.Split(string(out)[:len(out)-1], "\n") {
		fields := strings.Fields(line)
		var kind ObjectKind
		if fields[1] == "tree" {
			kind = Tree
		} else {
			kind = Blob
		}
		entries = append(entries, &LsTreeEntry{Filename: fields[3], Kind: kind})
	}

	return entries, nil
}

func Show(treeish string, filename string) ([]byte, error) {
	return exec.Command("git", "show", treeish+":"+filename).Output()
}
