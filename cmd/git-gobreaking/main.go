package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sprt/breaking"
	"github.com/sprt/breaking/cmd/git-gobreaking/git"
)

func main() {
	head, err := getHeadFiles()
	if err != nil {
		log.Fatalln(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	diffs, err := breaking.ComparePackages(head, wd)
	if err != nil {
		log.Fatalln(err)
	}

	for _, d := range diffs {
		fmt.Println(d.Name())
	}
}

func getHeadFiles() (map[string]io.Reader, error) {
	tree, err := git.LsTree(false, "HEAD", ".")
	if err != nil {
		return nil, err
	}

	readers := make(map[string]io.Reader)
	for _, entry := range tree {
		if entry.Kind != git.Blob || !strings.HasSuffix(entry.Filename, ".go") {
			continue
		}
		b, err := git.Show("HEAD", filepath.Base(entry.Filename))
		if err != nil {
			return nil, err
		}
		readers[entry.Filename] = bytes.NewReader(b)
	}

	return readers, nil
}
