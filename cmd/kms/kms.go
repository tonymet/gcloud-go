package main

import (
	"flag"
	"fmt"
	"github.com/tonymet/gcloud-go/kms"
	"os"
)

var (
	filename string
	output   string
	keypath  string
)

func init() {
	flag.StringVar(&filename, "f", "", "path \"message\" file")
	flag.StringVar(&output, "o", "", "output path for sig")
	flag.StringVar(&keypath, "k", "", "full path to key ID (including version)")
}

func main() {
	flag.Parse()
	// setup in & out
	outputWriter, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	defer outputWriter.Close()
	inputReader, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer inputReader.Close()
	err = kms.SignAsymmetric(outputWriter, keypath, inputReader)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "Signature written to: %s\n", outputWriter.Name())
}
