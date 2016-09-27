package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/johnsto/go-highlight/output/term"

	"github.com/spf13/pflag"
)

var (
	ErrIgnoringRedirect = errors.New("ignoring redirects")
)

func main() {
	cmd := NewCommand()
	err := cmd.ParseArgs(os.Args[1:])

	if err == pflag.ErrHelp {
		// Print usage and exit
		cmd.Usage()
		return
	} else if err != nil {
		// Parsing args failed
		log.Fatalln(err)
		return
	}

	output := term.NewOutput()

	// Create client
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return ErrIgnoringRedirect
		},
	}

	// Create request
	req, err := cmd.Request()
	if err != nil {
		log.Fatalln(err)
		return
	}
	req.Header.Set("User-Agent", "httpud")

	// Emit request in verbose mode
	if cmd.Verbose {
		fmt.Println("[HTTP Request:]")

		// Write request body to a temporary buffer
		buf := &bytes.Buffer{}
		req.Body = ioutil.NopCloser(io.TeeReader(req.Body, buf))

		// Emit to output
		err = PrintEntity(output, req, req.Header.Get("Content-Type"),
			PrintEntityOptions{
				Headers: cmd.HeadersOnly || !cmd.BodyOnly,
				Body:    !cmd.HeadersOnly || cmd.BodyOnly,
			})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Reset request body
		req.Body = ioutil.NopCloser(buf)

		fmt.Println("\n\n[HTTP Response:]")
	}

	resp, err := client.Do(req)

	switch err := err.(type) {
	case nil:
	case *url.Error:
		if err.Err == ErrIgnoringRedirect {
		} else if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	// Emit response to output
	err = PrintEntity(output, resp, resp.Header.Get("Content-Type"),
		PrintEntityOptions{
			Headers: cmd.HeadersOnly || !cmd.BodyOnly,
			Body:    !cmd.HeadersOnly || cmd.BodyOnly,
		})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
