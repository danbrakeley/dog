package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/danbrakeley/dog"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	fmt.Printf("Listening on %s...\n", *addr)
	dog.ListenAndServe(*addr)
}
