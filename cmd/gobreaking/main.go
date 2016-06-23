package main

import (
	"fmt"
	"io"
	"log"
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
			log.Fatalln(err)
		}
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		a, b = head, wd
	case 3:
		x, err := treeFiles(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
		y, err := treeFiles(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
		a, b = x, y
	default:
		log.Fatalln("wrong number of arguments")
	}

	diffs, err := breaking.ComparePackages(a, b)
	if err != nil {
		log.Fatalln(err)
	}

	for _, d := range diffs {
		fmt.Println(d.Name())
	}
}

func treeFiles(treeish string) (map[string]io.Reader, error) {
	tree, err := git.LsTree(treeish)
	if err != nil {
		return nil, err
	}
	return tree.GoFiles()
}
