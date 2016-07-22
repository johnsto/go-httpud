package main

import (
	"errors"
	"fmt"
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
		cmd.Usage()
		return
	}

	if err != nil {
		log.Fatalln(err)
	}

	req, err := cmd.Request()
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", "httpud")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return ErrIgnoringRedirect
		},
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

	output := term.NewOutput()

	err = PrintResponse(output, resp, PrintResponseOptions{
		Headers: cmd.HeadersOnly || !cmd.BodyOnly,
		Body:    !cmd.HeadersOnly || cmd.BodyOnly,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
