package main

import (
	"fmt"
	"io"
	"os"

	"github.com/sprt/breaking"
	"github.com/sprt/breaking/cmd/gobreaking/git"
)

func main() {
	var a, b interface{}
	switch len(os.Args) {
	case 1:
		head, err := treeFiles("HEAD")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		a, b = head, wd
	case 3:
		x, err := treeFiles(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		y, err := treeFiles(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		a, b = x, y
	default:
		fmt.Fprintln(os.Stderr, "wrong number of arguments")
		os.Exit(2)
	}

	diffs, err := breaking.ComparePackages(a, b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	for _, d := range diffs {
		fmt.Println(d.Name())
	}

	if len(diffs) != 0 {
		os.Exit(1)
	}
}

func treeFiles(treeish string) (map[string]io.Reader, error) {
	tree, err := git.LsTree(treeish)
	if err != nil {
		return nil, err
	}
	return tree.GoFiles()
}
