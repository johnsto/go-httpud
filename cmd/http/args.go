package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

const (
	JSONContentType = "application/json"
	HTMLContentType = "text/html"
	FormContentType = "application/x-www-form-urlencoded"
)

const (
	TokenString  = "([a-zA-Z0-9_\\-\\+]+)"
	QuotedString = "\"?(\\.|[^\"]+)\"?"
	JsonString   = "'?(.+)'?"
)

var (
	headerRegexp = regexp.MustCompile("^" + TokenString + ":" + QuotedString + "$")
	dataRegexp   = regexp.MustCompile("^" + QuotedString + "=" + QuotedString + "$")
	jsonRegexp   = regexp.MustCompile("^" + TokenString + ":=" + JsonString + "$")
	queryRegexp  = regexp.MustCompile("^" + QuotedString + "==" + QuotedString + "$")
)

type UnsupportedContentTypeError struct {
	ContentType string
}

func (e UnsupportedContentTypeError) Error() string {
	return "unsupported content type: " + e.ContentType
}

type Command struct {
	// Request data
	Method    string
	URL       *url.URL
	Headers   http.Header
	Query     url.Values
	Data      map[string]interface{}
	BasicAuth string

	// Request handling
	ParamsAsJSON bool
	ParamsAsForm bool

	// Response handling
	FollowRedirects bool

	// Output options
	HeadersOnly bool
	BodyOnly    bool
	Pretty      string
	Verbose     bool

	// Command line flags
	flags *pflag.FlagSet
}

func NewCommand() *Command {
	c := Command{
		Method:  "GET",
		Headers: http.Header{},
		Data:    map[string]interface{}{},

		flags: pflag.NewFlagSet("http", pflag.ExitOnError),
	}

	fs := c.flags

	fs.BoolVar(&c.ParamsAsJSON, "json", false, "send parameters as JSON document")
	fs.BoolVar(&c.ParamsAsForm, "form", false, "send parameters as URL-encoded form")

	fs.BoolVar(&c.HeadersOnly, "headers", false, "only emit response headers")
	fs.BoolVar(&c.BodyOnly, "body", false, "only emit response body")
	fs.StringVar(&c.Pretty, "pretty", "all", "output style <all|color|format|none>")

	fs.StringVar(&c.BasicAuth, "auth", "", "HTTP basic auth (user[:pass])")
	fs.BoolVarP(&c.FollowRedirects, "follow", "f", false, "follow HTTP redirects")

	fs.BoolVarP(&c.Verbose, "verbose", "v", false, "verbose output")

	fs.Usage = c.Usage

	return &c
}

func (c *Command) ParseArgs(args []string) error {
	// Parse command line flags
	args, err := c.ParseFlags(args)
	if err != nil {
		return err
	}

	// Parse request parameters
	section := "method"
	for _, arg := range args {
		switch section {
		case "method":
			if ok := c.ParseMethod(arg); ok {
				section = "url"
				continue
			}
			fallthrough
		case "url":
			if ok, err := c.ParseURL(arg); err != nil {
				return err
			} else if ok {
				section = "params"
				continue
			}
		case "params":
			if ok, err := c.ParseParam(arg); err != nil {
				return err
			} else if ok {
				continue
			}
		}
	}

	if section != "params" {
		// No URL specified on command line
		return pflag.ErrHelp
	}

	return nil
}

func (c *Command) ParseFlags(args []string) ([]string, error) {
	if err := c.flags.Parse(args); err != nil {
		return nil, err
	}

	if c.ParamsAsJSON {
		c.Headers.Set("Content-Type", JSONContentType)
		c.Headers.Set("Accept", "*/*")
	} else if c.ParamsAsForm {
		c.Headers.Set("Content-Type", FormContentType)
	}

	return c.flags.Args(), nil
}

func (c *Command) ParseMethod(arg string) bool {
	if IsMethodString(arg) {
		c.Method = strings.ToUpper(arg)
		return true
	}
	return false
}

func (c *Command) ParseURL(arg string) (bool, error) {
	if u, err := url.Parse(arg); err != nil {
		return false, fmt.Errorf("invalid url %s: %s", arg, err)
	} else {
		c.URL = u
		if query, err := url.ParseQuery(u.RawQuery); err != nil {
			return false, fmt.Errorf("invalid url %s: %s", arg, err)
		} else {
			c.Query = query
			return true, nil
		}
	}
}

