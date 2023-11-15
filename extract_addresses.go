package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/godfried/extract-addresses/email"
)

func main() {
	var inputDir string
	addressContext := email.AddressContextForwardedFrom
	flag.StringVar(&inputDir, "input", "", "Email input")
	flag.Var(&addressContext, "context", "Email context")
	flag.Parse()
	files, err := filepath.Glob(filepath.Join(inputDir, "*.eml"))
	if err != nil {
		log.Fatal(err)
	}
	for _, fname := range files {
		err = processFile(fname, addressContext)
		if err != nil {
			log.Fatal(fname, err)
		}
	}
}

func processFile(fname string, addressContext email.AddressContext) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	addresses, err := email.Parse(f)
	if err != nil {
		return err
	}
	for _, address := range addresses[addressContext] {
		fmt.Printf("%s\n", address.Email.Address)
	}
	return nil
}
