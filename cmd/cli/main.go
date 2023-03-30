package main

import (
	"log"
	"os"

	converter "github.com/kislerdm/usd2eur"
)

func main() {
	conv, err := converter.NewConverter(nil)
	if err != nil {
		log.Fatalln(err)
	}

	cli := NewCLI(os.Stdin, os.Stdout, conv)

	if err := cli.Run(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}
