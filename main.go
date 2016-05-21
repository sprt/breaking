package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sprt/gobreaking/breaking"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("need 2 arguments")
	}

	a := os.Args[1]
	b := os.Args[2]

	report, err := breaking.CompareFiles(a, b)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Deleted:")
	for _, obj := range report.Deleted {
		fmt.Println(obj)
	}
}
