package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/johnsto/go-highlight"
	_ "github.com/johnsto/go-highlight/lexers"
	"github.com/johnsto/go-highlight/output/term"
)

type PrintEntityOptions struct {
	Headers bool
	Body    bool
}

type Writable interface {
	Write(w io.Writer) error
}

func PrintEntity(output *term.Output, writable Writable, contentType string,
	opts PrintEntityOptions) error {
	httpTokenizer := highlight.GetTokenizer("http")

	bodyTokenizer, err := highlight.GetTokenizerForContentType(contentType)
	if contentType != "" && err != nil {
		return err
	}

	r, w := io.Pipe()

	// Write to pipe
	go func() {
		err := writable.Write(w)
		if err != nil {
			log.Fatalln(err)
		}
		w.Close()
	}()

	br := bufio.NewReader(r)

	// Tokenize headers
	err = httpTokenizer.Tokenize(br, func(t highlight.Token) error {
		if opts.Headers {
			err := output.Emit(t)
			return err
		}
		return nil
	})

	if err != nil && err != io.EOF {
		return err
	}

	// Tokenize body
	if bodyTokenizer == nil {
		// No tokenizer; emit straight to output
		if _, err := io.Copy(os.Stdout, br); err != nil {
			return fmt.Errorf("couldn't write to stdout: %s", err)
		}
	} else {
		err = bodyTokenizer.Tokenize(br, func(t highlight.Token) error {
			if opts.Body {
				err := output.Emit(t)
				return err
			}
			return nil
		})
		if err != nil && err != io.EOF {
			return fmt.Errorf("couldn't tokenise to stdout: %s", err)
		}
	}

	return nil
}
