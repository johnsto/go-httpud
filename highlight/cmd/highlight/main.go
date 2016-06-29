package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"bitbucket.org/johnsto/go-httpud/highlight"
	"bitbucket.org/johnsto/go-httpud/highlight/output/term"
	"github.com/spf13/pflag"
)

func main() {
	pflag.Parse()
	filename := pflag.Arg(0)

	if filename == "" {
		fmt.Println("No file specified")
		return
	}

	tokenizer, err := highlight.GetTokenizerForFilename(path.Base(filename))
	if err != nil {
		fmt.Println("couldn't get tokenizer for file type:", err)
		return
	} else if tokenizer == nil {
		fmt.Println("couldn't find tokenizer for file type")
		return
	}

	output := term.NewOutput()

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}

	tokens := make(chan highlight.Token)
	go func() {
		for token := range tokens {
			output.Emit(token)
		}
	}()
	err = tokenizer.Tokenize(f, tokens)
	close(tokens)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}
}
