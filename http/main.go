package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

var (
	ErrIgnoringRedirect = errors.New("ignoring redirects")
)

func main() {
	cmd := NewCommand()
	err := cmd.ParseArgs(os.Args[1:])
	if err != nil {
		log.Fatalln(err)
	}

	req, err := cmd.Request()
	if err != nil {
		log.Fatalln(err)
	}

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

	if cmd.HeadersOnly || !cmd.BodyOnly {
		PrintStatusLine(resp)
		PrintHeaders(resp.Header)
	}
	if !cmd.HeadersOnly && !cmd.BodyOnly {
		fmt.Println()
	}
	if !cmd.HeadersOnly || cmd.BodyOnly {
		PrintBody(resp)
	}
}
