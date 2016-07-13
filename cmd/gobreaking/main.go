// gobreaking reports breaking changes in a Git repository.
//
// Below is the exhaustive list of changes that are considered breaking.
// It applies exclusively to exposed names.
//
//  - Removing a name (constant, type, variable, function).
//  - Changing the kind of a name.
//  - Adding or removing a method in an interface.
//  - Adding or removing a parameter in a function or interface.
//  - Changing the type of a parameter or result in a function or interface.
//  - Adding or removing a result in a function or interface.
//  - Changing the type of an exported struct field.
//  - Removing an exported field from a struct.
//  - Adding an unexported field to a struct containing only exported fields.
//  - Adding an exported field before the last field of a struct
//    containing only exported fields.
//  - Repositioning a field in a struct containing only exported fields.
//
// gobreaking can be invoked two ways:
//
// By providing one argument treeish: it reports the breaking changes between
// treeish and the working directory.
//
// By providing two arguments treeish1 and treeish2: it reports the breaking
// changes between treeish1 and treeish2.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/sprt/breaking"
	"github.com/sprt/breaking/cmd/gobreaking/internal/git"
)

func init() {
	flag.Usage = usage
}

func main() {
	flag.Parse()

	var a, b interface{}
	switch flag.NArg() {
	case 1:
		head, err := treeFiles(flag.Arg(0))
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
	fmt.Fprintf(os.Stderr, "Usage: %s treeish1 [treeish2]\n", os.Args[0])
	flag.PrintDefaults()
}

func treeFiles(treeish string) (map[string]io.Reader, error) {
	tree, err := git.LsTree(treeish)
	if err != nil {
		return nil, err
	}
	return tree.GoFiles()
}
