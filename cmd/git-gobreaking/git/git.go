package git

import (
	"os/exec"
	"path/filepath"
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

func LsTree(recursive bool, treeish string, path string) ([]*LsTreeEntry, error) {
	args := make([]string, 0, 4)
	args = append(args, "ls-tree")
	if recursive {
		args = append(args, "-r")
	}
	args = append(args, treeish)
	args = append(args, path)

	out, err := exec.Command("git", args...).Output()
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

		abs, err := filepath.Abs(fields[3])
		if err != nil {
			return nil, err
		}

		entries = append(entries, &LsTreeEntry{Filename: abs, Kind: kind})
	}

	return entries, nil
}

func Show(treeish string, filename string) ([]byte, error) {
	return exec.Command("git", "show", treeish+":"+filename).Output()
}
