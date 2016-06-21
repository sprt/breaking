package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sprt/breaking"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("need 2 arguments")
	}

	a := os.Args[1]
	b := os.Args[2]

	diffs, err := breaking.ComparePackages(a, b)
	if err != nil {
		log.Fatalln(err)
	}

	for _, d := range diffs {
		fmt.Println(d.Name())
	}
}
