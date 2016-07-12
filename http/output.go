package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"bitbucket.org/johnsto/go-httpud/highlight"
	"bitbucket.org/johnsto/go-httpud/highlight/output/term"
)

type PrintResponseOptions struct {
	Headers bool
	Body    bool
}

func PrintResponse(output *term.Output, resp *http.Response,
	opts PrintResponseOptions) error {
	httpTokenizer := highlight.GetTokenizer("http")

	contentType := resp.Header.Get("Content-Type")
	bodyTokenizer, err := highlight.GetTokenizerForContentType(contentType)
	if err != nil {
		return err
	}

	r, w := io.Pipe()

	// Write Response to pipe
	go func() {
		err := resp.Write(w)
		if err != nil {
			log.Fatalln(err)
		}
		w.Close()
	}()

	// Tokenize headers
	err = httpTokenizer.Tokenize(r, func(t highlight.Token) error {
		if opts.Headers {
			_, err := output.Emit(t)
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
		if _, err := io.Copy(os.Stdout, r); err != nil {
			return fmt.Errorf("couldn't write to stdout: %s", err)
		}
	} else {
		err = bodyTokenizer.Tokenize(r, func(t highlight.Token) error {
			if opts.Body {
				_, err := output.Emit(t)
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
