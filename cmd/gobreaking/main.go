package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sprt/breaking"
	"github.com/sprt/breaking/cmd/gobreaking/git"
)

func init() {
	flag.Usage = usage
}

func main() {
	flag.Parse()

	var a, b interface{}
	switch flag.NArg() {
	case 0:
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
	case 2:
		x, err := treeFiles(flag.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		y, err := treeFiles(flag.Arg(1))
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

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [treeish treeish]\n", os.Args[0])
	flag.PrintDefaults()
}

func treeFiles(treeish string) (map[string]io.Reader, error) {
	tree, err := git.LsTree(treeish)
	if err != nil {
		return nil, err
	}
	return tree.GoFiles()
}
