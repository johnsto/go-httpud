package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"bitbucket.org/johnsto/go-httpud/highlight"
	"bitbucket.org/johnsto/go-httpud/highlight/output/term"
)

func PrintHeaders(output *term.Output, resp *http.Response) error {
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
		_, err := output.Emit(t)
		return err
	})
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	// Tokenize body
	err = bodyTokenizer.Tokenize(r, func(t highlight.Token) error {
		_, err := output.Emit(t)
		return err
	})
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	return nil
}

func PrintBody(output *term.Output, resp *http.Response) {
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]

	var err error

	tokenizer, err := highlight.GetTokenizerForContentType(contentType)
	if err != nil {
		fmt.Printf("error getting tokenizer: %s", err)
	}

	if tokenizer == nil {
		// Echo response body to stdout verbatim
		w := bufio.NewWriter(os.Stdout)
		w.ReadFrom(resp.Body)
		return
	}

	err = tokenizer.Tokenize(resp.Body, func(t highlight.Token) error {
		_, err := output.Emit(t)
		return err
	})
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}
}
