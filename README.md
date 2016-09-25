# go-httpud

## Description

HTTPud is a shameless Go clone of the superb [HTTPie](https://github.com/jkbrzt/httpie)
by Jakub Roztočil, but with only a fraction of the features. Intended for 
cases where a Python/pip installation of HTTPie isn't possible.

## Installation

go-httpud's `http` command can be installed with the regular `go get` command:

    go get github.com/johnsto/go-httpud/cmd/http

## Usage

Use the `http` command, e.g.:

    http get yourapihere.com

This will fetch the document, emitting the syntax-highlighted HTTP headers and
body to stdout. 

## Features

Use `http --help` to view a list of options:

    Usage of http:
      --auth="": HTTP basic auth (user[:pass])
      --body=false: only emit response body
      --follow=false: follow HTTP redirects
      --form=false: send parameters as URL-encoded form
      --headers=false: only emit response headers
      --json=false: send parameters as JSON document
      --pretty="all": output style <all|color|format|none>

HTTPud uses the [go-highlight](https://github.com/johnsto/go-highlight) library
for syntax highlighting and formatting, and should work on all platforms. 
Currently only CSS, HTTP, and JSON documents are supported.
