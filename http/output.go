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
	htmlTokenizer, err := highlight.GetTokenizerForContentType(contentType)
	if err != nil {
		return err
	}

	r, w := io.Pipe()

	tokens := make(chan highlight.Token)
	done := make(chan bool)

	// Output emitter
	go func() {
		for token := range tokens {
			output.Emit(token)
		}
		done <- true
	}()

	// Write Response to pipe
	go func() {
		err := resp.Write(w)
		if err != nil {
			log.Fatalln(err)
		}
		w.Close()
		done <- true
	}()

	// Tokenize headers
	err = httpTokenizer.Tokenize(r, tokens)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	// Tokenize body
	err = htmlTokenizer.Tokenize(r, tokens)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}

	log.Println("THNG")
	close(tokens)

	// Wait for completion
	<-done
	<-done

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

	tokens := make(chan highlight.Token)
	done := make(chan bool)
	go func() {
		for token := range tokens {
			output.Emit(token)
		}
		done <- true
	}()
	err = tokenizer.Tokenize(resp.Body, tokens)
	if err != nil && err != io.EOF {
		log.Fatalln(err)
	}
	close(tokens)
	<-done
}