func (c *Command) ParseParam(arg string) (bool, error) {
	if jsonRegexp.MatchString(arg) {
		parts := jsonRegexp.FindStringSubmatch(arg)
		var v interface{}
		if err := json.Unmarshal([]byte(parts[2]), &v); err != nil {
			return false, fmt.Errorf("couldn't parse JSON value '%s': %s",
				parts[2], err)
		}
		c.Data[parts[1]] = v
		return true, nil
	} else if headerRegexp.MatchString(arg) {
		parts := headerRegexp.FindStringSubmatch(arg)
		c.Headers.Add(parts[1], parts[2])
		return true, nil
	} else if queryRegexp.MatchString(arg) {
		parts := queryRegexp.FindStringSubmatch(arg)
		c.Query.Add(parts[1], parts[2])
		return true, nil
	} else if dataRegexp.MatchString(arg) {
		parts := dataRegexp.FindStringSubmatch(arg)
		c.Data[parts[1]] = parts[2]
		return true, nil
	} else {
		return false, fmt.Errorf("unknown argument: %s\n", arg)
	}
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]... [METHOD] url [DATA]...\n", os.Args[0])
	fmt.Fprintf(os.Stderr,
		"Requests url with METHOD, passing DATA in the request.\n")
	fmt.Fprintf(os.Stderr,
		"DATA can take any of the forms listed at the end of this output.\n\n")

	c.flags.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nData types:\n"+
		"  DATA             VALUE                 LOCATION\n"+
		"  Header:Value     Header: Value         Request header\n"+
		"  query==param     ?query=param          Request URL\n"+
		"  key=value        {'key': 'value'}      JSON Request body\n"+
		"  key=123          {'key': '123'}        JSON Request body\n"+
		"  key:=123         {'key': 123}          JSON Request body\n"+
		"  key=value        key=value             Request body (with --form)\n")
}

// Request returns an `http.Request` representing the arguments entered by the
// user.
func (c *Command) Request() (*http.Request, error) {
	if c.URL.Scheme == "" {
		c.URL.Scheme = "http"
	}

	// Update query part of URI
	c.URL.RawQuery = c.Query.Encode()

	// Encode body
	contentType := c.Headers.Get("Content-Type")
	if contentType == "" {
		contentType = JSONContentType
	}

	var body io.Reader
	if len(c.Data) > 0 {
		c.Headers.Set("Content-Type", contentType)

		var err error
		body, err = MakeBody(contentType, c.Data)

		if err != nil {
			return nil, err
		}
	}

	// Create request
	req, err := http.NewRequest(c.Method, c.URL.String(), body)
	if err != nil {
		return nil, err
	}

	// Configure Basic Authentication parameters
	if c.BasicAuth != "" {
		parts := strings.SplitN(c.BasicAuth, ":", 2)
		req.SetBasicAuth(parts[1], parts[2])
	}

	// Set headers
	req.Header.Set("Host", c.URL.Host)
	for k := range c.Headers {
		req.Header.Set(k, c.Headers.Get(k))
	}

	return req, nil
}

// MakeBody converts the provided data to the specified type, returning a
// Reader emitting the content. If the specified content type is not
// supported, an UnsupportedContentTypeError will be returned.
func MakeBody(contentType string, data map[string]interface{}) (
	io.Reader, error) {
	switch contentType {
	case "application/json":
		return MakeJSONBody(data)
	case "application/x-www-form-urlencoded":
		return MakeFormBody(data)
	default:
		return nil, UnsupportedContentTypeError{contentType}
	}
}

// MakeJSONBody converts the provided key/values to JSON data.
func MakeJSONBody(data map[string]interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal data: %s", err)
	}
	return bytes.NewReader(b), nil
}

// MakeFormBody converts the provided key/values to HTTP form data.
func MakeFormBody(data map[string]interface{}) (io.Reader, error) {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, fmt.Sprint(v))
	}
	return strings.NewReader(form.Encode()), nil
}

// IsMethodString returns true if `s` is a valid and supported HTTP method.
func IsMethodString(s string) bool {
	_, ok := map[string]bool{
		"get":     true,
		"head":    true,
		"post":    true,
		"put":     true,
		"patch":   true,
		"delete":  true,
		"connect": true,
		"options": true,
		"trace":   true,
	}[strings.ToLower(s)]
	return ok
}
