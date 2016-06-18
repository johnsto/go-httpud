package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	m := map[highlight.TokenType]Color{
		highlight.Error:       ColorError,
		highlight.Comment:     ColorComment,
		highlight.Text:        ColorText,
		highlight.Number:      ColorString,
		highlight.String:      ColorString,
		highlight.Attribute:   ColorAttribute,
		highlight.Assignment:  ColorOperator,
		highlight.Operator:    ColorOperator,
		highlight.Punctuation: ColorPunctuation,
		highlight.Constant:    ColorOperator,
		highlight.Entity:      ColorOperator,
	}

	tokens := make(chan highlight.Token)
	go func() {
		for token := range tokens {
			//fmt.Printf("[%s] {%v} %s\n", token.State, token.Type, token.Value)
			c, ok := m[token.Type]
			if !ok {
				c = ColorNormal
			}
			c.Print(token.Value)
		}
	}()

	tokenizer, err := highlight.GetTokenizerForContentType(contentType)
	if err != nil {
		fmt.Printf("error getting tokenizer: %s", err)
	}

	if tokenizer != nil {
		// Run body through tokenizer
		err = tokenizer.Tokenize(resp.Body, tokens)
		if err != io.EOF {
			log.Fatalln(err)
		}
	} else {
		// Echo response body to stdout verbatim
		w := bufio.NewWriter(os.Stdout)
		w.ReadFrom(resp.Body)
	}

}
