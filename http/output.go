package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"bitbucket.org/johnsto/go-httpud/highlight"
)

func PrintResponse(resp *http.Response) {
	PrintStatusLine(resp)
	PrintHeaders(resp.Header)
}

func PrintRegexp(s, re string, cs []Color) {
	r := regexp.MustCompile(re)
	indices := r.FindStringSubmatchIndex(s)

	// print prefix
	if indices[2] > indices[0] {
		ColorNormal.Printf(s[indices[0]:indices[2]])
	}

	for i := 2; i < len(indices); i += 2 {
		start, end := indices[i], indices[i+1]
		sub := s[start:end]
		c := cs[(i-2)/2]
		c.Printf(sub)
	}

	// print suffix
	if len(s) > indices[1] {
		ColorNormal.Printf(s[indices[1]:len(s)])
	}
}

func PrintStatusLine(resp *http.Response) {
	PrintRegexp(
		fmt.Sprintf("HTTP/%d.%d %s\n", resp.ProtoMajor, resp.ProtoMinor,
			resp.Status),
		"(HTTP)(/)(\\d+\\.\\d+)( +)(\\d{3})( +)(.+)",
		[]Color{ColorReserved, ColorOperator, ColorNumber, nil,
			ColorNumber, nil, ColorStatus},
	)
}

func PrintHeaders(h http.Header) {
	var ks []string
	for k := range h {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for _, k := range ks {
		v := h.Get(k)
		PrintRegexp(
			fmt.Sprintf("%s: %s\n", k, v),
			"(.*?)( *)(:)( *)(.+)",
			[]Color{ColorAttribute, nil, ColorOperator, nil, ColorText},
		)
	}
}

func PrintBody(resp *http.Response) {
	contentType := strings.Split(resp.Header.Get("Content-Type"), ";")[0]

	var err error

	tokens := make(chan highlight.Token)
	go func() {
		for token := range tokens {
			c := string(token.Value)
			switch token.Type {
			case highlight.Separator:
				ColorOperator.Print(c)
			case highlight.Operator:
				ColorOperator.Print(c)
			case highlight.Attribute:
				ColorAttribute.Print(c)
			case highlight.Entity:
				ColorOperator.Print(c)
			case highlight.Punctuation:
				ColorPunctuation.Print(c)
			case highlight.Comment:
				ColorComment.Print(c)
			case highlight.String:
				ColorString.Print(c)
			default:
				ColorText.Print(c)
			}
		}
	}()

	switch contentType {
	case HTMLContentType:
		err = highlight.HTML.Tokenise(resp.Body, tokens)
		if err != io.EOF {
			log.Fatalln(err)
		}
		/*case JSONContentType:
		err = highlight.JSON.Tokenise(resp.Body, tokens)
		if err != io.EOF {
			log.Fatalln(err)
		}*/
	}

}
